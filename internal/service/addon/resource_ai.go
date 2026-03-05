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

type aiResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &aiResource{}
	_ resource.ResourceWithConfigure   = &aiResource{}
	_ resource.ResourceWithImportState = &aiResource{}
)

func NewAIResource() resource.Resource {
	return &aiResource{}
}

func (r *aiResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_ai"
}

func (r *aiResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureAddonClient(req, resp)
}

type aiResourceModel struct {
	aiBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *aiResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              common.SchemaResourceId("Addon AI"),
			"location":        schemaResourceAddonLocation("Addon AI"),
			"sku":             schemaResourceAddonSKU("Addon AI", []int32{1}),
			"deployment_name": schemaResourceAddonDeploymentName("Addon AI"),
			"url":             schemaResourceAddonURL("Addon AI"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Addon AI.",
	}
}

func (r *aiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *aiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aiResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	op := addon.NewAIOp(r.client)
	result, err := op.Create(ctx, plan.Location.ValueString(), v1.AiServiceSku(plan.Sku.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Addon AI: %s", err))
		return
	}

	id, deploymentName, ok := getAddonIDsFromDeployment("AI", result, &resp.Diagnostics)
	if !ok {
		return
	}

	ai, err := waitDeployment(ctx, "AI", op.Read, id)
	if err != nil {
		resp.Diagnostics.AddError("Create: Resource Error", fmt.Sprintf("failed to wait for Addon AI[%s] deployment: %s", id, err))
		return
	}

	plan.updateState(id, deploymentName, ai.URL.Value, &v1.AiRequestBody{
		Location: plan.Location.ValueString(),
		Sku:      v1.AiServiceSku(plan.Sku.ValueInt32()),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state aiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := addon.NewAIOp(r.client)
	result := getAddon(ctx, "AI", state.ID.ValueString(), op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeAIResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon AI[%s] response: %s", state.ID.ValueString(), err))
		return
	}

	state.updateState(state.ID.ValueString(), state.DeploymentName.ValueString(), result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Addon AI does not support updates")
}

func (r *aiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state aiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	if err := addon.NewAIOp(r.client).Delete(ctx, state.ID.ValueString()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Addon AI[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func decodeAIResponse(resp *v1.GetResourceResponse) (v1.AiRequestBody, error) {
	var result v1.AiRequestBody
	if resp == nil || len(resp.Data) == 0 {
		return result, errors.New("got invalid response from Addon AI API")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return result, err
	}

	result.Location = data["location"].(map[string]interface{})["name"].(string)
	result.Sku = v1.AiServiceSku(1) // TODO: get actual SKU from response when Addon API supports multiple SKUs
	return result, nil
}
