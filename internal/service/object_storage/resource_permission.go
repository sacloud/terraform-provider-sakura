// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	objectstorage "github.com/sacloud/object-storage-api-go"
	v1 "github.com/sacloud/object-storage-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStoragePermissionResource struct {
	client *objectstorage.Client
}

var (
	_ resource.Resource                = &objectStoragePermissionResource{}
	_ resource.ResourceWithConfigure   = &objectStoragePermissionResource{}
	_ resource.ResourceWithImportState = &objectStoragePermissionResource{}
)

func NewObjectStoragePermissionResource() resource.Resource {
	return &objectStoragePermissionResource{}
}

func (r *objectStoragePermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_permission"
}

func (r *objectStoragePermissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ObjectStorageClient
}

type objectStoragePermissionResourceModel struct {
	ID             types.String                       `tfsdk:"id"`
	Name           types.String                       `tfsdk:"name"`
	SiteID         types.String                       `tfsdk:"site_id"`
	AccessKey      types.String                       `tfsdk:"access_key"`
	SecretKey      types.String                       `tfsdk:"secret_key"`
	BucketControls []*objectStorageBucketControlModel `tfsdk:"bucket_controls"`
	Timeouts       timeouts.Value                     `tfsdk:"timeouts"`
}

type objectStorageBucketControlModel struct {
	Bucket   types.String `tfsdk:"bucket"`
	CanRead  types.Bool   `tfsdk:"can_read"`
	CanWrite types.Bool   `tfsdk:"can_write"`
}

func (r *objectStoragePermissionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":      common.SchemaResourceId("Object Storage Permission"),
			"site_id": SchemaResourceSiteID("Object Storage Permission"),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Object Storage Permission.",
			},
			"access_key": schema.StringAttribute{
				Computed:    true,
				Description: "The access key for the Object Storage Permission.",
			},
			"secret_key": schema.StringAttribute{
				Computed:    true,
				Description: "The secret key for the Object Storage Permission.",
			},
			"bucket_controls": schema.ListNestedAttribute{
				Required:    true,
				Description: "The bucket controls for the Object Storage Permission.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"bucket": schema.StringAttribute{
							Required:    true,
							Description: "The bucket name for the Object Storage Permission.",
						},
						"can_read": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the permission can read from the bucket.",
							Default:     booldefault.StaticBool(true),
						},
						"can_write": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the permission can write to the bucket.",
							Default:     booldefault.StaticBool(true),
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *objectStoragePermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *objectStoragePermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStoragePermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	permissionOp := objectstorage.NewPermissionOp(r.client)
	permission, err := permissionOp.Create(ctx, plan.SiteID.ValueString(), &v1.CreatePermissionParams{
		BucketControls: getBucketControls(&plan),
		DisplayName:    v1.DisplayName(plan.Name.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create Object Storage Permission: %s", err.Error()))
		return
	}

	accessKey, err := permissionOp.CreateAccessKey(ctx, plan.SiteID.ValueString(), permission.Id.Int64())
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create Object Storage Permission Access Key: %s", err.Error()))
		return
	}

	plan.ID = types.StringValue(permission.Id.String())
	plan.AccessKey = types.StringValue(accessKey.Id.String())
	plan.SecretKey = types.StringValue(accessKey.Secret.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStoragePermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStoragePermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	permissionOp := objectstorage.NewPermissionOp(r.client)
	permission, err := permissionOp.Read(ctx, state.SiteID.ValueString(), common.MustAtoInt64(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to read Object Storage Permission: %s", err.Error()))
		return
	}

	state.Name = types.StringValue(permission.DisplayName.String())
	state.BucketControls = make([]*objectStorageBucketControlModel, 0, len(permission.BucketControls))
	for _, bc := range permission.BucketControls {
		state.BucketControls = append(state.BucketControls, &objectStorageBucketControlModel{
			Bucket:   types.StringValue(bc.BucketName.String()),
			CanRead:  types.BoolValue(bool(bc.CanRead)),
			CanWrite: types.BoolValue(bool(bc.CanWrite)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStoragePermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan objectStoragePermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	permissionOp := objectstorage.NewPermissionOp(r.client)
	_, err := permissionOp.Update(ctx, plan.SiteID.ValueString(), common.MustAtoInt64(plan.ID.ValueString()), &v1.UpdatePermissionParams{
		BucketControls: getBucketControls(&plan),
		DisplayName:    v1.DisplayName(plan.Name.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update Object Storage Permission: %s", err.Error()))
		return
	}

	// Need AccessKey/SecretKey update?

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStoragePermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectStoragePermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	permissionOp := objectstorage.NewPermissionOp(r.client)
	if err := permissionOp.Delete(ctx, state.SiteID.ValueString(), common.MustAtoInt64(state.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete Object Storage Permission: %s", err.Error()))
		return
	}
}

func getBucketControls(model *objectStoragePermissionResourceModel) []v1.BucketControl {
	bucketControls := make([]v1.BucketControl, 0, len(model.BucketControls))
	for _, bc := range model.BucketControls {
		bucketControls = append(bucketControls, v1.BucketControl{
			BucketName: v1.BucketName(bc.Bucket.ValueString()),
			CanRead:    v1.CanRead(bc.CanRead.ValueBool()),
			CanWrite:   v1.CanWrite(bc.CanWrite.ValueBool()),
		})
	}
	return bucketControls
}
