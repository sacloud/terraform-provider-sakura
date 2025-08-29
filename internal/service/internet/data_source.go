// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type internetDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &internetDataSource{}
	_ datasource.DataSourceWithConfigure = &internetDataSource{}
)

func NewInternetDataSource() datasource.DataSource {
	return &internetDataSource{}
}

func (d *internetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet"
}

func (d *internetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type internetDataSourceModel struct {
	internetBaseModel
}

func (d *internetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resourceName := "Switch+Router"

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId(resourceName),
			"name":        common.SchemaDataSourceName(resourceName),
			"description": common.SchemaDataSourceDescription(resourceName),
			"tags":        common.SchemaDataSourceTags(resourceName),
			"icon_id":     common.SchemaDataSourceIconID(resourceName),
			"zone":        common.SchemaDataSourceZone(resourceName),
			"switch_id":   common.SchemaDataSourceSwitchID(resourceName),
			"netmask":     common.SchemaDataSourceNetMask(resourceName),
			"gateway":     common.SchemaDataSourceGateway(resourceName),
			"band_width": schema.Int32Attribute{
				Computed:    true,
				Description: "The bandwidth of the network connected to the Internet in Mbps",
			},
			"server_ids": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: desc.Sprintf("A set of the ID of Servers connected to the %s", resourceName),
			},
			"network_address": schema.StringAttribute{
				Computed:    true,
				Description: "The network address assigned to the Switch+Router",
			},
			"min_ip_address": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("Minimum IP address in assigned global addresses to the %s", resourceName),
			},
			"max_ip_address": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("Maximum IP address in assigned global addresses to the %s", resourceName),
			},
			"ip_addresses": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: desc.Sprintf("A set of assigned global address to the %s", resourceName),
			},
			"enable_ipv6": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable IPv6",
			},
			"ipv6_prefix": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The network prefix of assigned IPv6 addresses to the %s", resourceName),
			},
			"ipv6_prefix_len": schema.Int32Attribute{
				Computed:    true,
				Description: desc.Sprintf("The bit length of IPv6 network prefix for %s", resourceName),
			},
			"ipv6_network_address": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The IPv6 network address assigned to the %s", resourceName),
			},
			"assigned_tags": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: desc.Sprintf("The auto assigned tags of the %s when band_width is changed", resourceName),
			},
		},
	}
}

func (d *internetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data internetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewInternetOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud Internet resource: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.Internet) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(ctx, d.client, zone, res.Internet[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
