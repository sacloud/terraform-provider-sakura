// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	objectstorage "github.com/sacloud/object-storage-api-go"
	v2 "github.com/sacloud/object-storage-api-go/apis/v2"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStoragePermissionResource struct {
	saClient saclient.ClientAPI
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
	r.saClient = apiclient.SaClient
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
		MarkdownDescription: "Manages an Object Storage's Permission.",
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

	siteClient, err := objectstorage.NewSiteClient(r.saClient, plan.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err.Error()))
		return
	}

	permissionOp := objectstorage.NewPermissionOp(siteClient)
	permission, err := permissionOp.Create(ctx, plan.Name.ValueString(), getBucketControls(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Object Storage Permission: %s", err.Error()))
		return
	}

	permissionID := strconv.FormatInt(int64(permission.ID.Value), 10)
	accessKey, err := permissionOp.CreateAccessKey(ctx, permissionID)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Object Storage Permission Access Key: %s", err.Error()))
		return
	}

	plan.ID = types.StringValue(permissionID)
	plan.AccessKey = types.StringValue(string(accessKey.ID.Value))
	plan.SecretKey = types.StringValue(string(accessKey.Secret.Value))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStoragePermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStoragePermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteClient, err := objectstorage.NewSiteClient(r.saClient, state.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err.Error()))
		return
	}

	permissionOp := objectstorage.NewPermissionOp(siteClient)
	permission, err := permissionOp.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Object Storage Permission: %s", err.Error()))
		return
	}

	state.Name = types.StringValue(string(permission.DisplayName.Value))
	state.BucketControls = make([]*objectStorageBucketControlModel, 0, len(permission.BucketControls))
	for _, bc := range permission.BucketControls {
		state.BucketControls = append(state.BucketControls, &objectStorageBucketControlModel{
			Bucket:   types.StringValue(string(bc.BucketName.Value)),
			CanRead:  types.BoolValue(bool(bc.CanRead.Value)),
			CanWrite: types.BoolValue(bool(bc.CanWrite.Value)),
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

	siteClient, err := objectstorage.NewSiteClient(r.saClient, plan.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err.Error()))
		return
	}

	permissionOp := objectstorage.NewPermissionOp(siteClient)
	_, err = permissionOp.Update(ctx, plan.ID.ValueString(), plan.Name.ValueString(), getBucketControls(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Object Storage Permission: %s", err.Error()))
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

	siteClient, err := objectstorage.NewSiteClient(r.saClient, state.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete: Client Error", fmt.Sprintf("failed to create Object Storage site client: %s", err.Error()))
		return
	}

	permissionOp := objectstorage.NewPermissionOp(siteClient)
	if err := permissionOp.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Object Storage Permission: %s", err.Error()))
		return
	}
}

func getBucketControls(model *objectStoragePermissionResourceModel) v2.BucketControls {
	bucketControls := make(v2.BucketControls, 0, len(model.BucketControls))
	for _, bc := range model.BucketControls {
		bucketControls = append(bucketControls, v2.BucketControlsItem{
			BucketName: v2.NewOptBucketName(v2.BucketName(bc.Bucket.ValueString())),
			CanRead:    v2.NewOptCanRead(v2.CanRead(bc.CanRead.ValueBool())),
			CanWrite:   v2.NewOptCanWrite(v2.CanWrite(bc.CanWrite.ValueBool())),
		})
	}
	return bucketControls
}
