// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iam-api-go"
	"github.com/sacloud/iam-api-go/apis/folder"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type folderDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &folderDataSource{}
	_ datasource.DataSourceWithConfigure = &folderDataSource{}
)

func NewFolderDataSource() datasource.DataSource {
	return &folderDataSource{}
}

func (d *folderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_folder"
}

func (d *folderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type folderDataSourceModel struct {
	folderBaseModel
}

func (d *folderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM Folder"),
			"name":        common.SchemaDataSourceName("IAM Folder"),
			"description": common.SchemaDataSourceDescription("IAM Folder"),
			"parent_id": schema.StringAttribute{
				Computed:    true,
				Description: "The parent ID of the IAM Folder",
			},
			"created_at": common.SchemaDataSourceCreatedAt("IAM Folder"),
			"updated_at": common.SchemaDataSourceUpdatedAt("IAM Folder"),
		},
	}
}

func (d *folderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data folderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *v1.Folder
	var err error
	folderOp := iam.NewFolderOp(d.client)
	if utils.IsKnown(data.Name) {
		perPage := 100 // TODO: Proper pagination if needed
		folders, err := folderOp.List(ctx, folder.ListParams{PerPage: &perPage})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list IAM Folder resources: %s", err))
			return
		}
		res, err = filterIAMFolderByName(folders.Items, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = folderOp.Read(ctx, utils.MustAtoI(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM Folder resource: %s", err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterIAMFolderByName(keys []v1.Folder, name string) (*v1.Folder, error) {
	match := slices.Collect(func(yield func(v1.Folder) bool) {
		for _, v := range keys {
			if name != v.Name {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple IAM Fold	ers found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
