// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package vpn_router

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/power"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

const vpnRouterWaitAfterCreateDuration = 2 * time.Minute

type vpnRouterResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &vpnRouterResource{}
	_ resource.ResourceWithConfigure   = &vpnRouterResource{}
	_ resource.ResourceWithImportState = &vpnRouterResource{}
)

func NewVPNRouterResource() resource.Resource {
	return &vpnRouterResource{}
}

func (d *vpnRouterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_router"
}

func (d *vpnRouterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type vpnRouterResourceModel struct {
	vpnRouterBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (d *vpnRouterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("VPNRouter"),
			"name":        common.SchemaResourceName("VPNRouter"),
			"description": common.SchemaResourceDescription("VPNRouter"),
			"tags":        common.SchemaResourceTags("VPNRouter"),
			"zone":        common.SchemaResourceZone("VPNRouter"),
			"icon_id":     common.SchemaResourceIconID("VPNRouter"),
			"plan":        common.SchemaResourcePlan("VPNRouter", "standard", iaastypes.VPCRouterPlanStrings),
			"version": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(2),
				Description: "The version of the VPN Router.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"public_network_interface": schema.SingleNestedAttribute{
				Optional: true, // only required when `plan` is not `standard`
				Attributes: map[string]schema.Attribute{
					"switch_id": schema.StringAttribute{
						Optional:    true,
						Description: "The id of the switch to connect. This is only required when when `plan` is not `standard`",
						Validators: []validator.String{
							sacloudvalidator.SakuraIDValidator(),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"vip": schema.StringAttribute{
						Optional:    true,
						Description: "The virtual IP address of the VPC Router. This is only required when `plan` is not `standard`",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"ip_addresses": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "The list of the IP address to assign to the VPC Router. This is required only one value when `plan` is `standard`, two values otherwise",
						Validators: []validator.List{
							listvalidator.SizeAtLeast(2),
							listvalidator.SizeAtMost(2),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"vrid": schema.Int64Attribute{
						Optional:    true,
						Description: "The Virtual Router Identifier. This is only required when `plan` is not `standard`",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"aliases": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "A list of ip alias to assign to the VPC Router. This can only be specified if `plan` is not `standard`",
						Validators: []validator.List{
							listvalidator.SizeAtMost(19),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplaceIfConfigured(),
						},
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
				Optional:    true,
				Description: "The ip address of the syslog host to which the VPC Router sends logs",
			},
			"internet_connection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "The flag to enable connecting to the Internet from the VPN Router",
			},
			"private_network_interface": schema.ListNestedAttribute{
				Optional:    true,
				Description: "A list of additional network interface setting. This doesn't include primary network interface setting",
				Validators: []validator.List{
					listvalidator.SizeAtMost(7),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int32Attribute{
							Required:    true,
							Description: desc.Sprintf("The index of the network interface. %s", desc.Range(1, 7)),
							Validators: []validator.Int32{
								int32validator.Between(1, 7),
							},
						},
						"switch_id": schema.StringAttribute{
							Required:    true,
							Description: "The id of the connected switch",
							Validators: []validator.String{
								sacloudvalidator.SakuraIDValidator(),
							},
						},
						"vip": schema.StringAttribute{
							Optional:    true,
							Description: "The virtual IP address to assign to the network interface. This is only required when `plan` is not `standard`",
						},
						"ip_addresses": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "A set of ip address to assign to the network interface. This is required only one value when `plan` is `standard`, two values otherwise",
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(2),
							},
						},
						"netmask": schema.Int32Attribute{
							Required:    true,
							Description: "The bit length of the subnet assigned to the network interface",
							Validators: []validator.Int32{
								int32validator.Between(16, 29),
							},
						},
					},
				},
			},
			"dhcp_server": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"interface_index": schema.Int32Attribute{
							Required:    true,
							Description: desc.Sprintf("The index of the network interface on which to enable the DHCP service. %s", desc.Range(1, 7)),
							Validators: []validator.Int32{
								int32validator.Between(1, 7),
							},
						},
						"range_start": schema.StringAttribute{
							Required:    true,
							Description: "The start value of IP address range to assign to DHCP client",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
						"range_stop": schema.StringAttribute{
							Required:    true,
							Description: "The end value of IP address range to assign to DHCP client",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
						"dns_servers": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "A list of IP address of DNS server to assign to DHCP client",
						},
					},
				},
			},
			"dhcp_static_mapping": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Required:    true,
							Description: "The static IP address to assign to DHCP client",
						},
						"mac_address": schema.StringAttribute{
							Required:    true,
							Description: "The source MAC address of static mapping",
						},
					},
				},
			},
			"dns_forwarding": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"interface_index": schema.Int32Attribute{
						Required:    true,
						Description: desc.Sprintf("The index of the network interface on which to enable the DNS forwarding service. %s", desc.Range(1, 7)),
						Validators: []validator.Int32{
							int32validator.Between(1, 7),
						},
					},
					"dns_servers": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "A list of IP address of DNS server to forward to",
					},
				},
			},
			"firewall": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"interface_index": schema.Int32Attribute{
							Optional:    true,
							Computed:    true,
							Description: desc.Sprintf("The index of the network interface on which to enable filtering. %s", desc.Range(0, 7)),
							Validators: []validator.Int32{
								int32validator.Between(0, 7),
							},
						},
						"direction": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The direction to apply the firewall. This must be one of [%s]", []string{"send", "receive"}),
							Validators: []validator.String{
								stringvalidator.OneOf("send", "receive"),
							},
						},
						"expression": schema.ListNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"protocol": schema.StringAttribute{
										Required:    true,
										Description: desc.Sprintf("The protocol used for filtering. This must be one of [%s]", iaastypes.VPCRouterFirewallProtocolStrings),
										Validators: []validator.String{
											stringvalidator.OneOf(iaastypes.VPCRouterFirewallProtocolStrings...),
										},
									},
									"source_network": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: "A source IP address or CIDR block used for filtering (e.g. `192.0.2.1`, `192.0.2.0/24`)",
									},
									"source_port": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: "A source port number or port range used for filtering (e.g. `1024`, `1024-2048`). This is only used when `protocol` is `tcp` or `udp`",
									},
									"destination_network": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: "A destination IP address or CIDR block used for filtering (e.g. `192.0.2.1`, `192.0.2.0/24`)",
									},
									"destination_port": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: "A destination port number or port range used for filtering (e.g. `1024`, `1024-2048`). This is only used when `protocol` is `tcp` or `udp`",
									},
									"allow": schema.BoolAttribute{
										Required:    true,
										Description: "The flag to allow the packet through the filter",
									},
									"logging": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Description: "The flag to enable packet logging when matching the expression",
									},
									"description": common.SchemaResourceDescription("firewall expression"),
								},
							},
						},
					},
				},
			},
			"l2tp": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"pre_shared_secret": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "The pre shared secret for L2TP/IPsec",
						Validators: []validator.String{
							stringvalidator.LengthBetween(0, 40),
						},
					},
					"range_start": schema.StringAttribute{
						Required:    true,
						Description: "The start value of IP address range to assign to L2TP/IPsec client",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
					"range_stop": schema.StringAttribute{
						Required:    true,
						Description: "The end value of IP address range to assign to L2TP/IPsec client",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
				},
			},
			"port_forwarding": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The protocol used for port forwarding. This must be one of [%s]", []string{"tcp", "udp"}),
							Validators: []validator.String{
								stringvalidator.OneOf("tcp", "udp"),
							},
						},
						"public_port": schema.Int32Attribute{
							Required:    true,
							Description: "The source port number of the port forwarding. This must be a port number on a public network",
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
							},
						},
						"private_ip": schema.StringAttribute{
							Required:    true,
							Description: "The destination ip address of the port forwarding",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
						"private_port": schema.Int32Attribute{
							Required:    true,
							Description: "The destination port number of the port forwarding. This will be a port number on a private network",
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
							},
						},
						"description": common.SchemaResourceDescription("port forwarding"),
					},
				},
			},
			"pptp": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"range_start": schema.StringAttribute{
						Required:    true,
						Description: "The start value of IP address range to assign to PPTP client",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
					"range_stop": schema.StringAttribute{
						Required:    true,
						Description: "The end value of IP address range to assign to PPTP client",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
				},
			},
			"wire_guard": schema.SingleNestedAttribute{ // TODO: Use SingleNestedAttribute
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"ip_address": schema.StringAttribute{
						Required:    true,
						Description: "The IP address for WireGuard server. This must be formatted with xxx.xxx.xxx.xxx/nn",
					},
					"public_key": schema.StringAttribute{
						Computed:    true,
						Description: "the public key of the WireGuard server",
					},
					"peer": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "the name of the peer",
								},
								"ip_address": schema.StringAttribute{
									Required:    true,
									Description: "the IP address of the peer",
								},
								"public_key": schema.StringAttribute{
									Required:    true,
									Description: "the public key of the WireGuard client",
								},
							},
						},
					},
				},
			},
			"site_to_site_vpn": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"peer": schema.StringAttribute{
							Required:    true,
							Description: "The IP address of the opposing appliance connected to the VPN Router",
						},
						"remote_id": schema.StringAttribute{
							Required:    true,
							Description: "The id of the opposing appliance connected to the VPN Router. This is typically set same as value of `peer`",
						},
						"pre_shared_secret": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: desc.Sprintf("The pre shared secret for the VPN. %s", desc.Length(0, 40)),
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 40),
							},
						},
						"routes": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "A list of CIDR block of VPN connected networks",
						},
						"local_prefix": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "A list of CIDR block of the network under the VPN Router",
						},
					},
				},
			},
			"site_to_site_vpn_parameter": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"ike": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"lifetime": schema.Int64Attribute{
								Optional:    true,
								Computed:    true,
								Description: "Lifetime of IKE SA. Default: 28800",
							},
							"dpd": schema.SingleNestedAttribute{
								Optional: true,
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"interval": schema.Int32Attribute{
										Optional:    true,
										Computed:    true,
										Description: "Default: 15",
									},
									"timeout": schema.Int32Attribute{
										Optional:    true,
										Computed:    true,
										Description: "Default: 30",
									},
								},
							},
						},
					},
					"esp": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"lifetime": schema.Int64Attribute{
								Optional:    true,
								Computed:    true,
								Description: "Default: 1800",
							},
						},
					},
					"encryption_algo": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: desc.Sprintf("This must be one of [%s]", iaastypes.VPCRouterSiteToSiteVPNEncryptionAlgos),
						Validators: []validator.String{
							stringvalidator.OneOf(iaastypes.VPCRouterSiteToSiteVPNEncryptionAlgos...),
						},
					},
					"hash_algo": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: desc.Sprintf("This must be one of [%s]", iaastypes.VPCRouterSiteToSiteVPNHashAlgos),
						Validators: []validator.String{
							stringvalidator.OneOf(iaastypes.VPCRouterSiteToSiteVPNHashAlgos...),
						},
					},
					"dh_group": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: desc.Sprintf("This must be one of [%s]", iaastypes.VPCRouterSiteToSiteVPNDHGroups),
						Validators: []validator.String{
							stringvalidator.OneOf(iaastypes.VPCRouterSiteToSiteVPNDHGroups...),
						},
					},
				},
			},
			"static_nat": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"public_ip": schema.StringAttribute{
							Required:    true,
							Description: "The public IP address used for the static NAT",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
						"private_ip": schema.StringAttribute{
							Required:    true,
							Description: "The private IP address used for the static NAT",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
						"description": common.SchemaResourceDescription("static NAT"),
					},
				},
			},
			"static_route": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"prefix": schema.StringAttribute{
							Required:    true,
							Description: "The CIDR block of destination",
						},
						"next_hop": schema.StringAttribute{
							Required:    true,
							Description: "The IP address of the next hop",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
					},
				},
			},
			"scheduled_maintenance": schema.SingleNestedAttribute{ // TODO: Use SingleNestedAttribute
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"day_of_week": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(iaastypes.DaysOfTheWeek.Monday.String()),
						Description: desc.Sprintf("The value must be in [%s]", iaastypes.DaysOfTheWeekStrings),
						Validators: []validator.String{
							stringvalidator.OneOf(iaastypes.DaysOfTheWeekStrings...),
						},
					},
					"hour": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(3),
						Description: "The time to start maintenance",
						Validators: []validator.Int32{
							int32validator.Between(0, 23),
						},
					},
				},
			},
			"user": schema.ListNestedAttribute{
				Optional: true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The user name used to authenticate remote access",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 20),
							},
						},
						"password": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The password used to authenticate remote access",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 20),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *vpnRouterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *vpnRouterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpnRouterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	builder := expandVPNRouterBuilder(&plan, r.client, zone)
	if err := builder.Validate(ctx, zone); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("validating parameter for SakuraCloud VPCRouter is failed: %s", err))
		return
	}

	vpnRouter, err := builder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud VPCRouter is failed: %s", err))
		return
	}

	// Note: 起動してからしばらくは/:id/Statusが空となるため、数秒待つようにする。
	time.Sleep(vpnRouterWaitAfterCreateDuration)

	if rmResource, err := plan.updateState(ctx, r.client, zone, vpnRouter); err != nil {
		if rmResource {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("could not update state for SakuraCloud VPCRouter[%s] resource: %s", vpnRouter.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vpnRouterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpnRouterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := state.ID.ValueString()
	vpnRouter := getRouter(ctx, r.client, zone, common.SakuraCloudID(sid), &req.State, &resp.Diagnostics)

	if rmResource, err := state.updateState(ctx, r.client, zone, vpnRouter); err != nil {
		if rmResource {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not update state for SakuraCloud VPCRouter[%s] resource: %s", sid, err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vpnRouterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vpnRouterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vrOp := iaas.NewVPCRouterOp(r.client)
	sid := plan.ID.ValueString()

	common.SakuraMutexKV.Lock(sid)
	defer common.SakuraMutexKV.Unlock(sid)

	builder := expandVPNRouterBuilder(&plan, r.client, zone)
	if err := builder.Validate(ctx, zone); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("validating parameter for SakuraCloud VPCRouter[%s] is failed: %s", sid, err))
		return
	}
	builder.ID = common.SakuraCloudID(sid)

	_, err := builder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud VPCRouter[%s] is failed: %s", sid, err))
		return
	}

	// Note: 起動してからしばらくは/:id/Statusが空となるため、数秒待つようにする。
	time.Sleep(vpnRouterWaitAfterCreateDuration)

	vpnRouter, err := vrOp.Read(ctx, zone, common.SakuraCloudID(sid))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("could not read SakuraCloud VPCRouter[%s]: %s", sid, err))
		return
	}

	if rmResource, err := plan.updateState(ctx, r.client, zone, vpnRouter); err != nil {
		if rmResource {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("could not update state for SakuraCloud VPCRouter[%s] resource: %s", sid, err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vpnRouterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpnRouterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vrOp := iaas.NewVPCRouterOp(r.client)
	sid := state.ID.ValueString()

	common.SakuraMutexKV.Lock(sid)
	defer common.SakuraMutexKV.Unlock(sid)

	vpnRouter := getRouter(ctx, r.client, zone, common.SakuraCloudID(sid), &req.State, &resp.Diagnostics)
	if vpnRouter == nil {
		return
	}

	if vpnRouter.InstanceStatus.IsUp() {
		if err := power.ShutdownVPCRouter(ctx, vrOp, zone, vpnRouter.ID, true); err != nil {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("stopping VPCRouter[%s] is failed: %s", sid, err))
			return
		}
	}

	if err := vrOp.Delete(ctx, zone, vpnRouter.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud VPCRouter[%s] is failed: %s", sid, err))
		return
	}
}

func getRouter(ctx context.Context, client *common.APIClient, zone string, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.VPCRouter {
	vrOp := iaas.NewVPCRouterOp(client)
	vpnRouter, err := vrOp.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get VPNRouter Error", fmt.Sprintf("could not read SakuraCloud VPCRouter[%s]: %s", id, err))
		return nil
	}
	return vpnRouter
}
