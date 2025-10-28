// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dns

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type dnsDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &dnsDataSource{}
	_ datasource.DataSourceWithConfigure = &dnsDataSource{}
)

func NewDNSDataSource() datasource.DataSource {
	return &dnsDataSource{}
}

func (d *dnsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns"
}

func (d *dnsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type dnsDataSourceModel struct {
	dnsBaseModel
}

func (d *dnsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("DNS"),
			"description": common.SchemaDataSourceDescription("DNS"),
			"tags":        common.SchemaDataSourceTags("DNS"),
			"icon_id":     common.SchemaDataSourceIconID("DNS"),
			"zone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of managed domain",
			},
			"dns_servers": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of IP address of DNS server that manage this zone",
			},
			"record": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of DNS Record",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The type of DNS Record. This will be one of [%s]", iaastypes.DNSRecordTypeStrings),
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the DNS Record",
						},
						"ttl": schema.Int64Attribute{
							Computed:    true,
							Description: "The number of the TTL",
						},
						"priority": schema.Int32Attribute{
							Computed:    true,
							Description: "The priority of target DNS Record",
						},
						"weight": schema.Int32Attribute{
							Computed:    true,
							Description: "The weight of target DNS Record",
						},
						"port": schema.Int32Attribute{
							Computed:    true,
							Description: "The number of port",
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing DNS.",
	}
}

func (d *dnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dnsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewDNSOp(d.client)
	// name attribute is needed?
	res, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, types.StringNull(), data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Search DNS Error", "could not find SakuraCloud DNS resource: "+err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.DNS) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.DNS[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
