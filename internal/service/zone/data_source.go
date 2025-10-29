// Copyright 2016-2025 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package zone

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type zoneDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &zoneDataSource{}
	_ datasource.DataSourceWithConfigure = &zoneDataSource{}
)

func NewZoneDataSource() datasource.DataSource {
	return &zoneDataSource{}
}

func (d *zoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *zoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type zoneDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ZoneID      types.String `tfsdk:"zone_id"`
	RegionID    types.String `tfsdk:"region_id"`
	RegionName  types.String `tfsdk:"region_name"`
	DNSServers  types.List   `tfsdk:"dns_servers"`
}

func (d *zoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Zone"),
			"name":        common.SchemaDataSourceName("Zone"),
			"description": common.SchemaDataSourceDescription("Zone"),
			"zone_id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the zone",
			},
			"region_id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the region that the zone belongs",
			},
			"region_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the region that the zone belongs",
			},
			"dns_servers": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of IP address of DNS server in the zone",
			},
		},
	}
}

func (d *zoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data zoneDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Name, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneOp := iaas.NewZoneOp(d.client)
	res, err := zoneOp.Find(ctx, &iaas.FindCondition{})
	if err != nil {
		resp.Diagnostics.AddError("Read: Find Error", fmt.Sprintf("failed to find SakuraCloud Zone resource: %s", err))
		return
	}
	if res == nil || len(res.Zones) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	var found *iaas.Zone
	for _, z := range res.Zones {
		if z.Name == zone {
			found = z
			break
		}
	}
	if found == nil {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.ID = types.StringValue(found.ID.String())
	data.Name = types.StringValue(found.Name)
	data.ZoneID = types.StringValue(found.ID.String())
	data.Description = types.StringValue(found.Description)
	data.RegionID = types.StringValue(found.Region.ID.String())
	data.RegionName = types.StringValue(found.Region.Name)
	data.DNSServers = common.StringsToTlist(found.Region.NameServers)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
