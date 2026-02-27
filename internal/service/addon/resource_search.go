// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/addon-api-go"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type searchResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &searchResource{}
	_ resource.ResourceWithConfigure   = &searchResource{}
	_ resource.ResourceWithImportState = &searchResource{}
)

func NewSearchResource() resource.Resource {
	return &searchResource{}
}

func (r *searchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_search"
}

func (r *searchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureAddonClient(req, resp)
}

type searchResourceModel struct {
	searchBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *searchResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       common.SchemaResourceId("Addon Search"),
			"location": schemaResourceAddonLocation("Addon Search"),
			"partition_count": schema.Int32Attribute{
				Required:    true,
				Description: "The partition count of the Addon Search.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"replica_count": schema.Int32Attribute{
				Required:    true,
				Description: "The replica count of the Addon Search.",
				Validators: []validator.Int32{
					int32validator.Between(1, 12),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"sku":             schemaResourceAddonSKU("Addon Search", common.MapTo(v1.SearchSku1.AllValues(), common.ToInt32)),
			"deployment_name": schemaResourceAddonDeploymentName("Addon Search"),
			"url":             schemaResourceAddonURL("Addon Search"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Addon Search.",
	}
}

func (r *searchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *searchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan searchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	op := addon.NewSearchOp(r.client)
	result, err := op.Create(ctx, addon.SearchCreateParams{
		Location:       plan.Location.ValueString(),
		PartitionCount: plan.PartitionCount.ValueInt32(),
		ReplicaCount:   plan.ReplicaCount.ValueInt32(),
		Sku:            v1.SearchSku(plan.Sku.ValueInt32()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Addon Search: %s", err))
		return
	}

	id, deploymentName, ok := getAddonIDsFromDeployment("Search", result, &resp.Diagnostics)
	if !ok {
		return
	}

	search, err := waitDeployment(ctx, "Search", op.Read, id)
	if err != nil {
		resp.Diagnostics.AddError("Create: Resource Error", fmt.Sprintf("failed to wait for Addon Search[%s] deployment: %s", id, err))
		return
	}

	plan.updateState(id, deploymentName, search.URL.Value, &v1.SearchPostRequestBody{
		Location:       plan.Location.ValueString(),
		PartitionCount: plan.PartitionCount.ValueInt32(),
		ReplicaCount:   plan.ReplicaCount.ValueInt32(),
		Sku:            v1.SearchSku(plan.Sku.ValueInt32()),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *searchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state searchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := addon.NewSearchOp(r.client)
	result := getAddon(ctx, "Search", state.ID.ValueString(), op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeSearchResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon Search[%s] response: %s", state.ID.ValueString(), err))
		return
	}
	state.updateState(state.ID.ValueString(), state.DeploymentName.ValueString(), result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *searchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Addon Search does not support updates")
}

func (r *searchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state searchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	op := addon.NewSearchOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Addon Search[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

var searchSkuMap = map[string]v1.SearchSku{
	"free":                 v1.SearchSku1,
	"basic":                v1.SearchSku2,
	"standard1":            v1.SearchSku3,
	"standard2":            v1.SearchSku4,
	"standard3":            v1.SearchSku5,
	"storage_optimized_l1": v1.SearchSku7,
	"storage_optimized_l2": v1.SearchSku8,
}

func decodeSearchResponse(resp *v1.GetResourceResponse) (v1.SearchPostRequestBody, error) {
	var result v1.SearchPostRequestBody
	if resp == nil || len(resp.Data) == 0 {
		return result, errors.New("got invalid response from Addon Search API")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return result, err
	}

	props := data["properties"].(map[string]interface{})
	result.Location = loweredLocation(data["location"].(string))
	result.PartitionCount = int32(props["partitionCount"].(float64))
	result.ReplicaCount = int32(props["replicaCount"].(float64))
	result.Sku = searchSkuMap[data["sku"].(map[string]any)["name"].(string)]
	if props["hostingMode"].(string) == "highDensity" {
		result.Sku = v1.SearchSku6
	}
	return result, nil
}
