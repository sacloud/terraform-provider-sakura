// Copyright 2016-2025 terraform-provider-sakuracloud authors
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/power"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	databaseBuilder "github.com/sacloud/iaas-service-go/database/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type databaseReadReplicaResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &databaseReadReplicaResource{}
	_ resource.ResourceWithConfigure   = &databaseReadReplicaResource{}
	_ resource.ResourceWithImportState = &databaseReadReplicaResource{}
)

func NewDatabaseReadReplicaResource() resource.Resource {
	return &databaseReadReplicaResource{}
}

func (r *databaseReadReplicaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_read_replica"
}

func (r *databaseReadReplicaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type databaseReadReplicaResourceModel struct {
	common.SakuraBaseModel
	IconID           types.String                   `tfsdk:"icon_id"`
	Zone             types.String                   `tfsdk:"zone"`
	MasterID         types.String                   `tfsdk:"master_id"`
	NetworkInterface *databaseNetworkInterfaceModel `tfsdk:"network_interface"`
	Timeouts         timeouts.Value                 `tfsdk:"timeouts"`
}

func (r *databaseReadReplicaResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Database Read Replica"),
			"name":        common.SchemaResourceName("Database Read Replica"),
			"description": common.SchemaResourceDescription("Database Read Replica"),
			"tags":        common.SchemaResourceTags("Database Read Replica"),
			"icon_id":     common.SchemaResourceIconID("Database Read Replica"),
			"zone":        common.SchemaResourceZone("Database Read Replica"),
			"master_id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the replication master database.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_interface": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"switch_id": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The id of the switch to which the Database Replica connects. If `switch_id` isn't specified, it will be set to the same value of the master database",
						Validators: []validator.String{
							sacloudvalidator.SakuraIDValidator(),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"ip_address": schema.StringAttribute{
						Required:    true,
						Description: "The IP address to assign to the Database Replica",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"netmask": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Description: desc.Sprintf("The bit length of the subnet to assign to the Database Replica. %s. If `netmask` isn't specified, it will be set to the same value of the master database", desc.Range(8, 29)),
						Validators: []validator.Int32{
							int32validator.Between(8, 29),
						},
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"gateway": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The IP address of the gateway used by Database Replica. If `gateway` isn't specified, it will be set to the same value of the master database",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"source_ranges": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Description: "The range of source IP addresses that allow to access to the Database Replica via network",
					},
					"port": schema.Int32Attribute{
						Computed:    true,
						Description: "Placeholder for same code. Always 0",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *databaseReadReplicaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *databaseReadReplicaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseReadReplicaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, 60*time.Minute)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	builder, err := expandDatabaseReadReplicaBuilder(ctx, &plan, r.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to build SakuraCloud Database Read Replica builder: %s", err))
		return
	}
	db, err := builder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create SakuraCloud Database Read Replica: %s", err))
		return
	}

	// HACK データベースアプライアンスの電源投入後すぐに他の操作(Updateなど)を行うと202(Accepted)が返ってくるものの無視される。
	// この挙動はテストなどで問題となる。このためここで少しsleepすることで対応する。
	time.Sleep(databaseWaitAfterCreateDuration)

	if err := plan.updateState(zone, db); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to update SakuraCloud Database Read Replica state: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseReadReplicaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseReadReplicaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	db := getDatabase(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if db == nil {
		return
	}

	if err := state.updateState(zone, db); err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update state: %s", err))
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseReadReplicaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan databaseReadReplicaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, 60*time.Minute)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	dbID := common.ExpandSakuraCloudID(plan.ID)
	db := getDatabase(ctx, r.client, dbID, zone, &resp.State, &resp.Diagnostics)
	if db == nil {
		return
	}

	builder, err := expandDatabaseReadReplicaBuilder(ctx, &plan, r.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to build Database Read Replica builder: %s", err))
		return
	}
	builder.ID = dbID

	db, err = builder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update Database Read Replica: %s", err))
		return
	}

	if err := plan.updateState(zone, db); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update state: %s", err))
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseReadReplicaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state databaseReadReplicaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, state.Timeouts, 20*time.Minute)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	dbOp := iaas.NewDatabaseOp(r.client)
	db := getDatabase(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if db == nil {
		return
	}

	// shutdown(force) if running
	if db.InstanceStatus.IsUp() {
		if err := power.ShutdownDatabase(ctx, dbOp, zone, db.ID, true); err != nil {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to shutdown Database Read Replica: %s", err))
			return
		}
	}
	if err := dbOp.Delete(ctx, zone, db.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete Database Read Replica: %s", err))
		return
	}
}

