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
	objectstorage "github.com/sacloud/object-storage-api-go"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageBucketEncryptionConfigResource struct {
	fedClient *objectstorage.FedClient
	saClient  saclient.ClientAPI
}

var (
	_ resource.Resource              = &objectStorageBucketEncryptionConfigResource{}
	_ resource.ResourceWithConfigure = &objectStorageBucketEncryptionConfigResource{}
)

func NewObjectStorageBucketEncryptionConfigResource() resource.Resource {
	return &objectStorageBucketEncryptionConfigResource{}
}

func (r *objectStorageBucketEncryptionConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_bucket_encryption_config"
}

func (r *objectStorageBucketEncryptionConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.fedClient = apiclient.ObjectStorageFedClient
	r.saClient = apiclient.SaClient
}

type objectStorageBucketEncryptionConfigResourceModel struct {
	objectStorageBucketEncryptionBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *objectStorageBucketEncryptionConfigResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("Object Storage Encryption Configuration"),
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "The source bucket name for encryption.",
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
			"kms_key_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the KMS key for encryption.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Object Storage's Encryption Configuration.",
	}
}

func (r *objectStorageBucketEncryptionConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 2)

	if len(parts) != 2 {
		resp.Diagnostics.AddError("Import Error",
			fmt.Sprintf("invalid import ID format. Please specify the import ID in the format of {site_id}_{bucket}: %s", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bucket"), parts[1])...)
}

func (r *objectStorageBucketEncryptionConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStorageBucketEncryptionConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	siteId := plan.SiteID.ValueString()
	bucket := plan.Bucket.ValueString()
	keyId := plan.KmsKeyID.ValueString()
	siteClient, err := objectstorage.NewSiteClient(r.saClient, siteId)
	if err != nil {
		resp.Diagnostics.AddError("Create: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err))
		return
	}

	exOp := objectstorage.NewBucketExtraOp(siteClient, r.fedClient, bucket)
	if err := exOp.EnableEncryption(ctx, keyId); err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to enable encryption for Object Storage Bucket[%s]: %s.", bucket, err))
		return
	}

	plan.updateState(siteId, bucket, keyId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketEncryptionConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStorageBucketEncryptionConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteId := state.SiteID.ValueString()
	bucket := state.Bucket.ValueString()
	keyId := ""

	siteClient, err := objectstorage.NewSiteClient(r.saClient, siteId)
	if err != nil {
		resp.Diagnostics.AddError("Read: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err))
		return
	}
	exOp := objectstorage.NewBucketExtraOp(siteClient, r.fedClient, bucket)
	if res, err := exOp.ReadEncryption(ctx); err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read encryption settings for Object Storage Bucket[%s]: %s", bucket, err))
		return
	} else {
		keyId = string(res.KmsKeyID.Value)
	}

	state.updateState(siteId, bucket, keyId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStorageBucketEncryptionConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan objectStorageBucketEncryptionConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	siteId := plan.SiteID.ValueString()
	bucket := plan.Bucket.ValueString()
	keyId := plan.KmsKeyID.ValueString()
	siteClient, err := objectstorage.NewSiteClient(r.saClient, siteId)
	if err != nil {
		resp.Diagnostics.AddError("Update: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err))
		return
	}

	exOp := objectstorage.NewBucketExtraOp(siteClient, r.fedClient, bucket)
	if err := exOp.EnableEncryption(ctx, keyId); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update encryption for Object Storage Bucket[%s]: %s.", bucket, err))
		return
	}

	plan.updateState(siteId, bucket, keyId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketEncryptionConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectStorageBucketEncryptionConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	bucket := state.Bucket.ValueString()

	siteClient, err := objectstorage.NewSiteClient(r.saClient, state.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err.Error()))
		return
	}

	bucketAPI := objectstorage.NewBucketExtraOp(siteClient, r.fedClient, bucket)
	if err := bucketAPI.DisableEncryption(ctx); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to disable replication for Object Storage Bucket[%s]: %s", bucket, err))
		return
	}
}
