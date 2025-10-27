// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	objectstorage "github.com/sacloud/object-storage-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageBucketResource struct {
	client *objectstorage.Client
}

var (
	_ resource.Resource                = &objectStorageBucketResource{}
	_ resource.ResourceWithConfigure   = &objectStorageBucketResource{}
	_ resource.ResourceWithImportState = &objectStorageBucketResource{}
)

func NewObjectStorageBucketResource() resource.Resource {
	return &objectStorageBucketResource{}
}

func (r *objectStorageBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_bucket"
}

func (r *objectStorageBucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ObjectStorageClient
}

type objectStorageBucketResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	SiteID   types.String   `tfsdk:"site_id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *objectStorageBucketResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("Object Storage Bucket"),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Object Storage Bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"site_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Object Storage Site.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *objectStorageBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *objectStorageBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStorageBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	bucketAPI := objectstorage.NewBucketOp(r.client)
	bucket, err := bucketAPI.Create(ctx, plan.SiteID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create Object Storage Bucket: %s", err.Error()))
		return
	}

	plan.ID = types.StringValue(bucket.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStorageBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 将来的にはBucketの情報をAPIから取得して状態に反映

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStorageBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan objectStorageBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	// name/site_idがRequiresReplaceでre-createするので、現状更新する項目は無し

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectStorageBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	bucketAPI := objectstorage.NewBucketOp(r.client)
	if err := bucketAPI.Delete(ctx, state.SiteID.ValueString(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete Object Storage Bucket: %s", err.Error()))
		return
	}
}