func (model *databaseReadReplicaResourceModel) updateState(zone string, db *iaas.Database) error {
	if db.Availability.IsFailed() {
		return fmt.Errorf("got unexpected state: Database[%d].Availability is failed", db.ID)
	}

	model.UpdateBaseState(db.ID.String(), db.Name, db.Description, db.Tags)
	model.Tags = flattenDatabaseTags(db)
	model.Zone = types.StringValue(zone)
	model.MasterID = types.StringValue(db.ReplicationSetting.ApplianceID.String())
	model.NetworkInterface = flattenDatabaseReadReplicaNetworkInterface(db)

	return nil
}

func flattenDatabaseReadReplicaNetworkInterface(db *iaas.Database) *databaseNetworkInterfaceModel {
	return &databaseNetworkInterfaceModel{
		SwitchID:     types.StringValue(db.SwitchID.String()),
		IPAddress:    types.StringValue(db.IPAddresses[0]),
		Netmask:      types.Int32Value(int32(db.NetworkMaskLen)),
		Gateway:      types.StringValue(db.DefaultRoute),
		SourceRanges: common.StringsToTlist(db.CommonSetting.SourceNetwork),
		Port:         types.Int32Value(0),
	}
}

func expandDatabaseReadReplicaBuilder(ctx context.Context, model *databaseReadReplicaResourceModel, client *common.APIClient, zone string) (*databaseBuilder.Builder, error) {
	masterID := model.MasterID.ValueString()
	masterDB, err := iaas.NewDatabaseOp(client).Read(ctx, zone, common.SakuraCloudID(masterID))
	if err != nil {
		return nil, fmt.Errorf("master database instance[%s] is not found", masterID)
	}
	if masterDB.ReplicationSetting.Model != iaastypes.DatabaseReplicationModels.MasterSlave {
		return nil, fmt.Errorf("master database instance[%s] is not configured as ReplicationMaster", masterID)
	}

	nic := model.NetworkInterface
	switchID := masterDB.SwitchID.String()
	if !nic.SwitchID.IsNull() && !nic.SwitchID.IsUnknown() {
		switchID = nic.SwitchID.ValueString()
	}
	netmask := masterDB.NetworkMaskLen
	if !nic.Netmask.IsNull() && !nic.Netmask.IsUnknown() {
		netmask = int(nic.Netmask.ValueInt32())
	}
	gateway := masterDB.DefaultRoute
	if !nic.Gateway.IsNull() && !nic.Gateway.IsUnknown() {
		gateway = nic.Gateway.ValueString()
	}

	return &databaseBuilder.Builder{
		Zone:           zone,
		Name:           model.Name.ValueString(),
		Description:    model.Description.ValueString(),
		Tags:           common.TsetToStrings(model.Tags),
		IconID:         common.ExpandSakuraCloudID(model.IconID),
		PlanID:         iaastypes.ID(masterDB.PlanID.Int64() + 1),
		SwitchID:       common.SakuraCloudID(switchID),
		IPAddresses:    []string{nic.IPAddress.ValueString()},
		NetworkMaskLen: netmask,
		DefaultRoute:   gateway,
		Conf: &iaas.DatabaseRemarkDBConfCommon{
			DatabaseName:     masterDB.Conf.DatabaseName,
			DatabaseVersion:  masterDB.Conf.DatabaseVersion,
			DatabaseRevision: masterDB.Conf.DatabaseRevision,
		},
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:   masterDB.CommonSetting.ServicePort,
			SourceNetwork: common.TlistToStringsOrDefault(nic.SourceRanges),
		},
		ReplicationSetting: &iaas.DatabaseReplicationSetting{
			Model:       iaastypes.DatabaseReplicationModels.AsyncReplica,
			IPAddress:   masterDB.IPAddresses[0],
			Port:        masterDB.CommonSetting.ServicePort,
			User:        masterDB.ReplicationSetting.User,
			Password:    masterDB.ReplicationSetting.Password,
			ApplianceID: masterDB.ID,
		},
		Client: databaseBuilder.NewAPIClient(client),
	}, nil
}
