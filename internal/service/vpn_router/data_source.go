// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package vpn_router

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type vpnRouterDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &vpnRouterDataSource{}
	_ datasource.DataSourceWithConfigure = &vpnRouterDataSource{}
)

func NewVPNRouterDataSource() datasource.DataSource {
	return &vpnRouterDataSource{}
}

func (d *vpnRouterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_router"
}

func (d *vpnRouterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type vpnRouterDataSourceModel struct {
	vpnRouterBaseModel
}

func (d *vpnRouterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("VPNRouter"),
			"name":        common.SchemaDataSourceName("VPNRouter"),
			"description": common.SchemaDataSourceDescription("VPNRouter"),
			"tags":        common.SchemaDataSourceTags("VPNRouter"),
			"zone":        common.SchemaDataSourceZone("VPNRouter"),
			"icon_id":     common.SchemaDataSourceIconID("VPNRouter"),
			"plan":        common.SchemaDataSourcePlan("VPNRouter", iaastypes.VPCRouterPlanStrings),
			"version": schema.Int32Attribute{
				Computed:    true,
				Description: "The version of the VPN Router.",
			},
			"public_network_interface": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "A list of additional network interface setting. This doesn't include primary network interface setting",
				Attributes: map[string]schema.Attribute{
					"switch_id": common.SchemaDataSourceSwitchID("VPNRouter"),
					"vip": schema.StringAttribute{
						Computed:    true,
						Description: "The virtual IP address of the VPN Router. This is only used when `plan` is not `standard`",
					},
					"ip_addresses": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "The list of the IP address assigned to the VPN Router. This will be only one value when `plan` is `standard`, two values otherwise",
					},
					"vrid": schema.Int64Attribute{
						Computed:    true,
						Description: "The Virtual Router Identifier. This is only used when `plan` is not `standard`",
					},
					"aliases": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "A list of ip alias assigned to the VPN Router. This is only used when `plan` is not `standard`",
					},
				},
			},
			"public_ip": schema.StringAttribute{
				Computed:    true,
				Description: "The public ip address of the VPN Router",
			},
			"public_netmask": schema.Int64Attribute{
				Computed:    true,
				Description: "The bit length of the subnet to assign to the public network interface",
			},
			"syslog_host": schema.StringAttribute{
				Computed:    true,
				Description: "The ip address of the syslog host to which the VPN Router sends logs",
			},
			"internet_connection": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable connecting to the Internet from the VPN Router",
			},
			"private_network_interface": schema.ListNestedAttribute{
				Computed:    true,
				Description: "A list of additional network interface setting. This doesn't include primary network interface setting",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int32Attribute{
							Computed:    true,
							Description: "The index of the network interface. This will be between `1`-`7`",
						},
						"switch_id": common.SchemaDataSourceSwitchID("VPNRouter"),
						"vip": schema.StringAttribute{
							Computed:    true,
							Description: "The virtual IP address assigned to the network interface. This is only used when `plan` is not `standard`",
						},
						"ip_addresses": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "A list of ip address assigned to the network interface. This will be only one value when `plan` is `standard`, two values otherwise",
						},
						"netmask": schema.Int32Attribute{
							Computed:    true,
							Description: "The bit length of the subnet assigned to the network interface",
						},
					},
				},
			},
			"dhcp_server": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"interface_index": schema.Int32Attribute{
							Computed:    true,
							Description: "The index of the network interface on which to enable the DHCP service. This will be between `1`-`7`",
						},
						"range_start": schema.StringAttribute{
							Computed:    true,
							Description: "The start value of IP address range to assign to DHCP client",
						},
						"range_stop": schema.StringAttribute{
							Computed:    true,
							Description: "The end value of IP address range to assign to DHCP client",
						},
						"dns_servers": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "A list of IP address of DNS server to assign to DHCP client",
						},
					},
				},
			},
			"dhcp_static_mapping": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Computed:    true,
							Description: "The static IP address to assign to DHCP client",
						},
						"mac_address": schema.StringAttribute{
							Computed:    true,
							Description: "The source MAC address of static mapping",
						},
					},
				},
			},
			"dns_forwarding": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"interface_index": schema.Int32Attribute{
						Computed:    true,
						Description: "The index of the network interface on which to enable the DNS forwarding service",
					},
					"dns_servers": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "A list of IP address of DNS server to forward to",
					},
				},
			},
			"firewall": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"interface_index": schema.Int32Attribute{
							Computed:    true,
							Description: "The index of the network interface on which to enable filtering. This will be between `0`-`7`",
						},
						"direction": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The direction to apply the firewall. This will be one of [%s]", []string{"send", "receive"}),
						},
						"expression": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"protocol": schema.StringAttribute{
										Computed:    true,
										Description: desc.Sprintf("The protocol used for filtering. This will be one of [%s]", iaastypes.VPCRouterFirewallProtocolStrings),
									},
									"source_network": schema.StringAttribute{
										Computed:    true,
										Description: "A source IP address or CIDR block used for filtering (e.g. `192.0.2.1`, `192.0.2.0/24`)",
									},
									"source_port": schema.StringAttribute{
										Computed:    true,
										Description: "A source port number or port range used for filtering (e.g. `1024`, `1024-2048`). This is only used when `protocol` is `tcp` or `udp`",
									},
									"destination_network": schema.StringAttribute{
										Computed:    true,
										Description: "A destination IP address or CIDR block used for filtering (e.g. `192.0.2.1`, `192.0.2.0/24`)",
									},
									"destination_port": schema.StringAttribute{
										Computed:    true,
										Description: "A destination port number or port range used for filtering (e.g. `1024`, `1024-2048`). This is only used when `protocol` is `tcp` or `udp`",
									},
									"allow": schema.BoolAttribute{
										Computed:    true,
										Description: "The flag to allow the packet through the filter",
									},
									"logging": schema.BoolAttribute{
										Computed:    true,
										Description: "The flag to enable packet logging when matching the expression",
									},
									"description": common.SchemaDataSourceDescription("firewall expression"),
								},
							},
						},
					},
				},
			},
			"l2tp": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"pre_shared_secret": schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "The pre shared secret for L2TP/IPsec",
					},
					"range_start": schema.StringAttribute{
						Computed:    true,
						Description: "The start value of IP address range to assign to L2TP/IPsec client",
					},
					"range_stop": schema.StringAttribute{
						Computed:    true,
						Description: "The end value of IP address range to assign to L2TP/IPsec client",
					},
				},
			},
			"port_forwarding": schema.ListNestedAttribute{
				Computed:    true,
				Description: "A list of `port_forwarding` blocks as defined below. This represents a `Reverse NAT`",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The protocol used for port forwarding. This will be one of [%s]", []string{"tcp", "udp"}),
						},
						"public_port": schema.Int32Attribute{
							Computed:    true,
							Description: "The source port number of the port forwarding. This will be a port number on a public network",
						},
						"private_ip": schema.StringAttribute{
							Computed:    true,
							Description: "The destination ip address of the port forwarding",
						},
						"private_port": schema.Int32Attribute{
							Computed:    true,
							Description: "The destination port number of the port forwarding. This will be a port number on a private network",
						},
						"description": common.SchemaDataSourceDescription("port forwarding"),
					},
				},
			},
			"pptp": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"range_start": schema.StringAttribute{
						Computed:    true,
						Description: "The start value of IP address range to assign to PPTP client",
					},
					"range_stop": schema.StringAttribute{
						Computed:    true,
						Description: "The end value of IP address range to assign to PPTP client",
					},
				},
			},
			"wire_guard": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"ip_address": schema.StringAttribute{
						Computed:    true,
						Description: "The IP address for WireGuard server",
					},
					"public_key": schema.StringAttribute{
						Computed:    true,
						Description: "the public key of the WireGuard server",
					},
					"peer": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "the name of the peer",
								},
								"ip_address": schema.StringAttribute{
									Computed:    true,
									Description: "the IP address of the peer",
								},
								"public_key": schema.StringAttribute{
									Computed:    true,
									Description: "the public key of the WireGuard client",
								},
							},
						},
					},
				},
			},
			"site_to_site_vpn": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"peer": schema.StringAttribute{
							Computed:    true,
							Description: "The IP address of the opposing appliance connected to the VPN Router",
						},
						"remote_id": schema.StringAttribute{
							Computed:    true,
							Description: "The id of the opposing appliance connected to the VPN Router. This is typically set same as value of `peer`",
						},
						"pre_shared_secret": schema.StringAttribute{
							Computed:    true,
							Sensitive:   true,
							Description: "The pre shared secret for the VPN",
						},
						"routes": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "A list of CIDR block of VPN connected networks",
						},
						"local_prefix": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "A list of CIDR block of the network under the VPN Router",
						},
					},
				},
			},
			"site_to_site_vpn_parameter": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"ike": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"lifetime": schema.Int64Attribute{Computed: true},
							"dpd": schema.SingleNestedAttribute{
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"interval": schema.Int32Attribute{Computed: true},
									"timeout":  schema.Int32Attribute{Computed: true},
								},
							},
						},
					},
					"esp": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"lifetime": schema.Int64Attribute{Computed: true},
						},
					},
					"encryption_algo": schema.StringAttribute{Computed: true},
					"hash_algo":       schema.StringAttribute{Computed: true},
					"dh_group":        schema.StringAttribute{Computed: true},
				},
			},
			"static_nat": schema.ListNestedAttribute{
				Computed:    true,
				Description: "A list of `static_nat` blocks as defined below. This represents a `1:1 NAT`, doing static mapping to both send/receive to/from the Internet. This is only used when `plan` is not `standard`",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"public_ip": schema.StringAttribute{
							Computed:    true,
							Description: "The public IP address used for the static NAT",
						},
						"private_ip": schema.StringAttribute{
							Computed:    true,
							Description: "The private IP address used for the static NAT",
						},
						"description": common.SchemaDataSourceDescription("static NAT"),
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
			"scheduled_maintenance": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"day_of_week": schema.StringAttribute{
						Computed:    true,
						Description: desc.Sprintf("The value must be in [%s]", iaastypes.DaysOfTheWeekStrings),
					},
					"hour": schema.Int32Attribute{
						Computed:    true,
						Description: "The time to start maintenance",
					},
				},
			},
			"user": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The user name used to authenticate remote access",
						},
						"password": schema.StringAttribute{
							Computed:    true,
							Sensitive:   true,
							Description: "The password used to authenticate remote access",
						},
					},
				},
			},
		},
	}
}

func (d *vpnRouterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vpnRouterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewVPCRouterOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", "could not find SakuraCloud VPCRouter resource: "+err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.VPCRouters) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	vpnRouter := res.VPCRouters[0]
	if _, err := data.updateState(ctx, d.client, zone, vpnRouter); err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not update state for SakuraCloud VPCRouter resource: %s", err))
		return
	}

	data.IconID = types.StringValue(vpnRouter.IconID.String())
	data.SyslogHost = types.StringValue(vpnRouter.Settings.SyslogHost)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
