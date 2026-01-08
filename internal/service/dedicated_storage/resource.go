// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dedicated_storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	dedicatedstorage "github.com/sacloud/dedicated-storage-api-go"
	v1 "github.com/sacloud/dedicated-storage-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type dedicatedStorageResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &dedicatedStorageResource{}
	_ resource.ResourceWithConfigure   = &dedicatedStorageResource{}
	_ resource.ResourceWithImportState = &dedicatedStorageResource{}
)

func NewDedicatedStorageResource() resource.Resource {
	return &dedicatedStorageResource{}
}

func (r *dedicatedStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dedicated_storage"
}

func (r *dedicatedStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	r.client = apiClient.DedicatedStorageClient
}

type dedicatedStorageResourceModel struct {
	dedicatedStorageBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *dedicatedStorageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Dedicated Storage"),
			"name":        common.SchemaResourceName("Dedicated Storage"),
			"description": common.SchemaResourceDescription("Dedicated Storage"),
			"tags":        common.SchemaResourceTags("Dedicated Storage"),
			"icon_id":     common.SchemaResourceIconID("Dedicated Storage"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Dedicated Storage contract.",
	}
}

func (r *dedicatedStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dedicatedStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dedicatedStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	dsOp := dedicatedstorage.NewContractOp(r.client)

	dsPlans, err := dsOp.ListPlans(ctx)
	if err != nil || len(dsPlans.DedicatedStorageContractPlans) == 0 {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read DedicatedStorage Plans: %s", err))
		return
	}

	res, err := dsOp.Create(ctx, expandDedicatedStorageCreateRequest(plan, dsPlans.DedicatedStorageContractPlans[0].ID))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create DedicatedStorage: %s", err))
		return
	}
	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dedicatedStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dedicatedStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dsOp := dedicatedstorage.NewContractOp(r.client)
	id := common.ExpandSakuraCloudID(state.ID).Int64()
	data, err := dsOp.Read(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read DedicatedStorage[%s]: %s", state.ID.ValueString(), err))
		return
	}

	state.updateState(data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dedicatedStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dedicatedStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	dsOp := dedicatedstorage.NewContractOp(r.client)
	id := common.SakuraCloudID(plan.ID.ValueString()).Int64()
	res, err := dsOp.Update(ctx, id, expandDedicatedStorageUpdateRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update DedicatedStorage[%s]: %s", plan.ID.ValueString(), err))
		return
	}
	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dedicatedStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dedicatedStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	dsOp := dedicatedstorage.NewContractOp(r.client)
	id := common.SakuraCloudID(state.ID.ValueString()).Int64()
	if err := dsOp.Delete(ctx, id); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete DedicatedStorage[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func expandDedicatedStorageCreateRequest(plan dedicatedStorageResourceModel, planId int64) v1.CreateDedicatedStorageContractRequest {
	createReq := v1.CreateDedicatedStorageContractRequest{
		DedicatedStorageContract: v1.CreateDedicatedStorageContractRequestDedicatedStorageContract{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
			Tags:        common.TsetToStrings(plan.Tags),
			Plan: v1.CreateDedicatedStorageContractRequestDedicatedStorageContractPlan{
				ID: planId,
			},
			Icon: v1.NewOptNilIcon(v1.Icon{}),
		},
	}
	if !plan.IconID.IsNull() && !plan.IconID.IsUnknown() {
		createReq.DedicatedStorageContract.Icon.SetTo(v1.Icon{
			ID: common.ExpandSakuraCloudID(plan.IconID).Int64(),
		})
	}
	return createReq
}

func expandDedicatedStorageUpdateRequest(plan dedicatedStorageResourceModel) v1.UpdateDedicatedStorageContractRequest {
	updateReq := v1.UpdateDedicatedStorageContractRequest{
		DedicatedStorageContract: v1.UpdateDedicatedStorageContractRequestDedicatedStorageContract{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
			Tags:        common.TsetToStrings(plan.Tags),
			Icon:        v1.NewOptNilIcon(v1.Icon{}),
		},
	}
	if !plan.IconID.IsNull() && !plan.IconID.IsUnknown() {
		updateReq.DedicatedStorageContract.Icon.SetTo(v1.Icon{
			ID: common.ExpandSakuraCloudID(plan.IconID).Int64(),
		})
	}
	return updateReq
}
