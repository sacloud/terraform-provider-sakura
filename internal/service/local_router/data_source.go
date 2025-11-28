// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package local_router

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type localRouterDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &localRouterDataSource{}
	_ datasource.DataSourceWithConfigure = &localRouterDataSource{}
)

func NewLocalRouterDataSource() datasource.DataSource {
	return &localRouterDataSource{}
}

func (d *localRouterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_router"
}

func (d *localRouterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type localRouterDataSourceModel struct {
	localRouterBaseModel
}

func (d *localRouterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Local Router"),
			"name":        common.SchemaDataSourceName("Local Router"),
			"description": common.SchemaDataSourceDescription("Local Router"),
			"tags":        common.SchemaDataSourceTags("Local Router"),
			"icon_id":     common.SchemaDataSourceIconID("Local Router"),
			"switch": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"code": schema.StringAttribute{
						Computed:    true,
						Description: "The resource ID of the Switch",
					},
					"category": schema.StringAttribute{
						Computed:    true,
						Description: "The category name of connected services (e.g. `cloud`, `vps`)",
					},
					"zone": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the Zone",
					},
				},
			},
			"network_interface": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"vip": schema.StringAttribute{
						Computed:    true,
						Description: "The virtual IP address",
					},
					"ip_addresses": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "The list of the IP address assigned",
					},
					"netmask": schema.Int32Attribute{
						Computed:    true,
						Description: "The bit length of the subnet assigned to the network interface",
					},
					"vrid": schema.Int64Attribute{
						Computed:    true,
						Description: "The Virtual Router Identifier",
					},
				},
			},
			"peer": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"peer_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the peer LocalRouter",
						},
						"secret_key": schema.StringAttribute{
							Computed:    true,
							Sensitive:   true,
							Description: "The secret key of the peer LocalRouter",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "The flag to enable the LocalRouter",
						},
						"description": common.SchemaDataSourceDescription("Local Router peer"),
					},
				},
			},
			"static_route": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"prefix": schema.StringAttribute{
							Computed:    true,
							Description: "The CIDR block of destination",
						},
						"next_hop": schema.StringAttribute{
							Computed:    true,
							Description: "The IP address of the next hop",
						},
					},
				},
			},
			"secret_keys": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Sensitive:   true,
				Description: "A list of secret key used for peering from other LocalRouters",
			},
		},
		MarkdownDescription: "Get information about an existing Local Router.",
	}
}

func (d *localRouterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data localRouterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewLocalRouterOp(d.client)
	res, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", "failed to find SakuraCloud LocalRouter resource: "+err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.LocalRouters) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.LocalRouters[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
