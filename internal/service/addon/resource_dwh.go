// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/sacloud/addon-api-go"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type dwhResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &dwhResource{}
	_ resource.ResourceWithConfigure   = &dwhResource{}
	_ resource.ResourceWithImportState = &dwhResource{}
)

func NewDWHResource() resource.Resource {
	return &dwhResource{}
}

func (r *dwhResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_dwh"
}

func (r *dwhResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureAddonClient(req, resp)
}

type dwhResourceModel struct {
	dwhBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *dwhResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              common.SchemaResourceId("Addon DWH"),
			"location":        schemaResourceAddonLocation("Addon DWH"),
			"deployment_name": schemaResourceAddonDeploymentName("Addon DWH"),
			"url":             schemaResourceAddonURL("Addon DWH"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Addon DWH.",
	}
}

func (r *dwhResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dwhResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dwhResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	op := addon.NewDWHOp(r.client)
	result, err := op.Create(ctx, plan.Location.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Addon DWH: %s", err))
		return
	}

	id, deploymentName, ok := getAddonIDsFromDeployment("DWH", result, &resp.Diagnostics)
	if !ok {
		return
	}

	dwh, err := waitDeployment(ctx, "DWH", op.Read, id)
	if err != nil {
		resp.Diagnostics.AddError("Create: Resource Error", fmt.Sprintf("failed to wait for Addon DWH[%s] deployment: %s", id, err))
		return
	}

	plan.updateState(id, deploymentName, dwh.URL.Value, &v1.DatawarehousePostRequestBody{
		Location: plan.Location.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dwhResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dwhResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := addon.NewDWHOp(r.client)
	result := getAddon(ctx, "DWH", state.ID.ValueString(), op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeDWHResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon DWH[%s] response: %s", state.ID.ValueString(), err))
		return
	}
	state.updateState(state.ID.ValueString(), state.DeploymentName.ValueString(), result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dwhResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Addon DWH does not support updates")
}

func (r *dwhResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dwhResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	op := addon.NewDWHOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Addon DWH[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func decodeDWHResponse(resp *v1.GetResourceResponse) (v1.DatawarehousePostRequestBody, error) {
	var result v1.DatawarehousePostRequestBody
	if resp == nil || len(resp.Data) == 0 {
		return result, errors.New("got invalid response from Addon DWH API")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return result, err
	}

	result.Location = data["location"].(string)
	return result, nil
}
