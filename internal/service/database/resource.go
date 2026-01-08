// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
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
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
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
	databaseBaseModel
	Password          types.String   `tfsdk:"password"`
	PasswordWO        types.String   `tfsdk:"password_wo"`
	ReplicaPassword   types.String   `tfsdk:"replica_password"`
	ReplicaPasswordWO types.String   `tfsdk:"replica_password_wo"`
	PasswordWOVersion types.Int32    `tfsdk:"password_wo_version"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
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
				Optional:    true,
				Sensitive:   true,
				Description: "The password of default user on the database. Use password_wo instead for newer deployments.",
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("password_wo")),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_wo")),
				},
			},
			"password_wo": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "The password of default user on the database",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password")),
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
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
				Sensitive:   true,
				Description: "The password of user that processing a replication. Use replica_password_wo instead for newer deployments.",
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("replica_password_wo")),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("replica_password_wo")),
				},
			},
			"replica_password_wo": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "The password of user that processing a replication",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("replica_password")),
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
				},
			},
			"password_wo_version": schema.Int32Attribute{
				Optional:    true,
				Description: "The version of the password_wo/replica_password_wo field. This value must be greater than 0 when set. Increment this when changing password.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"network_interface": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"vswitch_id": common.SchemaResourceSwitchID("Database"),
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
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("continuous_backup")),
				},
				Attributes: map[string]schema.Attribute{
					"days_of_week": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: desc.Sprintf("A list of days of week to backed up. The values in the list must be in [%s]", iaastypes.DaysOfTheWeekStrings),
					},
					"time": schema.StringAttribute{
						Required:    true,
						Description: "The time to take backup. This must be formatted with `HH:mm`",
						Validators: []validator.String{
							sacloudvalidator.BackupTimeValidator(),
						},
					},
				},
			},
			"continuous_backup": schema.SingleNestedAttribute{
				Optional: true,
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("backup")),
				},
				Attributes: map[string]schema.Attribute{
					"days_of_week": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: desc.Sprintf("A list of days of week to backed up. The values in the list must be in [%s]", iaastypes.DaysOfTheWeekStrings),
					},
					"time": schema.StringAttribute{
						Required:    true,
						Description: "The time to take backup. This must be formatted with `HH:mm`",
						Validators: []validator.String{
							sacloudvalidator.BackupTimeValidator(),
						},
					},
					"connect": schema.StringAttribute{
						Required:    true,
						Description: "NFS server address for storing backups (e.g., `nfs://192.0.2.1/export`)",
					},
				},
			},
			"parameters": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The map for setting RDBMS-specific parameters. Valid keys can be found with the `usacloud database list-parameters` command",
			},
			"disk":             common.SchemaResourceEncryptionDisk("Database"),
			"monitoring_suite": common.SchemaResourceMonitoringSuite("Database"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Database appliance.",
	}
}

