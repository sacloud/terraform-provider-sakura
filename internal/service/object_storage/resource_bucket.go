// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
				},
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
		MarkdownDescription: "Manages an Object Storage's Bucket.",
	}
}

func (r *objectStorageBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 2)

	if len(parts) != 2 {
		resp.Diagnostics.AddError("Import Error",
			fmt.Sprintf("invalid import ID format. Please specify the import ID in the format of {site_id}_{name}: %s", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *objectStorageBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStorageBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	siteId := plan.SiteID.ValueString()
	bucketAPI := objectstorage.NewBucketOp(r.client)
	bucket, err := bucketAPI.Create(ctx, siteId, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Object Storage Bucket: %s", err))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s_%s", siteId, bucket.Name))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStorageBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	siteId := state.SiteID.ValueString()

	// 将来的にはBucketの情報をAPIから取得して状態に反映
	state.ID = types.StringValue(fmt.Sprintf("%s_%s", siteId, name))
	state.Name = types.StringValue(name)
	state.SiteID = types.StringValue(siteId)

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
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Object Storage Bucket: %s", err))
		return
	}
}
