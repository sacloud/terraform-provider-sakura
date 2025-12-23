// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package bridge

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type bridgeDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &bridgeDataSource{}
	_ datasource.DataSourceWithConfigure = &bridgeDataSource{}
)

func NewBridgeDataSource() datasource.DataSource {
	return &bridgeDataSource{}
}

func (d *bridgeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bridge"
}

func (d *bridgeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type bridgeDataSourceModel struct {
	bridgeBaseModel
}

func (d *bridgeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Bridge"),
			"name":        common.SchemaDataSourceName("Bridge"),
			"description": common.SchemaDataSourceDescription("Bridge"),
			"zone":        common.SchemaDataSourceZone("Bridge"),
		},
		MarkdownDescription: "Get information about an existing Bridge.",
	}
}

func (d *bridgeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bridgeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridgeOp := iaas.NewBridgeOp(d.client)
	res, err := bridgeOp.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, types.SetNull(types.StringType)))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find Bridge: %s", err))
		return
	}
	if res == nil || len(res.Bridges) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.Bridges[0], zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