func (r *databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config databaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dbBuilder := expandDatabaseBuilder(&plan, &config, r.client)
	dbBuilder.Zone = zone
	db, err := dbBuilder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Database: %s", err))
		return
	}

	// HACK データベースアプライアンスの電源投入後すぐに他の操作(Updateなど)を行うと202(Accepted)が返ってくるものの無視される。
	// この挙動はテストなどで問題となる。このためここで少しsleepすることで対応する。
	time.Sleep(databaseWaitAfterCreateDuration)

	if _, err := plan.updateState(ctx, r.client, zone, db); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update Database state: %s", err))
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

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dbID := common.ExpandSakuraCloudID(state.ID)
	db := getDatabase(ctx, r.client, dbID, zone, &resp.State, &resp.Diagnostics)
	if db == nil {
		return
	}

	if removeDB, err := state.updateState(ctx, r.client, zone, db); err != nil {
		if removeDB {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update Database[%s] state: %s", db.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config databaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dbID := common.ExpandSakuraCloudID(plan.ID)
	dbBuilder := expandDatabaseBuilder(&plan, &config, r.client)
	dbBuilder.Zone = zone
	dbBuilder.ID = dbID
	if _, err := dbBuilder.Build(ctx); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Database[%s]: %s", dbID.String(), err))
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
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update Database[%s] state: %s", db.ID.String(), err))
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

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	zone := state.Zone.ValueString()
	dbOp := iaas.NewDatabaseOp(r.client)
	db := getDatabase(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if db == nil {
		return
	}

	if db.InstanceStatus.IsUp() {
		if err := power.ShutdownDatabase(ctx, dbOp, zone, db.ID, true); err != nil {
			resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to shutdown Database[%s]: %s", db.ID.String(), err))
			return
		}
	}

	if err := dbOp.Delete(ctx, zone, db.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Database[%s]: %s", db.ID.String(), err))
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
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Database[%s]: %s", id, err))
		return nil
	}
	return db
}

func expandDatabaseBuilder(model, config *databaseResourceModel, client *common.APIClient) *databaseBuilder.Builder {
	dbType := model.DatabaseType.ValueString()
	dbName := iaastypes.RDBMSTypeFromString(dbType)
	nic := model.NetworkInterface
	replicaUser := model.ReplicaUser.ValueString()
	password := config.PasswordWO.ValueString()
	if password == "" {
		password = model.Password.ValueString()
	}
	replicaPassword := config.ReplicaPasswordWO.ValueString()
	if replicaPassword == "" {
		replicaPassword = model.ReplicaPassword.ValueString()
	}

	req := &databaseBuilder.Builder{
		PlanID:         iaastypes.DatabasePlanIDMap[model.Plan.ValueString()],
		SwitchID:       common.ExpandSakuraCloudID(nic.VSwitchID),
		IPAddresses:    []string{nic.IPAddress.ValueString()},
		NetworkMaskLen: int(nic.Netmask.ValueInt32()),
		DefaultRoute:   nic.Gateway.ValueString(),
		Conf: &iaas.DatabaseRemarkDBConfCommon{
			DatabaseName:    dbName.String(),
			DatabaseVersion: model.DatabaseVersion.ValueString(),
			DefaultUser:     model.Username.ValueString(),
			UserPassword:    password,
		},
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     int(nic.Port.ValueInt32()),
			SourceNetwork:   common.TlistToStringsOrDefault(nic.SourceRanges),
			DefaultUser:     model.Username.ValueString(),
			UserPassword:    password,
			ReplicaUser:     replicaUser,
			ReplicaPassword: replicaPassword,
		},
		Name:               model.Name.ValueString(),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		Client:             databaseBuilder.NewAPIClient(client),
		BackupSetting:      expandDatabaseBackupSetting(model),
		Backupv2Setting:    expandDatabaseContinuousBackup(model),
		Parameters:         expandParameters(model.Parameters),
		ReplicationSetting: &iaas.DatabaseReplicationSetting{},
		Disk:               expandDatabaseDisk(model.Disk),
		MonitoringSuite:    common.ExpandMonitoringSuite(model.MonitoringSuite),
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
	backupDaysOfWeek := common.ExpandBackupWeekdays(backup.DaysOfWeek)
	return &iaas.DatabaseSettingBackup{
		Time:      backupTime,
		DayOfWeek: backupDaysOfWeek,
		Rotate:    8,
	}
}

func expandDatabaseContinuousBackup(model *databaseResourceModel) *iaas.DatabaseSettingBackupv2 {
	if model.ContinuousBackup == nil {
		return nil
	}

	return &iaas.DatabaseSettingBackupv2{
		Time:      model.ContinuousBackup.Time.ValueString(),
		DayOfWeek: common.ExpandBackupWeekdays(model.ContinuousBackup.DaysOfWeek),
		Connect:   model.ContinuousBackup.Connect.ValueString(),
		Rotate:    8,
	}
}

func expandDatabaseDisk(disk types.Object) *iaas.DatabaseDisk {
	if disk.IsNull() || disk.IsUnknown() {
		return nil
	}

	var diskModel common.SakuraEncryptionDiskModel
	diags := disk.As(context.Background(), &diskModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &iaas.DatabaseDisk{
		EncryptionAlgorithm: iaastypes.EDiskEncryptionAlgorithm(diskModel.EncryptionAlgorithm.ValueString()),
		EncryptionKeyID:     iaastypes.StringID(diskModel.KMSKeyID.ValueString()),
	}
}
