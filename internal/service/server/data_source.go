// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package server

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

type serverDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &serverDataSource{}
	_ datasource.DataSourceWithConfigure = &serverDataSource{}
)

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type serverDataSourceModel struct {
	serverBaseModel
}

func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Server"),
			"name":        common.SchemaDataSourceName("Server"),
			"description": common.SchemaDataSourceDescription("Server"),
			"tags":        common.SchemaDataSourceTags("Server"),
			"zone":        common.SchemaDataSourceZone("Server"),
			"icon_id":     common.SchemaDataSourceIconID("Server"),
			"core": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of virtual CPUs",
			},
			"memory": schema.Int64Attribute{
				Computed:    true,
				Description: "The size of memory in GiB",
			},
			"gpu": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of GPUs",
			},
			"gpu_model": schema.StringAttribute{
				Computed:    true,
				Description: "The model of gpu",
			},
			"cpu_model": schema.StringAttribute{
				Computed:    true,
				Description: "The model of cpu",
			},
			"commitment": schema.StringAttribute{
				Computed: true,
				Description: desc.Sprintf(
					"The policy of how to allocate virtual CPUs to the server. This will be one of [%s]",
					iaastypes.CommitmentStrings,
				),
			},
			"disks": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of disk id connected to the server",
			},
			"interface_driver": schema.StringAttribute{
				Computed: true,
				Description: desc.Sprintf(
					"The driver name of network interface. This will be one of [%s]",
					iaastypes.InterfaceDriverStrings,
				),
			},
			"network_interface": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"upstream": schema.StringAttribute{
							Computed: true,
							Description: desc.Sprintf(
								"The upstream type or upstream switch id. This will be one of [%s]",
								[]string{"shared", "disconnect", "<switch id>"},
							),
						},
						"user_ip_address": schema.StringAttribute{
							//CustomType:  iptypes.IPv4AddressType{},
							Computed:    true,
							Description: "The IP address for only display. This value doesn't affect actual NIC settings",
						},
						"packet_filter_id": schema.StringAttribute{
							Computed:    true,
							Description: "The id of the packet filter attached to the network interface",
						},
						"mac_address": schema.StringAttribute{
							Computed:    true,
							Description: "The MAC address",
						},
					},
				},
			},
			"cdrom_id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the CD-ROM attached to the server",
			},
			"private_host_id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the private host which the server is assigned",
			},
			"private_host_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the private host which the server is assigned",
			},
			"ip_address": common.SchemaDataSourceIPAddress("Server"),
			"gateway":    common.SchemaDataSourceGateway("Server"),
			"netmask":    common.SchemaDataSourceNetMask("Server"),
			"network_address": schema.StringAttribute{
				Computed:    true,
				Description: "The network address which the `ip_address` belongs",
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the Server",
			},
			"dns_servers": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of IP address of DNS server in the zone",
			},
		},
		MarkdownDescription: "Get information about an existing Server.",
	}
}

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewServerOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", "failed to find SakuraCloud Server: "+err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.Servers) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.Servers[0], zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
