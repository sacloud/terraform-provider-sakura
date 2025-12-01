// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nosql

import (
	"context"
	"fmt"
	"net/netip"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sacloud/nosql-api-go"
	v1 "github.com/sacloud/nosql-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type nosqlAdditionalNodesResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &nosqlAdditionalNodesResource{}
	_ resource.ResourceWithConfigure   = &nosqlAdditionalNodesResource{}
	_ resource.ResourceWithImportState = &nosqlAdditionalNodesResource{}
)

func NewNosqlAdditionalNodesResource() resource.Resource {
	return &nosqlAdditionalNodesResource{}
}

func (d *nosqlAdditionalNodesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nosql_additional_nodes"
}

func (d *nosqlAdditionalNodesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.NosqlClient
}

type nosqlAdditionalNodesResourceModel struct {
	nosqlBaseModel
	SwitchID types.String   `tfsdk:"switch_id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (d *nosqlAdditionalNodesResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Additional nodes of NoSQL appliance"),
			"name":        common.SchemaResourceName("Additional nodes of NoSQL appliance"),
			"description": common.SchemaResourceDescription("Additional nodes of NoSQL appliance"),
			"tags":        common.SchemaResourceTags("Additional nodes of NoSQL appliance"),
			"zone": schema.StringAttribute{
				Required:    true,
				Description: "Zone where the additional nodes of NoSQL appliance is located.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"switch_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the switch to connect to the Additional nodes of NoSQL appliance.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"settings": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Settings of the Additional nodes of NoSQL appliance",
				Attributes: map[string]schema.Attribute{
					"reserve_ip_address": schema.StringAttribute{ // OpenAPIでは必須ではないが、現状指定しないとエラーになる
						Required:    true,
						Description: "Reserved IP address. This address is used for dead node replacement",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
					"source_network": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "Source network address",
					},
					"backup": schema.SingleNestedAttribute{ // Additional nodesでは設定不可
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"connect": schema.StringAttribute{
								Computed:    true,
								Description: "Backup destination by NFS URL format",
							},
							"days_of_week": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "Backup schedule",
							},
							"time": schema.StringAttribute{
								Computed:    true,
								Description: "Time for backup execution",
							},
							"rotate": schema.Int32Attribute{
								Computed:    true,
								Description: "Number of backup rotations",
							},
						},
					},
					"repair": schema.SingleNestedAttribute{ // Additional nodesでは設定不可
						Computed:    true,
						Description: "Regular repair configuration",
						Attributes: map[string]schema.Attribute{
							"incremental": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Incremental repair configuration",
								Attributes: map[string]schema.Attribute{
									"days_of_week": schema.SetAttribute{
										ElementType: types.StringType,
										Computed:    true,
										Description: "Repair schedule",
										Validators: []validator.Set{
											setvalidator.ValueStringsAre(stringvalidator.OneOf(toStrs(v1.GetNosqlSettingsRepairIncrementalDaysOfWeekItemSun.AllValues())...)),
										},
									},
									"time": schema.StringAttribute{
										Computed:    true,
										Description: "Time for incremental repair execution",
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexp.MustCompile(`^(([0-1][0-9]|2[0-3]):(00|15|30|45))$`), "invalid time format (HH:MM)"),
										},
									},
								},
							},
							"full": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Full repair configuration",
								Attributes: map[string]schema.Attribute{
									"interval": schema.Int32Attribute{
										Computed:    true,
										Description: "Execution interval of 7 days. Supported values are 7 / 14 / 21 / 28",
										Validators: []validator.Int32{
											int32validator.OneOf(7, 14, 21, 28),
										},
									},
									"day_of_week": schema.StringAttribute{
										Computed:    true,
										Description: "Repair schedule",
										Validators: []validator.String{
											stringvalidator.OneOf(toStrs(v1.GetNosqlSettingsRepairFullDayOfWeekSun.AllValues())...),
										},
									},
									"time": schema.StringAttribute{
										Computed:    true,
										Description: "Time for full repair execution",
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexp.MustCompile(`^(([0-1][0-9]|2[0-3]):(00|15|30|45))$`), "invalid time format (HH:MM)"),
										},
									},
								},
							},
						},
					},
				},
			},
			"remark": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"nosql": schema.SingleNestedAttribute{
						Required:    true,
						Description: "NoSQL database information",
						Attributes: map[string]schema.Attribute{
							"primary_nodes": schema.SingleNestedAttribute{
								Required:    true,
								Description: "The primary node information for additional nodes",
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Required:    true,
										Description: "The resource ID of the primary NoSQL appliance",
									},
									"zone": schema.StringAttribute{
										Required:    true,
										Description: "Zone where the primary NoSQL appliance is located.",
									},
								},
							},
							"version": schema.StringAttribute{ // versionがないと400エラーになるため必須にする
								Required:    true,
								Description: "Version of database engine used by NoSQL appliance.",
							},
							// Use top-level zone attribute for resource creation
							"zone": schema.StringAttribute{
								Computed:    true,
								Description: "Zone where the additional nodes of NoSQL appliance is located.",
							},
							// これより下のフィールドはAdditional nodesでは設定不可。
							"port": schema.Int32Attribute{
								Computed:    true,
								Description: "Port number used by NoSQL appliance.",
							},
							"default_user": schema.StringAttribute{
								Computed:    true,
								Description: "Default user for NoSQL appliance.",
							},
							"engine": schema.StringAttribute{
								Computed:    true,
								Description: "Database engine used by NoSQL appliance.",
							},
							"disk_size": schema.Int32Attribute{
								Computed:    true,
								Description: "Disk size of NoSQL appliance.",
							},
							"memory": schema.Int32Attribute{
								Computed:    true,
								Description: "Memory size of NoSQL appliance.",
							},
							"nodes": schema.Int32Attribute{
								Computed:    true,
								Description: "Number of nodes. 3 for primary node, 2 for additional nodes",
							},
							"virtualcore": schema.Int32Attribute{
								Computed:    true,
								Description: "Number of virtual cores used by NoSQL appliance.",
							},
						},
					},
					"servers": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "IP addresses which connect to user's switch",
						Validators: []validator.List{
							listvalidator.ValueStringsAre(sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4)),
							listvalidator.SizeBetween(2, 2),
						},
					},
					"network": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Network information",
						Attributes: map[string]schema.Attribute{
							"gateway": schema.StringAttribute{
								Required:    true,
								Description: "The gateway address of the network",
								Validators: []validator.String{
									sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
								},
							},
							"netmask": schema.Int32Attribute{
								Required:    true,
								Description: "The netmask of the network",
							},
						},
					},
					"zone_id": schema.StringAttribute{
						Computed:    true,
						Description: "Zone ID where NoSQL appliance is located.",
					},
				},
			},
			"instance": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Instance and host information",
				Attributes: map[string]schema.Attribute{
					"status": schema.StringAttribute{
						Computed:    true,
						Description: "The NoSQL instance status. 'up' or 'down'",
					},
					"status_changed_at": schema.StringAttribute{
						Computed:    true,
						Description: "The time when the status was last changed",
					},
					"host": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed:    true,
								Description: "The host name where NoSQL appliance is running",
							},
							"info_url": schema.StringAttribute{
								Computed:    true,
								Description: "The information URL of the host where NoSQL appliance is running",
							},
						},
					},
					"hosts": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "The host name where NoSQL appliance is running",
								},
								"info_url": schema.StringAttribute{
									Computed:    true,
									Description: "The information URL of the host where NoSQL appliance is running",
								},
							},
						},
					},
				},
			},
			"disk": schema.SingleNestedAttribute{ // Additional nodesでは設定不可
				Computed:    true,
				Description: "Disk encryption information",
				Attributes: map[string]schema.Attribute{
					"encryption_key": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Encryption key setting. Specify KMS key ID.",
						Attributes: map[string]schema.Attribute{
							"kms_key_id": schema.StringAttribute{
								Required:    true,
								Description: "KMS key ID for Disk encryption",
							},
						},
					},
					"encryption_algorithm": schema.StringAttribute{
						Computed:    true,
						Description: "Encryption algorithm for Disk encryption",
					},
				},
			},
			"interfaces": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Network interfaces",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Computed:    true,
							Description: "IP Address assigned to the interface",
						},
						"user_ip_address": schema.StringAttribute{
							Computed:    true,
							Description: "IP Address which connect to user's switch",
						},
						"hostname": schema.StringAttribute{
							Computed:    true,
							Description: "Hostname assigned to the interface",
						},
						"switch": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:    true,
									Description: "The ID of the switch connected to the interface",
								},
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "The name of the switch connected to the interface",
								},
								"scope": schema.StringAttribute{
									Computed:    true,
									Description: "The scope of the switch connected to the interface",
								},
								"subnet": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"network_address": schema.StringAttribute{
											Computed:    true,
											Description: "The network address of the subnet connected to the interface",
										},
										"netmask": schema.Int32Attribute{
											Computed:    true,
											Description: "The netmask of the subnet connected to the interface",
										},
										"gateway": schema.StringAttribute{
											Computed:    true,
											Description: "The gateway of the subnet connected to the interface",
										},
										"band_width_mbps": schema.Int32Attribute{
											Computed:    true,
											Description: "The bandwidth in Mbps of the subnet connected to the interface",
										},
									},
								},
								"user_subnet": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"netmask": schema.Int32Attribute{
											Computed:    true,
											Description: "The netmask of the user subnet connected to the interface",
										},
										"gateway": schema.StringAttribute{
											Computed:    true,
											Description: "The gateway of the user subnet connected to the interface",
										},
									},
								},
							},
						},
					},
				},
			},
			"availability": schema.StringAttribute{
				Computed:    true,
				Description: "Availability state. state is one of migrating / available / failed",
			},
			"generation": schema.Int32Attribute{
				Computed:    true,
				Description: "Generation number",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation time",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages additional nodes of NoSQL appliance.",
	}
}

func (r *nosqlAdditionalNodesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *nosqlAdditionalNodesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nosqlAdditionalNodesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	pn := expandNosqlPrimaryNodes(&plan)

	instanceOp := nosql.NewInstanceOpWithZone(r.client, pn.Appliance.ID, pn.Appliance.Zone.Name)
	res, err := instanceOp.AddNodes(ctx, *expandNosqlAddNodesRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create NoSQL additional nodes: %s", err))
		return
	}

	if err := waitNosqlReady(ctx, r.client, res.ID); err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	if err := waitNosqlProcessingDone(ctx, r.client, res.ID, "AddNode"); err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	data, err := nosql.NewDatabaseOp(r.client).Read(ctx, res.ID)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to read NoSQL additional nodes[%s] for state update: %s", res.ID, err))
		return
	}

	plan.updateState(data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *nosqlAdditionalNodesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nosqlAdditionalNodesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := state.ID.ValueString()
	data := getNosql(ctx, r.client, sid, &req.State, &resp.Diagnostics)
	if data == nil {
		return
	}

	state.updateState(data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *nosqlAdditionalNodesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Error", "Updating additional nodes of NoSQL appliance is not supported.")
}

func (r *nosqlAdditionalNodesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nosqlAdditionalNodesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	sid := state.ID.ValueString()
	data := getNosql(ctx, r.client, sid, &req.State, &resp.Diagnostics)
	if data == nil {
		return
	}

	dbOp := nosql.NewDatabaseOp(r.client)
	instanceOp := nosql.NewInstanceOp(r.client, data.ID.Value)
	if data.Instance.Value.Status.Value != "down" {
		if err := instanceOp.Stop(ctx); err != nil {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to stop NoSQL additional nodes[%s]: %s", sid, err))
			return
		}
		if err := waitNosqlDown(ctx, r.client, data.ID.Value); err != nil {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to wait for NoSQL additional nodes[%s] stop: %s", sid, err))
			return
		}
	}

	if err := dbOp.Delete(ctx, data.ID.Value); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete NoSQL additional nodes[%s]: %s", sid, err))
		return
	}
}

func expandNosqlAddNodesRequest(model *nosqlAdditionalNodesResourceModel) *v1.NosqlCreateRequestAppliance {
	servers := common.TlistToStrings(model.Remark.Servers)
	appliance := &v1.NosqlCreateRequestAppliance{
		Name:        model.Name.ValueString(),
		Description: v1.NewOptString(model.Description.ValueString()),
		Tags:        v1.NewOptNilTags(common.TsetToStrings(model.Tags)),
		Settings: v1.NewOptNosqlCreateRequestApplianceSettings(v1.NosqlCreateRequestApplianceSettings{
			ReserveIPAddress: v1.NewOptIPv4(netip.MustParseAddr(model.Settings.ReserveIPAddress.ValueString())),
		}),
		Remark: v1.NosqlRemark{
			Nosql: v1.NosqlRemarkNosql{
				DatabaseVersion: v1.NewOptString(model.Remark.Nosql.Version.ValueString()),
				Nodes:           2,
				Zone:            model.Zone.ValueString(),
			},
			Servers: []v1.NosqlRemarkServersItem{
				{UserIPAddress: netip.MustParseAddr(servers[0])},
				{UserIPAddress: netip.MustParseAddr(servers[1])},
			},
			Network: v1.NosqlRemarkNetwork{
				DefaultRoute:   model.Remark.Network.Gateway.ValueString(),
				NetworkMaskLen: int(model.Remark.Network.Netmask.ValueInt32()),
			},
		},
		UserInterfaces: []v1.NosqlCreateRequestApplianceUserInterfacesItem{
			{
				Switch:         v1.NosqlCreateRequestApplianceUserInterfacesItemSwitch{ID: model.SwitchID.ValueString()},
				UserIPAddress1: netip.MustParseAddr(servers[0]),
				UserIPAddress2: v1.NewOptIPv4(netip.MustParseAddr(servers[1])),
				UserSubnet: v1.NewOptNosqlCreateRequestApplianceUserInterfacesItemUserSubnet(
					v1.NosqlCreateRequestApplianceUserInterfacesItemUserSubnet{
						DefaultRoute:   model.Remark.Network.Gateway.ValueString(),
						NetworkMaskLen: int(model.Remark.Network.Netmask.ValueInt32()),
					}),
			},
		},
	}

	return appliance
}

func expandNosqlPrimaryNodes(model *nosqlAdditionalNodesResourceModel) v1.NosqlRemarkNosqlPrimaryNodes {
	var primaryNodes nosqlRemarkNosqlPrimaryNodesModel
	_ = model.Remark.Nosql.PrimaryNodes.As(context.Background(), &primaryNodes, basetypes.ObjectAsOptions{})

	return v1.NosqlRemarkNosqlPrimaryNodes{
		Appliance: v1.NosqlRemarkNosqlPrimaryNodesAppliance{ID: primaryNodes.ID.ValueString(),
			Zone: v1.NosqlRemarkNosqlPrimaryNodesApplianceZone{Name: primaryNodes.Zone.ValueString()},
		},
	}
}
