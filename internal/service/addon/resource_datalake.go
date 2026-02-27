// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

type dataLakeResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &dataLakeResource{}
	_ resource.ResourceWithConfigure   = &dataLakeResource{}
	_ resource.ResourceWithImportState = &dataLakeResource{}
)

func NewDataLakeResource() resource.Resource {
	return &dataLakeResource{}
}

func (r *dataLakeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_datalake"
}

func (r *dataLakeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureAddonClient(req, resp)
}

type dataLakeResourceModel struct {
	dataLakeBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *dataLakeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       common.SchemaResourceId("Addon DataLake"),
			"location": schemaResourceAddonLocation("Addon DataLake"),
			"performance": schema.Int32Attribute{
				Required:    true,
				Description: "The performance setting of the Addon DataLake.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int32{
					int32validator.OneOf(common.MapTo(v1.DataLakePerformance1.AllValues(), common.ToInt32)...),
				},
			},
			"redundancy": schema.Int32Attribute{
				Required:    true,
				Description: "The redundancy setting of the Addon DataLake.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int32{
					int32validator.OneOf(common.MapTo(v1.DataLakeRedundancy1.AllValues(), common.ToInt32)...),
				},
			},
			"deployment_name": schemaResourceAddonDeploymentName("Addon DataLake"),
			"url":             schemaResourceAddonURL("Addon DataLake"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Addon DataLake.",
	}
}

func (r *dataLakeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dataLakeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dataLakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	op := addon.NewDataLakeOp(r.client)
	result, err := op.Create(ctx, addon.DataLakeCreateParams{
		Location:    plan.Location.ValueString(),
		Performance: v1.DataLakePerformance(plan.Performance.ValueInt32()),
		Redundancy:  v1.DataLakeRedundancy(plan.Redundancy.ValueInt32()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Addon DataLake: %s", err))
		return
	}

	id, deploymentName, ok := getAddonIDsFromDeployment("DataLake", result, &resp.Diagnostics)
	if !ok {
		return
	}

	datalake, err := waitDeployment(ctx, "DataLake", op.Read, id)
	if err != nil {
		resp.Diagnostics.AddError("Create: Resource Error", fmt.Sprintf("failed to wait for Addon DataLake[%s] deployment: %s", id, err))
		return
	}

	plan.updateState(id, deploymentName, datalake.URL.Value, &v1.DatalakePostRequestBody{
		Location:    plan.Location.ValueString(),
		Performance: v1.DataLakePerformance(plan.Performance.ValueInt32()),
		Redundancy:  v1.DataLakeRedundancy(plan.Redundancy.ValueInt32()),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataLakeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataLakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := addon.NewDataLakeOp(r.client)
	result := getAddon(ctx, "DataLake", state.ID.ValueString(), op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeDataLakeResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon DataLake[%s] response: %s", state.ID.ValueString(), err))
		return
	}
	state.updateState(state.ID.ValueString(), state.DeploymentName.ValueString(), result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dataLakeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Addon DataLake does not support updates")
}

func (r *dataLakeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dataLakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	op := addon.NewDataLakeOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Addon DataLake[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

var dataLakePerformanceMap = map[string]v1.DataLakePerformance{
	"Standard": v1.DataLakePerformance1,
	"Premium":  v1.DataLakePerformance2,
}

var dataLakeRedundancyMap = map[string]v1.DataLakeRedundancy{
	"LRS":  v1.DataLakeRedundancy1,
	"GRS":  v1.DataLakeRedundancy2,
	"ZRS":  v1.DataLakeRedundancy3,
	"GZRS": v1.DataLakeRedundancy4,
}

func decodeDataLakeResponse(resp *v1.GetResourceResponse) (v1.DatalakePostRequestBody, error) {
	var result v1.DatalakePostRequestBody
	if resp == nil || len(resp.Data) == 0 {
		return result, errors.New("got invalid response from Addon DataLake API")
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return result, err
	}

	// The SKU name is in the format of "{Performance}_{Redundancy}", e.g. "Standard_LRS"
	parts := strings.SplitN(data["sku"].(map[string]any)["name"].(string), "_", 2)
	result.Location = data["location"].(string)
	result.Performance = dataLakePerformanceMap[parts[0]]
	result.Redundancy = dataLakeRedundancyMap[parts[1]]
	return result, nil
}
