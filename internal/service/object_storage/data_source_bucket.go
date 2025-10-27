// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	objectstorage "github.com/sacloud/object-storage-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageBucketDataSource struct {
	client *objectstorage.Client
}

var (
	_ datasource.DataSource              = &objectStorageBucketDataSource{}
	_ datasource.DataSourceWithConfigure = &objectStorageBucketDataSource{}
)

func NewObjectStorageBucketDataSource() datasource.DataSource {
	return &objectStorageBucketDataSource{}
}

func (d *objectStorageBucketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_bucket"
}

func (d *objectStorageBucketDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.ObjectStorageClient
}

type objectStorageBucketDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	SiteID types.String `tfsdk:"site_id"`
}

func (d *objectStorageBucketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaDataSourceId("Object Storage Bucket"),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Object Storage Bucket.",
			},
			"site_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Object Storage Site.",
			},
		},
	}
}

func (d *objectStorageBucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data objectStorageBucketDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: APIを呼び出しての存在チェック。現状はobject-storage-api-goにAPIがないため未実装

	data.ID = types.StringValue(data.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
