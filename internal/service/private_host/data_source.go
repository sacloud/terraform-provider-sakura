// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package private_host

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type privateHostDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &privateHostDataSource{}
	_ datasource.DataSourceWithConfigure = &privateHostDataSource{}
)

func NewPrivateHostDataSource() datasource.DataSource {
	return &privateHostDataSource{}
}

func (d *privateHostDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_host"
}

func (d *privateHostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type privateHostDataSourceModel struct {
	privateHostBaseModel
}

func (d *privateHostDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("PrivateHost"),
			"name":        common.SchemaDataSourceName("PrivateHost"),
			"description": common.SchemaDataSourceDescription("PrivateHost"),
			"tags":        common.SchemaDataSourceTags("PrivateHost"),
			"zone":        common.SchemaDataSourceZone("PrivateHost"),
			"icon_id":     common.SchemaDataSourceIconID("PrivateHost"),
			"class":       common.SchemaDataSourceClass("PrivateHost", []string{iaastypes.PrivateHostClassDynamic, iaastypes.PrivateHostClassWindows}),
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the private host.",
			},
			"assigned_core": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPUs assigned to servers on the private host",
			},
			"assigned_memory": schema.Int32Attribute{
				Computed:    true,
				Description: "The total size of memory assigned to servers on the private host",
			},
		},
		MarkdownDescription: "Get information about an existing Private Host.",
	}
}

func (d *privateHostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data privateHostDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewPrivateHostOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.PrivateHosts) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	ph := res.PrivateHosts[0]
	data.updateState(ph, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
