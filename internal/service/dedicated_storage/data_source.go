// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dedicated_storage

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	dedicatedstorage "github.com/sacloud/dedicated-storage-api-go"
	v1 "github.com/sacloud/dedicated-storage-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type dedicatedStorageDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &dedicatedStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &dedicatedStorageDataSource{}
)

func NewDedicatedStorageDataSource() datasource.DataSource {
	return &dedicatedStorageDataSource{}
}

func (d *dedicatedStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dedicated_storage"
}

func (d *dedicatedStorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	d.client = apiClient.DedicatedStorageClient
}

type dedicatedStorageDataSourceModel struct {
	dedicatedStorageBaseModel
}

func (d *dedicatedStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Dedicated Storage"),
			"name":        common.SchemaDataSourceName("Dedicated Storage"),
			"description": common.SchemaDataSourceDescription("Dedicated Storage"),
			"tags":        common.SchemaDataSourceTags("Dedicated Storage"),
			"icon_id":     common.SchemaDataSourceIconID("Dedicated Storage"),
		},
		MarkdownDescription: "Get information about an existing Dedicated Storage contract.",
	}
}

func (d *dedicatedStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dedicatedStorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dsOp := dedicatedstorage.NewContractOp(d.client)
	var res *v1.DedicatedStorageContract
	var err error
	if !data.Name.IsNull() {
		contracts, err := dsOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list DedicatedStorage resources: %s", err))
			return
		}
		res, err = filterDedicatedStorageByName(contracts.DedicatedStorageContracts, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		id := common.SakuraCloudID(data.ID.ValueString()).Int64()
		res, err = dsOp.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read DedicatedStorage[%s] resource: %s", data.ID.ValueString(), err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterDedicatedStorageByName(contracts []v1.DedicatedStorageContract, name string) (*v1.DedicatedStorageContract, error) {
	match := slices.Collect(func(yield func(contract v1.DedicatedStorageContract) bool) {
		for _, v := range contracts {
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
		return nil, fmt.Errorf("multiple DedicatedStorage resources found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
