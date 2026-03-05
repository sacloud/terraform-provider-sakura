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

type etlResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &etlResource{}
	_ resource.ResourceWithConfigure   = &etlResource{}
	_ resource.ResourceWithImportState = &etlResource{}
)

func NewETLResource() resource.Resource {
	return &etlResource{}
}

func (r *etlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_etl"
}

func (r *etlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureAddonClient(req, resp)
}

type etlResourceModel struct {
	etlBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *etlResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              common.SchemaResourceId("Addon ETL"),
			"location":        schemaResourceAddonLocation("Addon ETL"),
			"deployment_name": schemaResourceAddonDeploymentName("Addon ETL"),
			"url":             schemaResourceAddonURL("Addon ETL"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Addon ETL.",
	}
}

func (r *etlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *etlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan etlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	op := addon.NewETLOp(r.client)
	result, err := op.Create(ctx, plan.Location.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Addon ETL: %s", err))
		return
	}

	resourceGroupName, deploymentName, ok := getAddonIDsFromDeployment("ETL", result, &resp.Diagnostics)
	if !ok {
		return
	}

	etl, err := waitDeployment(ctx, "ETL", op.Read, resourceGroupName)
	if err != nil {
		resp.Diagnostics.AddError("Create: Resource Error", fmt.Sprintf("failed to wait for Addon ETL[%s] deployment: %s", resourceGroupName, err))
		return
	}

	plan.updateState(resourceGroupName, deploymentName, etl.URL.Value, &v1.EtlPostRequestBody{
		Location: plan.Location.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *etlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state etlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := addon.NewETLOp(r.client)
	result := getAddon(ctx, "ETL", state.ID.ValueString(), op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeETLResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon ETL[%s] response: %s", state.ID.ValueString(), err))
		return
	}
	state.updateState(state.ID.ValueString(), state.DeploymentName.ValueString(), result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *etlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Addon ETL does not support updates")
}

func (r *etlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state etlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	op := addon.NewETLOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Addon ETL[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func decodeETLResponse(resp *v1.GetResourceResponse) (v1.EtlPostRequestBody, error) {
	var result v1.EtlPostRequestBody
	if resp == nil || len(resp.Data) == 0 {
		return result, errors.New("got invalid response from Addon ETL API")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return result, err
	}

	result.Location = data["location"].(string)
	return result, nil
}
