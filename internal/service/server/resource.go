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

package server

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/power"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	diskBuilder "github.com/sacloud/iaas-service-go/disk/builder"
	serverBuilder "github.com/sacloud/iaas-service-go/server/builder"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakuracloud/internal/validator"
)

type serverResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &serverResource{}
	_ resource.ResourceWithConfigure   = &serverResource{}
	_ resource.ResourceWithImportState = &serverResource{}
)

func NewServerResource() resource.Resource {
	return &serverResource{}
}

func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *serverResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type serverResourceModel struct {
	serverBaseModel
	UserData      types.String         `tfsdk:"user_data"`
	DiskEdit      *serverDiskEditModel `tfsdk:"disk_edit_parameter"`
	ForceShutdown types.Bool           `tfsdk:"force_shutdown"`
	Timeouts      timeouts.Value       `tfsdk:"timeouts"`
}

type serverDiskEditModel struct {
	Hostname            types.String               `tfsdk:"hostname"`
	Password            types.String               `tfsdk:"password"`
	SSHKeyIDs           types.Set                  `tfsdk:"ssh_key_ids"`
	SSHKeys             types.Set                  `tfsdk:"ssh_keys"`
	DisablePwAuth       types.Bool                 `tfsdk:"disable_pw_auth"`
	EnableDHCP          types.Bool                 `tfsdk:"enable_dhcp"`
	ChangePartitionUUID types.Bool                 `tfsdk:"change_partition_uuid"`
	IPAddress           iptypes.IPv4Address        `tfsdk:"ip_address"`
	Gateway             types.String               `tfsdk:"gateway"`
	Netmask             types.Int32                `tfsdk:"netmask"`
	Note                []*serverDiskEditNoteModel `tfsdk:"note"`
}

type serverDiskEditNoteModel struct {
	ID        types.String `tfsdk:"id"`
	APIKeyID  types.String `tfsdk:"api_key_id"`
	Variables types.Map    `tfsdk:"variables"`
}

