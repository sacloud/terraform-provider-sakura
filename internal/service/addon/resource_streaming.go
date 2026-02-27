// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/sacloud/addon-api-go"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type streamingResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &streamingResource{}
	_ resource.ResourceWithConfigure   = &streamingResource{}
	_ resource.ResourceWithImportState = &streamingResource{}
)

func NewStreamingResource() resource.Resource {
	return &streamingResource{}
}

func (r *streamingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_streaming"
}

func (r *streamingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureAddonClient(req, resp)
}

type streamingResourceModel struct {
	streamingBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *streamingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       common.SchemaResourceId("Addon Streaming"),
			"location": schemaResourceAddonLocation("Addon Streaming"),
			"unit_count": schema.StringAttribute{
				Required:    true,
				Description: "The unit count of the Addon Streaming.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployment_name": schemaResourceAddonDeploymentName("Addon Streaming"),
			"url":             schemaResourceAddonURL("Addon Streaming"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Addon Streaming.",
	}
}

func (r *streamingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *streamingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan streamingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	op := addon.NewStreamingOp(r.client)
	result, err := op.Create(ctx, plan.Location.ValueString(), plan.UnitCount.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Addon Streaming: %s", err))
		return
	}

	id, deploymentName, ok := getAddonIDsFromDeployment("Streaming", result, &resp.Diagnostics)
	if !ok {
		return
	}

	streaming, err := waitDeployment(ctx, "Streaming", op.Read, id)
	if err != nil {
		resp.Diagnostics.AddError("Create: Resource Error", fmt.Sprintf("failed to wait for Addon Streaming[%s] deployment: %s", id, err))
		return
	}

	plan.updateState(id, deploymentName, streaming.URL.Value, &v1.StreamingRequestBody{
		Location:  plan.Location.ValueString(),
		UnitCount: plan.UnitCount.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *streamingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state streamingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := addon.NewStreamingOp(r.client)
	result := getAddon(ctx, "Streaming", state.ID.ValueString(), op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeStreamingResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon Streaming[%s] response: %s", state.ID.ValueString(), err))
		return
	}
	state.updateState(state.ID.ValueString(), state.DeploymentName.ValueString(), result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *streamingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Addon Streaming does not support updates")
}

func (r *streamingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state streamingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	op := addon.NewStreamingOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Addon Streaming[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func decodeStreamingResponse(resp *v1.GetResourceResponse) (v1.StreamingRequestBody, error) {
	var result v1.StreamingRequestBody
	if resp == nil || len(resp.Data) == 0 {
		return result, errors.New("got invalid response from Addon Streaming API")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return result, err
	}

	result.Location = loweredLocation(data["location"].(string))
	result.UnitCount = strconv.Itoa(int(data["sku"].(map[string]any)["capacity"].(float64)))
	return result, nil
}
