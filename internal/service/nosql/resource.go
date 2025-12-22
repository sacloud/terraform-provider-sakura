// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nosql

import (
	"context"
	"fmt"
	"net/netip"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/nosql-api-go"
	v1 "github.com/sacloud/nosql-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type nosqlResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &nosqlResource{}
	_ resource.ResourceWithConfigure   = &nosqlResource{}
	_ resource.ResourceWithImportState = &nosqlResource{}
)

func NewNosqlResource() resource.Resource {
	return &nosqlResource{}
}

func (d *nosqlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nosql"
}

func (d *nosqlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.NosqlClient
}

type nosqlResourceModel struct {
	nosqlBaseModel
	Password   types.String   `tfsdk:"password"`
	VSwitchID  types.String   `tfsdk:"vswitch_id"`
	Parameters types.Map      `tfsdk:"parameters"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

func (d *nosqlResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("NoSQL appliance"),
			"name":        common.SchemaResourceName("NoSQL appliance"),
			"description": common.SchemaResourceDescription("NoSQL appliance"),
			"tags":        common.SchemaResourceTags("NoSQL appliance"),
			"plan":        common.SchemaResourcePlan("NoSQL appliance", "100GB", nosql.Plan100GB.AllValuesAsString()),
			"zone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("tk1b"),
				Description: "Zone where NoSQL appliance is located.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password for NoSQL appliance",
				Validators: []validator.String{
					stringvalidator.LengthBetween(12, 30),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9-._]+$`), "only alphanumeric characters and - . _ are allowed"),
				},
			},
			"vswitch_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the vSwitch to connect to the NoSQL appliance.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"parameters": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Parameters for the NoSQL appliance",
			},
			"settings": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Settings of the NoSQL appliance",
				Attributes: map[string]schema.Attribute{
					"source_network": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Description: "Source network address",
					},
					"reserve_ip_address": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Reserved IP address. This address is used for dead node replacement",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
					"backup": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"connect": schema.StringAttribute{
								Required:    true,
								Description: "Backup destination by NFS URL format",
								Validators: []validator.String{
									stringvalidator.RegexMatches(regexp.MustCompile(`^(nfs://[0-9\.]+/[A-Za-z0-9_\-\/]+)$`), "must be in NFS URL format (e.g. nfs://"),
								},
							},
							"days_of_week": schema.SetAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
								Description: "Backup schedule",
								Validators: []validator.Set{
									setvalidator.ValueStringsAre(stringvalidator.OneOf(toStrs(v1.GetNosqlSettingsBackupDayOfWeekItemSun.AllValues())...)),
								},
							},
							"time": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Time for backup execution",
								Validators: []validator.String{
									sacloudvalidator.BackupTimeValidator(),
								},
							},
							"rotate": schema.Int32Attribute{
								Required:    true,
								Description: "Number of backup rotations",
								Validators: []validator.Int32{
									int32validator.Between(1, 8),
								},
							},
						},
					},
					"repair": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Regular repair configuration",
						Attributes: map[string]schema.Attribute{
							"incremental": schema.SingleNestedAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Incremental repair configuration",
								Attributes: map[string]schema.Attribute{
									"days_of_week": schema.SetAttribute{
										ElementType: types.StringType,
										Required:    true,
										Description: "Repair schedule",
										Validators: []validator.Set{
											setvalidator.ValueStringsAre(stringvalidator.OneOf(toStrs(v1.GetNosqlSettingsRepairIncrementalDaysOfWeekItemSun.AllValues())...)),
										},
									},
									"time": schema.StringAttribute{
										Required:    true,
										Description: "Time for incremental repair execution",
										Validators: []validator.String{
											sacloudvalidator.BackupTimeValidator(),
										},
									},
								},
							},
							"full": schema.SingleNestedAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Full repair configuration",
								Attributes: map[string]schema.Attribute{
									"interval": schema.Int32Attribute{
										Required:    true,
										Description: "Execution interval of 7 days. Supported values are 7 / 14 / 21 / 28",
										Validators: []validator.Int32{
											int32validator.OneOf(7, 14, 21, 28),
										},
									},
									"day_of_week": schema.StringAttribute{
										Required:    true,
										Description: "Repair schedule",
										Validators: []validator.String{
											stringvalidator.OneOf(toStrs(v1.GetNosqlSettingsRepairFullDayOfWeekSun.AllValues())...),
										},
									},
									"time": schema.StringAttribute{
										Required:    true,
										Description: "Time for full repair execution",
										Validators: []validator.String{
											sacloudvalidator.BackupTimeValidator(),
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
							// Use top-level zone attribute for resource creation
							"zone": schema.StringAttribute{
								Computed:    true,
								Description: "Zone where NoSQL appliance is located.",
							},
							"port": schema.Int32Attribute{
								Optional:    true,
								Computed:    true,
								Default:     int32default.StaticInt32(9042),
								Description: "Port number used by NoSQL appliance.",
								Validators: []validator.Int32{
									int32validator.Between(1024, 65535),
								},
							},
							"default_user": schema.StringAttribute{
								Required:    true,
								Description: "Default user for NoSQL appliance.",
								Validators: []validator.String{
									stringvalidator.LengthBetween(4, 20),
									stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z][a-z0-9_]{3,19}$`), "invalid user name"),
								},
							},
							"version": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("4.1.9"),
								Description: "Version of database engine used by NoSQL appliance.",
								Validators: []validator.String{
									stringvalidator.RegexMatches(regexp.MustCompile(`^\d+\.\d+\.\d+$`), "invalid database version"),
								},
							},
							// これより下のフィールドは現状固定値。将来的には指定できるようになる可能性はある
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
							"primary_nodes": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "This is for additional nodes resource. Always null in this resource",
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:    true,
										Description: "The resource ID of the primary NoSQL appliance",
									},
									"zone": schema.StringAttribute{
										Computed:    true,
										Description: "Zone where the primary NoSQL appliance is located.",
									},
								},
							},
						},
					},
					"servers": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "IP addresses which connect to user's switch",
						Validators: []validator.List{
							listvalidator.ValueStringsAre(sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4)),
							listvalidator.SizeBetween(1, 3),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
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
			"disk": common.SchemaResourceEncryptionDisk("NoSQL appliance"),
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
						"vswitch": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:    true,
									Description: "The ID of the vSwitch connected to the interface",
								},
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "The name of the vSwitch connected to the interface",
								},
								"scope": schema.StringAttribute{
									Computed:    true,
									Description: "The scope of the vSwitch connected to the interface",
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
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a NoSQL appliance.",
	}
}

func (r *nosqlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *nosqlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nosqlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	dbOp := nosql.NewDatabaseOp(r.client)
	res, err := dbOp.Create(ctx, nosql.Plan(plan.Plan.ValueString()), *expandNosqlCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create NoSQL: %s", err))
		return
	}

	id := res.ID.Value
	if err := waitNosqlReady(ctx, r.client, id); err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	if len(plan.Parameters.Elements()) > 0 {
		if err := updateParameters(ctx, r.client, id, plan.Zone.ValueString(), &plan); err != nil {
			resp.Diagnostics.AddWarning("Create Warning", fmt.Sprintf("failed to update parameters for NoSQL[%s]. Update via control panel: %s", id, err))
			return
		} else {
			if err := waitNosqlProcessingDone(ctx, r.client, id, "SetParameter"); err != nil {
				resp.Diagnostics.AddError("Create Error", err.Error())
				return
			}
		}
	}

	data, err := dbOp.Read(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to read NoSQL[%s] for state update: %s", id, err))
		return
	}

	plan.updateState(data)
	if plan.Parameters.IsNull() || plan.Parameters.IsUnknown() {
		plan.Parameters = types.MapNull(types.StringType)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	time.Sleep(10 * time.Second)
}

func (r *nosqlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nosqlResourceModel
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

func (r *nosqlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state nosqlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	sid := plan.ID.ValueString()
	dbOp := nosql.NewDatabaseOp(r.client)
	if err := dbOp.Update(ctx, sid, *expandNosqlUpdateRequest(&plan, sid)); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update NoSQL[%s]: %s", sid, err))
		return
	}

	if err := dbOp.ApplyChanges(ctx, sid); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to apply changes to NoSQL[%s]: %s", sid, err))
		return
	}

	if err := waitNosqlProcessingDone(ctx, r.client, sid, "Update"); err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}

	if len(plan.Parameters.Elements()) > 0 && !plan.Parameters.Equal(state.Parameters) {
		if err := updateParameters(ctx, r.client, sid, plan.Zone.ValueString(), &plan); err != nil {
			resp.Diagnostics.AddWarning("Update Warning", fmt.Sprintf("failed to update parameters for NoSQL[%s]. Update via control panel: %s", sid, err))
		} else {
			if err := waitNosqlProcessingDone(ctx, r.client, sid, "SetParameter"); err != nil {
				resp.Diagnostics.AddError("Update Error", err.Error())
				return
			}
		}
	}

	data := getNosql(ctx, r.client, sid, &req.State, &resp.Diagnostics)
	if data == nil {
		return
	}

	plan.updateState(data)
	if plan.Parameters.IsNull() || plan.Parameters.IsUnknown() {
		plan.Parameters = types.MapNull(types.StringType)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	time.Sleep(10 * time.Second)
}

func (r *nosqlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nosqlResourceModel
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
	instanceOp := nosql.NewInstanceOp(r.client, sid, state.Zone.ValueString())
	if data.Instance.Value.Status.Value != "down" {
		if err := instanceOp.Stop(ctx); err != nil {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to stop NoSQL[%s]: %s", sid, err))
			return
		}
		if err := waitNosqlDown(ctx, r.client, data.ID.Value); err != nil {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to wait for NoSQL[%s] stop: %s", sid, err))
			return
		}
	}

	if err := dbOp.Delete(ctx, data.ID.Value); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete NoSQL[%s]: %s", sid, err))
		return
	}
}

func getNosql(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.GetNosqlAppliance {
	dbOp := nosql.NewDatabaseOp(client)
	res, err := dbOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			if state != nil {
				state.RemoveResource(ctx)
			}
			return nil
		}
		diags.AddError("Get NoSQL Error", fmt.Sprintf("failed to read NoSQL[%s]: %s", id, err))
		return nil
	}
	return res
}

func updateParameters(ctx context.Context, client *v1.Client, id, zone string, model *nosqlResourceModel) error {
	params := common.TmapToStrMap(model.Parameters)
	iOp := nosql.NewInstanceOp(client, id, zone)
	nosqlParameters, err := iOp.GetParameters(ctx)
	if err != nil {
		return err
	}

	settings := make([]v1.NosqlPutParameter, 0, len(params))
	for _, nosqlParam := range nosqlParameters {
		for key, value := range params {
			if nosqlParam.SettingItem == key {
				settings = append(settings, v1.NosqlPutParameter{
					SettingItemId: nosqlParam.SettingItemId,
					SettingValue:  value,
				})
			}
		}
	}

	err = iOp.SetParameters(ctx, settings)
	if err != nil {
		return err
	}

	return nil
}

func expandNosqlCreateRequest(model *nosqlResourceModel) *v1.NosqlCreateRequestAppliance {
	var remarkNosql nosqlRemarkNosqlModel
	_ = model.Remark.Nosql.As(context.Background(), &remarkNosql, basetypes.ObjectAsOptions{})
	appliance := &v1.NosqlCreateRequestAppliance{
		Name:        model.Name.ValueString(),
		Description: v1.NewOptString(model.Description.ValueString()),
		Tags:        v1.NewOptNilTags(common.TsetToStrings(model.Tags)),
		Settings: v1.NosqlSettings{
			SourceNetwork: common.TlistToStrings(model.Settings.SourceNetwork),
			Password:      v1.NewOptPassword(v1.Password(model.Password.ValueString())),
		},
		Remark: v1.NosqlRemark{
			Nosql: v1.NosqlRemarkNosql{
				DatabaseEngine:  v1.NewOptNilNosqlRemarkNosqlDatabaseEngine("Cassandra"),
				DatabaseVersion: v1.NewOptNilString(remarkNosql.Version.ValueString()),
				DefaultUser:     v1.NewOptNilString(remarkNosql.DefaultUser.ValueString()),
				Port:            v1.NewOptNilInt(int(remarkNosql.Port.ValueInt32())),
				Storage:         v1.NewOptNilNosqlRemarkNosqlStorage("SSD"),
				Zone:            model.Zone.ValueString(),
			},
			Network: v1.NosqlRemarkNetwork{
				DefaultRoute:   model.Remark.Network.Gateway.ValueString(),
				NetworkMaskLen: int(model.Remark.Network.Netmask.ValueInt32()),
			},
		},
		UserInterfaces: []v1.NosqlCreateRequestApplianceUserInterfacesItem{
			{
				Switch: v1.NosqlCreateRequestApplianceUserInterfacesItemSwitch{ID: model.VSwitchID.ValueString()},
				UserSubnet: v1.NosqlCreateRequestApplianceUserInterfacesItemUserSubnet{
					DefaultRoute:   model.Remark.Network.Gateway.ValueString(),
					NetworkMaskLen: int(model.Remark.Network.Netmask.ValueInt32()),
				},
			},
		},
	}

	setupServers(model, appliance)

	if !model.Settings.Backup.IsNull() && !model.Settings.Backup.IsUnknown() {
		appliance.Settings.Backup = expandNosqlBackup(model)
	}
	if !model.Settings.Repair.IsNull() && !model.Settings.Repair.IsUnknown() {
		appliance.Settings.Repair = expandNosqlRepair(model)
	}

	if !model.Disk.IsNull() && !model.Disk.IsUnknown() {
		var disk common.SakuraEncryptionDiskModel
		_ = model.Disk.As(context.Background(), &disk, basetypes.ObjectAsOptions{})
		if disk.EncryptionAlgorithm.ValueString() != iaastypes.DiskEncryptionAlgorithms.None.String() {
			appliance.Disk = v1.NewOptNilNosqlCreateRequestApplianceDisk(v1.NosqlCreateRequestApplianceDisk{
				EncryptionAlgorithm: v1.NewOptString(disk.EncryptionAlgorithm.ValueString()),
				EncryptionKey: v1.NewOptNilNosqlCreateRequestApplianceDiskEncryptionKey(
					v1.NosqlCreateRequestApplianceDiskEncryptionKey{KMSKeyID: v1.NewOptNilString(disk.KMSKeyID.ValueString())}),
			})
		}
	}

	return appliance
}

func setupServers(model *nosqlResourceModel, req *v1.NosqlCreateRequestAppliance) {
	plan := nosql.Plan(model.Plan.ValueString())
	servers := common.TlistToStrings(model.Remark.Servers)

	switch plan {
	case nosql.Plan40GB:
		req.Remark.Servers = []v1.NosqlRemarkServersItem{
			{UserIPAddress: netip.MustParseAddr(servers[0])},
		}
		req.UserInterfaces[0].UserIPAddress1 = netip.MustParseAddr(servers[0])
	case nosql.Plan100GB, nosql.Plan250GB:
		req.Settings.ReserveIPAddress = v1.NewOptIPv4(netip.MustParseAddr(model.Settings.ReserveIPAddress.ValueString()))
		req.Remark.Servers = []v1.NosqlRemarkServersItem{
			{UserIPAddress: netip.MustParseAddr(servers[0])},
			{UserIPAddress: netip.MustParseAddr(servers[1])},
			{UserIPAddress: netip.MustParseAddr(servers[2])},
		}
		req.UserInterfaces[0].UserIPAddress1 = netip.MustParseAddr(servers[0])
		req.UserInterfaces[0].UserIPAddress2 = v1.NewOptIPv4(netip.MustParseAddr(servers[1]))
		req.UserInterfaces[0].UserIPAddress3 = v1.NewOptIPv4(netip.MustParseAddr(servers[2]))
	}
}

func expandNosqlUpdateRequest(plan *nosqlResourceModel, id string) *v1.NosqlUpdateRequestAppliance {
	appliance := &v1.NosqlUpdateRequestAppliance{
		ID:          id,
		Name:        v1.NewOptString(plan.Name.ValueString()),
		Description: v1.NewOptString(plan.Description.ValueString()),
		Tags:        v1.NewOptNilTags(common.TsetToStrings(plan.Tags)),
		Settings: v1.NosqlSettings{
			SourceNetwork:    common.TlistToStrings(plan.Settings.SourceNetwork),
			Password:         v1.NewOptPassword(v1.Password(plan.Password.ValueString())),
			ReserveIPAddress: v1.NewOptIPv4(netip.MustParseAddr(plan.Settings.ReserveIPAddress.ValueString())),
		},
	}

	if !plan.Settings.Backup.IsNull() && !plan.Settings.Backup.IsUnknown() {
		appliance.Settings.Backup = expandNosqlBackup(plan)
	}
	if !plan.Settings.Repair.IsNull() && !plan.Settings.Repair.IsUnknown() {
		appliance.Settings.Repair = expandNosqlRepair(plan)
	}

	return appliance
}

func expandNosqlBackup(model *nosqlResourceModel) v1.OptNilNosqlSettingsBackup {
	var backup nosqlBackupModel
	_ = model.Settings.Backup.As(context.Background(), &backup, basetypes.ObjectAsOptions{})
	res := v1.NewOptNilNosqlSettingsBackup(v1.NosqlSettingsBackup{
		Connect: backup.Connect.ValueString(),
	})
	if !backup.DaysOfWeek.IsNull() && !backup.DaysOfWeek.IsUnknown() {
		values := common.TsetToStrings(backup.DaysOfWeek)
		settings := make([]v1.NosqlSettingsBackupDayOfWeekItem, len(values))
		for i, v := range values {
			settings[i] = v1.NosqlSettingsBackupDayOfWeekItem(v)
		}
		res.Value.DayOfWeek = v1.NewOptNilNosqlSettingsBackupDayOfWeekItemArray(settings)
	}
	if !backup.Time.IsNull() && !backup.Time.IsUnknown() {
		res.Value.Time = v1.NewOptNilString(backup.Time.ValueString())
	}
	res.Value.Rotate = int(backup.Rotate.ValueInt32())

	return res
}

func expandNosqlRepair(model *nosqlResourceModel) v1.OptNilNosqlSettingsRepair {
	var repair nosqlRepairModel
	_ = model.Settings.Repair.As(context.Background(), &repair, basetypes.ObjectAsOptions{})

	res := v1.NewOptNilNosqlSettingsRepair(v1.NosqlSettingsRepair{})
	if !repair.Incremental.IsNull() && !repair.Incremental.IsUnknown() {
		var inc nosqlRepairIncrementalModel
		_ = repair.Incremental.As(context.Background(), &inc, basetypes.ObjectAsOptions{})

		values := common.TsetToStrings(inc.DaysOfWeek)
		settings := make([]v1.NosqlSettingsRepairIncrementalDaysOfWeekItem, len(values))
		for i, v := range values {
			settings[i] = v1.NosqlSettingsRepairIncrementalDaysOfWeekItem(v)
		}
		res.Value.Incremental = v1.NewOptNosqlSettingsRepairIncremental(v1.NosqlSettingsRepairIncremental{
			DaysOfWeek: settings,
			Time:       inc.Time.ValueString(),
		})
	}
	if !repair.Full.IsNull() && !repair.Full.IsUnknown() {
		var full nosqlRepairFullModel
		_ = repair.Full.As(context.Background(), &full, basetypes.ObjectAsOptions{})
		res.Value.Full = v1.NewOptNosqlSettingsRepairFull(v1.NosqlSettingsRepairFull{
			Interval:  v1.NosqlSettingsRepairFullInterval(int(full.Interval.ValueInt32())),
			DayOfWeek: v1.NosqlSettingsRepairFullDayOfWeek(full.DayOfWeek.ValueString()),
			Time:      full.Time.ValueString(),
		})
	}

	return res
}