func (r *serverResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Server"),
			"name":        common.SchemaResourceName("Server"),
			"description": common.SchemaResourceDescription("Server"),
			"tags":        common.SchemaResourceTags("Server"),
			"zone":        common.SchemaResourceZone("Server"),
			"icon_id":     common.SchemaResourceIconID("Server"),
			"core": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "The number of virtual CPUs",
			},
			"memory": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "The size of memory in GiB",
			},
			"gpu": schema.Int64Attribute{
				Optional:    true,
				Description: "The number of GPUs",
			},
			"cpu_model": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The model of cpu",
				PlanModifiers: []planmodifier.String{
					// cpu_modelはUpdate時のシャットダウン判定(IsNeedShutdown)で使われるため、Unknown時はStateの値を利用する
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"commitment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(iaastypes.Commitments.Standard.String()),
				Description: desc.Sprintf("The policy of how to allocate virtual CPUs to the server. This must be one of [%s]", iaastypes.CommitmentStrings),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.CommitmentStrings...),
				},
			},
			"disks": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A set of disk id connected to the server",
			},
			"interface_driver": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(iaastypes.InterfaceDrivers.VirtIO.String()),
				Description: desc.Sprintf("The driver name of network interface. This must be one of [%s]", iaastypes.InterfaceDriverStrings),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.InterfaceDriverStrings...),
				},
			},
			"network_interface": schema.ListNestedAttribute{
				Optional: true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"upstream": schema.StringAttribute{
							Required: true,
							Description: desc.Sprintf(
								"The upstream type or upstream switch id. This must be one of [%s]",
								[]string{"shared", "disconnect", "<switch id>"},
							),
							Validators: []validator.String{
								// validateSakuraCloudServerNIC in v2. ここでしか使われていないのでStringFuncValidatorで実装
								sacloudvalidator.StringFuncValidator(func(v string) error {
									if v == "" || v == "shared" || v == "disconnect" {
										return nil
									}
									_, err := strconv.ParseInt(v, 10, 64)
									if err != nil {
										return fmt.Errorf("upstream must be SakuraCloud ID string(number only): %s", err)
									}

									return nil
								}),
							},
						},
						"user_ip_address": schema.StringAttribute{
							//CustomType: iptypes.IPv4AddressType{},
							Optional:    true,
							Computed:    true,
							Description: "The IP address for only display. This value doesn't affect actual NIC settings",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
						"packet_filter_id": schema.StringAttribute{
							Optional:    true,
							Description: "The id of the packet filter to attach to the network interface",
							Validators: []validator.String{
								sacloudvalidator.SakuraIDValidator(),
							},
						},
						"mac_address": schema.StringAttribute{
							Computed:    true,
							Description: "The MAC address",
						},
					},
				},
			},
			"cdrom_id": schema.StringAttribute{
				Optional:    true,
				Description: "The id of the CD-ROM to attach to the Server",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"private_host_id": schema.StringAttribute{
				Optional:    true,
				Description: "The id of the PrivateHost which the Server is assigned",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"private_host_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The id of the PrivateHost which the Server is assigned",
			},
			"user_data": schema.StringAttribute{
				Optional:    true,
				Description: desc.Sprintf("A string representing the user data used by cloud-init. %s", desc.Conflicts("disk_edit_parameter")),
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("disk_edit_parameter")),
				},
			},
			"disk_edit_parameter": schema.SingleNestedAttribute{
				Optional: true,
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("user_data")),
				},
				Attributes: map[string]schema.Attribute{
					"hostname": schema.StringAttribute{
						Optional:    true,
						Description: desc.Sprintf("The hostname of the Server. %s", desc.Length(1, 64)),
					},
					"password": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: desc.Sprintf("The password of default user. %s", desc.Length(12, 128)),
						Validators: []validator.String{
							stringvalidator.LengthBetween(12, 128),
						},
					},
					"ssh_key_ids": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "A set of the SSHKey id",
					},
					"ssh_keys": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "A set of the SSHKey text",
					},
					"disable_pw_auth": schema.BoolAttribute{
						Optional:    true,
						Description: "The flag to disable password authentication",
					},
					"enable_dhcp": schema.BoolAttribute{
						Optional:    true,
						Description: "The flag to enable DHCP client",
					},
					"change_partition_uuid": schema.BoolAttribute{
						Optional:    true,
						Description: "The flag to change partition uuid",
					},
					"note": schema.ListNestedAttribute{
						Optional:    true,
						Description: "A list of the Note/StartupScript",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required:    true,
									Description: "The id of the note",
									Validators: []validator.String{
										sacloudvalidator.SakuraIDValidator(),
									},
								},
								"api_key_id": schema.StringAttribute{
									Optional:    true,
									Description: "The id of the API key to be injected into note when editing the disk",
									Validators: []validator.String{
										sacloudvalidator.SakuraIDValidator(),
									},
								},
								"variables": schema.MapAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Description: "The value of the variable that be injected into note when editing the disk",
								},
							},
						},
					},
					"ip_address": schema.StringAttribute{
						//CustomType: iptypes.IPv4AddressType{},
						Optional:    true,
						Description: "The IP address to assign to the Server",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
					"gateway": schema.StringAttribute{
						//CustomType: iptypes.IPv4AddressType{},
						Optional:    true,
						Description: "The gateway address used by the Server",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
					"netmask": schema.Int32Attribute{
						Optional:    true,
						Description: "The bit length of the subnet to assign to the Server",
					},
				},
			},
			"ip_address": schema.StringAttribute{
				//CustomType:  iptypes.IPAddressType{},
				Computed:    true,
				Description: "The IP address assigned to the Server",
				Validators: []validator.String{
					sacloudvalidator.IPAddressValidator(sacloudvalidator.Both),
				},
			},
			"gateway": schema.StringAttribute{
				//CustomType:  iptypes.IPAddressType{},
				Computed:    true,
				Description: "The IP address of the gateway used by Server",
				Validators: []validator.String{
					sacloudvalidator.IPAddressValidator(sacloudvalidator.Both),
				},
			},
			"network_address": schema.StringAttribute{
				//CustomType:  iptypes.IPAddressType{},
				Computed:    true,
				Description: "The network address which the `ip_address` belongs",
				Validators: []validator.String{
					sacloudvalidator.IPAddressValidator(sacloudvalidator.Both),
				},
			},
			"netmask": schema.Int32Attribute{
				Computed:    true,
				Description: "The bit length of the subnet assigned to the Server",
			},
			"dns_servers": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A set of IP address of DNS server in the zone",
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the Server",
			},
			"force_shutdown": schema.BoolAttribute{
				Optional:    true,
				Description: "The flag to use force shutdown when need to reboot/shutdown while applying",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	builder, err := expandServerBuilder(ctx, r.client, zone, &plan, nil)
	if err != nil {
		resp.Diagnostics.AddError("Expand Server Builder Error", err.Error())
		return
	}
	if err := builder.Validate(ctx, zone); err != nil {
		resp.Diagnostics.AddError("Validate Server Builder Error", err.Error())
		return
	}
	result, err := builder.Build(ctx, zone)
	if err != nil {
		resp.Diagnostics.AddError("Build Server Error", err.Error())
		return
	}

	server := getServer(ctx, r.client, zone, result.ServerID, &resp.State, &resp.Diagnostics)
	if server == nil {
		return
	}

	plan.updateState(server, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	server := getServer(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if server == nil {
		return
	}

	state.updateState(server, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	sid := state.ID.ValueString()
	common.SakuraMutexKV.Lock(sid)
	defer common.SakuraMutexKV.Unlock(sid)

	builder, err := expandServerBuilder(ctx, r.client, zone, &plan, &state)
	if err != nil {
		resp.Diagnostics.AddError("Expand Server Builder Error", err.Error())
		return
	}
	if err := builder.Validate(ctx, zone); err != nil {
		resp.Diagnostics.AddError("Validate Server Builder Error", fmt.Sprintf("validating SakuraCloud Server[%s] is failed: %s", sid, err))
		return
	}
	result, err := builder.Update(ctx, zone)
	if err != nil {
		resp.Diagnostics.AddError("Update Server Error", fmt.Sprintf("updating SakuraCloud Server[%s] is failed: %s", sid, err))
		return

	}

	server := getServer(ctx, r.client, zone, result.ServerID, &resp.State, &resp.Diagnostics)
	if server == nil {
		return
	}

	plan.updateState(server, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	sid := state.ID.ValueString()
	common.SakuraMutexKV.Lock(sid)
	defer common.SakuraMutexKV.Unlock(sid)

	serverOp := iaas.NewServerOp(r.client)
	server := getServer(ctx, r.client, zone, common.SakuraCloudID(sid), &resp.State, &resp.Diagnostics)
	if server.InstanceStatus.IsUp() {
		if err := power.ShutdownServer(ctx, serverOp, zone, server.ID, state.ForceShutdown.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Shutdown Error", fmt.Sprintf("stopping SakuraCloud Server[%s] is failed: %s", server.ID, err))
			return
		}
	}

	if err := serverOp.Delete(ctx, zone, server.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud Server[%s] is failed: %s", server.ID, err))
		return
	}
}

func getServer(ctx context.Context, client *common.APIClient, zone string, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Server {
	serverOp := iaas.NewServerOp(client)
	server, err := serverOp.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get Server Error", fmt.Sprintf("could not read SakuraCloud Server[%s]: %s", id, err))
	}

	return server
}

func expandServerBuilder(ctx context.Context, client *common.APIClient, zone string, plan *serverResourceModel, state *serverResourceModel) (*serverBuilder.Builder, error) {
	diskBuilders, err := expandServerDisks(ctx, client, zone, plan, state)
	if err != nil {
		return nil, err
	}
	return &serverBuilder.Builder{
		ServerID:        common.ExpandSakuraCloudID(plan.ID),
		Name:            plan.Name.ValueString(),
		CPU:             int(plan.Core.ValueInt64()),
		MemoryGB:        int(plan.Memory.ValueInt64()),
		GPU:             int(plan.GPU.ValueInt64()),
		CPUModel:        plan.CPUModel.ValueString(),
		Commitment:      iaastypes.ECommitment(plan.Commitment.ValueString()),
		Generation:      iaastypes.PlanGenerations.Default,
		InterfaceDriver: iaastypes.EInterfaceDriver(plan.InterfaceDriver.ValueString()),
		Description:     plan.Description.ValueString(),
		IconID:          common.ExpandSakuraCloudID(plan.IconID),
		Tags:            common.TsetToStrings(plan.Tags),
		CDROMID:         common.ExpandSakuraCloudID(plan.CDROMID),
		PrivateHostID:   common.ExpandSakuraCloudID(plan.PrivateHostID),
		NIC:             expandServerNIC(plan),
		AdditionalNICs:  expandServerAdditionalNICs(plan),
		DiskBuilders:    diskBuilders,
		Client:          serverBuilder.NewBuildersAPIClient(client),
		ForceShutdown:   plan.ForceShutdown.ValueBool(),
		BootAfterCreate: true,
		UserData:        expandServerUserData(plan, state),
	}, nil
}

func expandServerUserData(plan *serverResourceModel, state *serverResourceModel) string {
	if state == nil {
		return plan.UserData.ValueString()
	} else {
		if !plan.UserData.Equal(state.UserData) {
			return plan.UserData.ValueString()
		}
	}
	return ""
}

func expandServerDisks(ctx context.Context, client *common.APIClient, zone string, plan *serverResourceModel, state *serverResourceModel) ([]diskBuilder.Builder, error) {
	var builders []diskBuilder.Builder
	diskIDs := common.ExpandSakuraCloudIDs(plan.Disks)
	diskOp := iaas.NewDiskOp(client)
	for i, diskID := range diskIDs {
		disk, err := diskOp.Read(ctx, zone, diskID)
		if err != nil {
			return nil, err
		}
		b := &diskBuilder.ConnectedDiskBuilder{
			ID:          diskID,
			Name:        disk.Name,
			Description: disk.Description,
			Tags:        disk.Tags,
			IconID:      disk.IconID,
			Connection:  disk.Connection,
			Client:      diskBuilder.NewBuildersAPIClient(client),
		}
		// set only when value was changed
		if i == 0 && isDiskEditParameterChanged(plan, state) {
			if plan.DiskEdit != nil {
				de := plan.DiskEdit
				log.Printf("[INFO] disk_edit_parameter is specified for Disk[%s]", diskID)
				b.EditParameter = &diskBuilder.UnixEditRequest{
					HostName:            de.Hostname.ValueString(),
					Password:            de.Password.ValueString(),
					DisablePWAuth:       de.DisablePwAuth.ValueBool(),
					EnableDHCP:          de.EnableDHCP.ValueBool(),
					ChangePartitionUUID: de.ChangePartitionUUID.ValueBool(),
					IPAddress:           de.IPAddress.ValueString(),
					NetworkMaskLen:      int(de.Netmask.ValueInt32()),
					DefaultRoute:        de.Gateway.ValueString(),
					SSHKeys:             common.TsetToStringsOrDefault(de.SSHKeys),
					SSHKeyIDs:           common.ExpandSakuraCloudIDs(de.SSHKeyIDs),
					Notes:               expandDiskEditNotes(de),
				}
			}
		}
		builders = append(builders, b)
	}
	return builders, nil
}

func expandDiskEditNotes(model *serverDiskEditModel) []*iaas.DiskEditNote {
	var notes []*iaas.DiskEditNote
	if model.Note != nil {
		for _, note := range model.Note {
			notes = append(notes, &iaas.DiskEditNote{
				ID:        common.ExpandSakuraCloudID(note.ID),
				APIKeyID:  common.ExpandSakuraCloudID(note.APIKeyID),
				Variables: tmapToVariables(note.Variables),
			})
		}
	}
	return notes
}

func tmapToVariables(d types.Map) map[string]interface{} {
	if d.IsNull() || d.IsUnknown() {
		return nil
	}

	kv := make(map[string]interface{})
	for k, v := range d.Elements() {
		if vStr, ok := v.(types.String); ok && !vStr.IsNull() && !vStr.IsUnknown() {
			kv[k] = vStr.ValueString()
		}
	}
	return kv
}

func expandServerNIC(model *serverResourceModel) serverBuilder.NICSettingHolder {
	nics := model.NetworkInterface
	if len(nics) == 0 {
		return nil
	}

	nic := nics[0]
	upstream := nic.Upstream.ValueString()
	switch upstream {
	case "", "shared":
		return &serverBuilder.SharedNICSetting{
			PacketFilterID: common.ExpandSakuraCloudID(nic.PacketFilterID),
		}
	case "disconnect":
		return &serverBuilder.DisconnectedNICSetting{}
	default:
		return &serverBuilder.ConnectedNICSetting{
			SwitchID:         common.SakuraCloudID(upstream),
			PacketFilterID:   common.ExpandSakuraCloudID(nic.PacketFilterID),
			DisplayIPAddress: nic.UserIPAddress.ValueString(),
		}
	}
}

func expandServerAdditionalNICs(model *serverResourceModel) []serverBuilder.AdditionalNICSettingHolder {
	var results []serverBuilder.AdditionalNICSettingHolder

	nics := model.NetworkInterface
	if len(nics) < 2 {
		return results
	}

	for i, nic := range nics {
		if i == 0 {
			continue
		}
		upstream := nic.Upstream.ValueString()
		switch upstream {
		case "disconnect":
			results = append(results, &serverBuilder.DisconnectedNICSetting{})
		default:
			results = append(results, &serverBuilder.ConnectedNICSetting{
				SwitchID:         common.SakuraCloudID(upstream),
				PacketFilterID:   common.ExpandSakuraCloudID(nic.PacketFilterID),
				DisplayIPAddress: nic.UserIPAddress.ValueString(),
			})
		}
	}

	return results
}

func isDiskEditParameterChanged(plan, state *serverResourceModel) bool {
	// CreateにおいてHasChangesは常にtrueなので、Stateがnilの場合は常にtrueを返す
	// https://developer.hashicorp.com/terraform/plugin/framework/migrating/benefits#expanded-access-to-configuration-plan-and-state-data
	if state == nil {
		return true
	} else {
		if common.HasChange(plan.NetworkInterface, state.NetworkInterface) && isUpstreamChanged(plan.NetworkInterface, state.NetworkInterface) {
			return true
		}
		if !state.Disks.Equal(plan.Disks) {
			return true
		}
		if common.HasChange(plan.DiskEdit, state.DiskEdit) {
			return true
		}
		return false
	}
}

func isUpstreamChanged(new, old []serverNetworkInterfaceModel) bool {
	oldIsNil := len(old) == 0
	newIsNil := len(new) == 0

	if oldIsNil && newIsNil {
		return false
	}
	if oldIsNil != newIsNil {
		return true
	}
	if len(old) != len(new) {
		return true
	}

	for i := range old {
		oldUpstream := old[i].Upstream.ValueString()
		newUpstream := new[i].Upstream.ValueString()
		if oldUpstream != newUpstream {
			return true
		}
	}

	return false
}
