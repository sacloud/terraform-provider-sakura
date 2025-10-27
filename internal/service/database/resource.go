// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/power"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	databaseBuilder "github.com/sacloud/iaas-service-go/database/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type databaseResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &databaseResource{}
	_ resource.ResourceWithConfigure   = &databaseResource{}
	_ resource.ResourceWithImportState = &databaseResource{}
)

func NewDatabaseResource() resource.Resource {
	return &databaseResource{}
}

func (r *databaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *databaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type databaseResourceModel struct {
	/*
		common.SakuraBaseModel
		IconID           types.String   `tfsdk:"icon_id"`
		Zone             types.String   `tfsdk:"zone"`
		DatabaseType     types.String   `tfsdk:"database_type"`
		DatabaseVersion  types.String   `tfsdk:"database_version"`
		Plan             types.String   `tfsdk:"plan"`
		Username         types.String   `tfsdk:"username"`
		Password         types.String   `tfsdk:"password"`
		ReplicaUser      types.String   `tfsdk:"replica_user"`
		ReplicaPassword  types.String   `tfsdk:"replica_password"`
		NetworkInterface types.List     `tfsdk:"network_interface"`
		Backup           types.List     `tfsdk:"backup"`
		Parameters       types.Map      `tfsdk:"parameters"`
	*/
	databaseBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *databaseResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Database"),
			"name":        common.SchemaResourceName("Database"),
			"icon_id":     common.SchemaResourceIconID("Database"),
			"description": common.SchemaResourceDescription("Database"),
			"tags":        common.SchemaResourceTags("Database"),
			"zone":        common.SchemaResourceZone("Database"),
			"plan":        common.SchemaResourcePlan("Database", "10g", iaastypes.DatabasePlanStrings),
			"database_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The type of the database. This must be one of [%s]", iaastypes.RDBMSTypeStrings),
				Default:     stringdefault.StaticString(iaastypes.RDBMSTypesPostgreSQL.String()),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.RDBMSTypeStrings...),
				},
			},
			"database_version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The version of the database",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The name of default user on the database. %s", desc.Length(3, 20)),
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 20),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The password of default user on the database",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"replica_user": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of user that processing a replication",
				Default:     stringdefault.StaticString("replica"),
			},
			"replica_password": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "The password of user that processing a replication",
			},
			"network_interface": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"switch_id": common.SchemaResourceSwitchID("Database"),
					"ip_address": schema.StringAttribute{
						Required:    true,
						Description: "The IP address to assign to the Database",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"netmask": schema.Int32Attribute{
						Required:    true,
						Description: desc.Sprintf("The bit length of the subnet to assign to the Database. %s", desc.Range(8, 29)),
						Validators: []validator.Int32{
							int32validator.Between(8, 29),
						},
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"gateway": schema.StringAttribute{
						Required:    true,
						Description: "The IP address of the gateway used by Database",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"port": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Description: desc.Sprintf("The number of the listening port. %s", desc.Range(1024, 65535)),
						Default:     int32default.StaticInt32(5432),
						Validators: []validator.Int32{
							int32validator.Between(1024, 65535),
						},
					},
					"source_ranges": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Description: "The range of source IP addresses that allow to access to the Database via network",
					},
				},
			},
			"backup": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"weekdays": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: desc.Sprintf("A list of weekdays to backed up. The values in the list must be in [%s]", iaastypes.DaysOfTheWeekStrings),
					},
					"time": schema.StringAttribute{
						Optional:    true,
						Description: "The time to take backup. This must be formatted with `HH:mm`",
					},
				},
			},
			"parameters": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The map for setting RDBMS-specific parameters. Valid keys can be found with the `usacloud database list-parameters` command",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseResourceModel
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

	dbBuilder := expandDatabaseBuilder(&plan, r.client)
	dbBuilder.Zone = zone

	db, err := dbBuilder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create SakuraCloud Database: %s", err))
		return
	}

	// HACK データベースアプライアンスの電源投入後すぐに他の操作(Updateなど)を行うと202(Accepted)が返ってくるものの無視される。
	// この挙動はテストなどで問題となる。このためここで少しsleepすることで対応する。
	time.Sleep(databaseWaitAfterCreateDuration)

	if _, err := plan.updateState(ctx, r.client, zone, db); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to update SakuraCloud Database state: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := state.Zone.ValueString()
	dbID := common.ExpandSakuraCloudID(state.ID)
	db := getDatabase(ctx, r.client, dbID, zone, &resp.State, &resp.Diagnostics)
	if db == nil {
		return
	}

	if removeDB, err := state.updateState(ctx, r.client, zone, db); err != nil {
		if removeDB {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update SakuraCloud Database[%s] state: %s", db.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan databaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dbID := common.ExpandSakuraCloudID(plan.ID)
	dbBuilder := expandDatabaseBuilder(&plan, r.client)
	dbBuilder.Zone = zone
	dbBuilder.ID = dbID
	if _, err := dbBuilder.Build(ctx); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update SakuraCloud Database[%s]: %s", dbID, err))
		return
	}

	db := getDatabase(ctx, r.client, dbID, zone, &resp.State, &resp.Diagnostics)
	if db == nil {
		return
	}

	if removeDB, err := plan.updateState(ctx, r.client, zone, db); err != nil {
		if removeDB {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update SakuraCloud Database[%s] state: %s", db.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state databaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := state.Zone.ValueString()
	dbOp := iaas.NewDatabaseOp(r.client)
	db := getDatabase(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if db == nil {
		return
	}

	if db.InstanceStatus.IsUp() {
		if err := power.ShutdownDatabase(ctx, dbOp, zone, db.ID, true); err != nil {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to shutdown SakuraCloud Database[%s]: %s", db.ID, err))
			return
		}
	}

	if err := dbOp.Delete(ctx, zone, db.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete SakuraCloud Database[%s]: %s", db.ID, err))
		return
	}
}

func getDatabase(ctx context.Context, client *common.APIClient, id iaastypes.ID, zone string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Database {
	dbOp := iaas.NewDatabaseOp(client)
	db, err := dbOp.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Read Error", fmt.Sprintf("failed to read SakuraCloud Database[%s] : %s", id, err))
		return nil
	}
	return db
}

func expandDatabaseBuilder(model *databaseResourceModel, client *common.APIClient) *databaseBuilder.Builder {
	dbType := model.DatabaseType.ValueString()
	dbName := iaastypes.RDBMSTypeFromString(dbType)
	nic := model.NetworkInterface
	replicaUser := model.ReplicaUser.ValueString()
	replicaPassword := model.ReplicaPassword.ValueString()

	req := &databaseBuilder.Builder{
		PlanID:         iaastypes.DatabasePlanIDMap[model.Plan.ValueString()],
		SwitchID:       common.ExpandSakuraCloudID(nic.SwitchID),
		IPAddresses:    []string{nic.IPAddress.ValueString()},
		NetworkMaskLen: int(nic.Netmask.ValueInt32()),
		DefaultRoute:   nic.Gateway.ValueString(),
		Conf: &iaas.DatabaseRemarkDBConfCommon{
			DatabaseName:    dbName.String(),
			DatabaseVersion: model.DatabaseVersion.ValueString(),
			DefaultUser:     model.Username.ValueString(),
			UserPassword:    model.Password.ValueString(),
		},
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     int(nic.Port.ValueInt32()),
			SourceNetwork:   common.TlistToStringsOrDefault(nic.SourceRanges),
			DefaultUser:     model.Username.ValueString(),
			UserPassword:    model.Password.ValueString(),
			ReplicaUser:     replicaUser,
			ReplicaPassword: replicaPassword,
		},
		Name:               model.Name.ValueString(),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		Client:             databaseBuilder.NewAPIClient(client),
		BackupSetting:      expandDatabaseBackupSetting(model),
		Parameters:         expandParameters(model.Parameters),
		ReplicationSetting: &iaas.DatabaseReplicationSetting{},
	}
	if replicaUser != "" && replicaPassword != "" {
		req.ReplicationSetting = &iaas.DatabaseReplicationSetting{
			Model:    iaastypes.DatabaseReplicationModels.MasterSlave,
			User:     replicaUser,
			Password: replicaPassword,
		}
	}

	return req
}

func expandParameters(values types.Map) map[string]interface{} {
	if values.IsNull() || values.IsUnknown() {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range values.Elements() {
		if vStr, ok := v.(types.String); ok && !vStr.IsNull() && !vStr.IsUnknown() {
			result[k] = vStr.ValueString()
		}
	}
	return result
}

func expandDatabaseBackupSetting(model *databaseResourceModel) *iaas.DatabaseSettingBackup {
	if model.Backup.IsNull() || model.Backup.IsUnknown() {
		return nil
	}

	var backup databaseBackupModel
	diags := model.Backup.As(context.Background(), &backup, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	backupTime := backup.Time.ValueString()
	backupWeekdays := common.ExpandBackupWeekdays(backup.Weekdays)
	if backupTime != "" && len(backupWeekdays) > 0 {
		return &iaas.DatabaseSettingBackup{
			Time:      backupTime,
			DayOfWeek: backupWeekdays,
		}
	}

	return nil
}
