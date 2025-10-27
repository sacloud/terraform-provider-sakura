// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/minio/minio-go/v7"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageObjectDataSource struct{}

var (
	_ datasource.DataSource = &objectStorageObjectDataSource{}
)

func NewObjectStorageObjectDataSource() datasource.DataSource {
	return &objectStorageObjectDataSource{}
}

func (d *objectStorageObjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_object"
}

type objectStorageObjectDataSourceModel struct {
	objectStorageObjectBaseModel
	Body           types.String `tfsdk:"body"`
	Size           types.Int64  `tfsdk:"size"`
	ACL            types.String `tfsdk:"acl"`
	Expires        types.String `tfsdk:"expires"`
	Metadata       types.Map    `tfsdk:"metadata"`
	UserMetadata   types.Map    `tfsdk:"user_metadata"`
	UserTags       types.Map    `tfsdk:"user_tags"`
	UserTagCount   types.Int64  `tfsdk:"user_tag_count"`
	NumVersions    types.Int64  `tfsdk:"num_versions"`
	IsLatest       types.Bool   `tfsdk:"is_latest"`
	IsDeleteMarker types.Bool   `tfsdk:"is_delete_marker"`
}

func (d *objectStorageObjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaDataSourceId("ObjectStorage Object"),
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The endpoint of the object storage site",
			},
			"access_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The access key for the object storage site",
			},
			"secret_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The secret key for the object storage site",
			},
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "The name of the ObjectStorage Bucket",
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "The key of the ObjectStorage Object",
			},
			"body": schema.StringAttribute{
				Computed:    true,
				Description: "The content of the ObjectStorage Object",
			},
			"etag": schema.StringAttribute{
				Computed:    true,
				Description: "The ETag of the ObjectStorage Object",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "The content size of the ObjectStorage Object",
			},
			"acl": schema.StringAttribute{
				Computed:    true,
				Description: "The ACL of the ObjectStorage Object",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "The last modified time of the ObjectStorage Object",
			},
			"expires": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration time of the ObjectStorage Object cache",
			},
			"content_type": schema.StringAttribute{
				Computed:    true,
				Description: "The content type of the ObjectStorage Object",
			},
			"storage_class": schema.StringAttribute{
				Computed:    true,
				Description: "The storage class of the ObjectStorage Object",
			},
			"metadata": schema.MapAttribute{
				Computed:    true,
				ElementType: types.ListType{ElemType: types.StringType},
				Description: "A metadata of the ObjectStorage Object",
			},
			"user_metadata": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "User-defined metadata for the ObjectStorage Object",
			},
			"user_tags": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "User-defined tags for the ObjectStorage Object",
			},
			"user_tag_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of user-defined tags for the ObjectStorage Object",
			},
			"version_id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The version ID of the ObjectStorage Object",
			},
			"num_versions": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of versions of the ObjectStorage Object",
			},
			"is_latest": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the ObjectStorage Object is the latest version",
			},
			"is_delete_marker": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the ObjectStorage Object is a delete marker",
			},
		},
	}
}

func (d *objectStorageObjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data objectStorageObjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	minioClient, err := data.getMinIOClient()
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to create MinIO client: %s", err.Error()))
		return
	}
	opts := minio.GetObjectOptions{}
	if data.VersionID.ValueString() != "" {
		opts.VersionID = data.VersionID.ValueString()
	}
	obj, err := minioClient.GetObject(ctx, data.Bucket.ValueString(), data.Key.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to get object: %s", err.Error()))
		return
	}
	info, err := obj.Stat()
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to get object information: %s", err.Error()))
		return
	}
	aclInfo, err := minioClient.GetObjectACL(ctx, data.Bucket.ValueString(), data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to get object ACL: %s", err.Error()))
		return
	}
	content, err := io.ReadAll(obj)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to read object data: %s", err.Error()))
		return
	}

	metadata, diag := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, info.Metadata)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.updateState(&info)
	data.Body = types.StringValue(string(content))
	data.Size = types.Int64Value(info.Size)
	data.ACL = types.StringValue(aclInfo.Metadata["X-Amz-Acl"][0])
	data.Expires = types.StringValue(info.Expires.String())
	data.Metadata = metadata
	data.UserMetadata = common.StrMapToTmap(info.UserMetadata)
	data.UserTags = common.StrMapToTmap(info.UserTags)
	data.UserTagCount = types.Int64Value(int64(info.UserTagCount))
	data.NumVersions = types.Int64Value(int64(info.NumVersions))
	data.IsLatest = types.BoolValue(info.IsLatest)
	data.IsDeleteMarker = types.BoolValue(info.IsDeleteMarker)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
