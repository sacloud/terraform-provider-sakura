// Copyright 2016-2026 The terraform-provider-sakura Authors
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
	objectstorage "github.com/sacloud/object-storage-api-go"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageBucketReplicationConfigResource struct {
	fedClient *objectstorage.FedClient
	saClient  saclient.ClientAPI
}

var (
	_ resource.Resource              = &objectStorageBucketResource{}
	_ resource.ResourceWithConfigure = &objectStorageBucketResource{}
)

func NewObjectStorageBucketReplicationConfigResource() resource.Resource {
	return &objectStorageBucketReplicationConfigResource{}
}

func (r *objectStorageBucketReplicationConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_bucket_replication_config"
}

func (r *objectStorageBucketReplicationConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.fedClient = apiclient.ObjectStorageFedClient
	r.saClient = apiclient.SaClient
}

type objectStorageBucketReplicationConfigResourceModel struct {
	objectStorageBucketReplicationBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *objectStorageBucketReplicationConfigResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("Object Storage Replication Configuration"),
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "The source bucket name for replication.",
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
			"destination_bucket": schema.StringAttribute{
				Required:    true,
				Description: "The destination bucket name for replication.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Object Storage's Replication Configuration.",
	}
}

func (r *objectStorageBucketReplicationConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 2)

	if len(parts) != 2 {
		resp.Diagnostics.AddError("Import Error",
			fmt.Sprintf("invalid import ID format. Please specify the import ID in the format of {site_id}_{bucket}: %s", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bucket"), parts[1])...)
}

func (r *objectStorageBucketReplicationConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStorageBucketReplicationConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	siteId := plan.SiteID.ValueString()
	src := plan.Bucket.ValueString()
	dst := plan.DestinationBucket.ValueString()
	siteClient, err := objectstorage.NewSiteClient(r.saClient, siteId)
	if err != nil {
		resp.Diagnostics.AddError("Create: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err))
		return
	}

	exOp := objectstorage.NewBucketExtraOp(siteClient, r.fedClient, src)
	if _, err := exOp.EnableReplication(ctx, dst); err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to enable replication for Object Storage Bucket[%s]: %s.", src, err))
		return
	}

	plan.updateState(siteId, src, dst)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketReplicationConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStorageBucketReplicationConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteId := state.SiteID.ValueString()
	src := state.Bucket.ValueString()
	dst := ""

	siteClient, err := objectstorage.NewSiteClient(r.saClient, siteId)
	if err != nil {
		resp.Diagnostics.AddError("Read: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err))
		return
	}
	exOp := objectstorage.NewBucketExtraOp(siteClient, r.fedClient, src)
	if res, err := exOp.ReadReplication(ctx); err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read replication settings for Object Storage Bucket[%s]: %s", src, err))
		return
	} else {
		dst = res.DestBucket.Name.Value
	}

	state.updateState(siteId, src, dst)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStorageBucketReplicationConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Object Storage Bucket Replication Configuration does not support updates.")
}

func (r *objectStorageBucketReplicationConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectStorageBucketReplicationConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	src := state.Bucket.ValueString()

	siteClient, err := objectstorage.NewSiteClient(r.saClient, state.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err.Error()))
		return
	}

	bucketAPI := objectstorage.NewBucketExtraOp(siteClient, r.fedClient, src)
	if err := bucketAPI.DisableReplication(ctx); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to disable replication for Object Storage Bucket[%s]: %s", src, err))
		return
	}
}
