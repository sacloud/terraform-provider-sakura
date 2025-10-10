// Copyright 2016-2025 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/minio/minio-go/v7"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageBucketVersioningResource struct{}

var (
	_ resource.Resource                = &objectStorageBucketVersioningResource{}
	_ resource.ResourceWithImportState = &objectStorageBucketVersioningResource{}
)

func NewObjectStorageBucketVersioningResource() resource.Resource {
	return &objectStorageBucketVersioningResource{}
}

func (r *objectStorageBucketVersioningResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_bucket_versioning"
}

type objectStorageBucketVersioningResourceModel struct {
	objectStorageS3CompatModel
	VersioningConfiguration *objectStorageBucketVersioningConfigModel `tfsdk:"versioning_configuration"`
	Timeouts                timeouts.Value                            `tfsdk:"timeouts"`
}

type objectStorageBucketVersioningConfigModel struct {
	Status types.String `tfsdk:"status"`
	//MFADelete types.String `tfsdk:"mfa_delete"`
}

func (r *objectStorageBucketVersioningResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("ObjectStorage Bucket Versioning"),
			"endpoint":   SchemaResourceEndpoint("Object Storage Bucket Versioning"),
			"access_key": SchemaResourceAccessKey("Object Storage Bucket Versioning"),
			"secret_key": SchemaResourceSecretKey("Object Storage Bucket Versioning"),
			"bucket":     SchemaResourceBucket("Object Storage Bucket Versioning"),
			"versioning_configuration": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The versioning configuration for the Object Storage Bucket.",
				Attributes: map[string]schema.Attribute{
					"status": schema.StringAttribute{
						Required:    true,
						Description: "The versioning status of the Object Storage Bucket. Supported values are `Enabled` and `Suspended`.",
						Validators: []validator.String{
							stringvalidator.OneOf(minio.Enabled, minio.Suspended),
						},
					},
					/*
						"mfa_delete": schema.StringAttribute{
							Optional:    true,
							Description: "The MFA delete status of the Object Storage Bucket. Supported values are `Enabled` and `Disabled`.",
							Validators: []validator.String{
								stringvalidator.OneOf("Enabled", "Disabled"),
							},
						},
					*/
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *objectStorageBucketVersioningResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *objectStorageBucketVersioningResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStorageBucketVersioningResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	if err := setBucketVersioningConfiguration(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	plan.updateS3State()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketVersioningResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStorageBucketVersioningResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := state.getMinIOClient()
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Errorf("failed to create MinIO client: %w", err).Error())
		return
	}

	versioningConfig, err := client.GetBucketVersioning(ctx, state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Errorf("failed to get bucket versioning: %w", err).Error())
		return
	}

	state.updateS3State()
	state.VersioningConfiguration.Status = types.StringValue(versioningConfig.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStorageBucketVersioningResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan objectStorageBucketVersioningResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	if err := setBucketVersioningConfiguration(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}

	plan.updateS3State()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketVersioningResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectStorageBucketVersioningResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := state.getMinIOClient()
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Errorf("failed to create MinIO client: %w", err).Error())
		return
	}

	err = client.SuspendVersioning(ctx, state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to suspend bucket versioning: %s", err.Error()))
		return
	}
}

func setBucketVersioningConfiguration(ctx context.Context, model *objectStorageBucketVersioningResourceModel) error {
	client, err := model.getMinIOClient()
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %w", err)
	}

	if model.VersioningConfiguration.Status.ValueString() == minio.Enabled {
		err = client.EnableVersioning(ctx, model.Bucket.ValueString())
	} else {
		err = client.SuspendVersioning(ctx, model.Bucket.ValueString())
	}
	if err != nil {
		return fmt.Errorf("failed to set bucket versioning: %w", err)
	}

	return nil
}
