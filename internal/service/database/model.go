// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

var databaseWaitAfterCreateDuration = 1 * time.Minute

type databaseBaseModel struct {
	common.SakuraBaseModel
	IconID           types.String                   `tfsdk:"icon_id"`
	Zone             types.String                   `tfsdk:"zone"`
	DatabaseType     types.String                   `tfsdk:"database_type"`
	DatabaseVersion  types.String                   `tfsdk:"database_version"`
	Plan             types.String                   `tfsdk:"plan"`
	Username         types.String                   `tfsdk:"username"`
	Password         types.String                   `tfsdk:"password"`
	ReplicaUser      types.String                   `tfsdk:"replica_user"`
	ReplicaPassword  types.String                   `tfsdk:"replica_password"`
	NetworkInterface *databaseNetworkInterfaceModel `tfsdk:"network_interface"`
	Backup           types.Object                   `tfsdk:"backup"`
	Parameters       types.Map                      `tfsdk:"parameters"`
}

type databaseNetworkInterfaceModel struct {
	SwitchID     types.String `tfsdk:"switch_id"`
	IPAddress    types.String `tfsdk:"ip_address"`
	Netmask      types.Int32  `tfsdk:"netmask"`
	Gateway      types.String `tfsdk:"gateway"`
	Port         types.Int32  `tfsdk:"port"`
	SourceRanges types.List   `tfsdk:"source_ranges"`
}

type databaseBackupModel struct {
	Weekdays types.Set    `tfsdk:"weekdays"`
	Time     types.String `tfsdk:"time"`
}

func (m databaseBackupModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"weekdays": types.SetType{ElemType: types.StringType},
		"time":     types.StringType,
	}
}

func (model *databaseBaseModel) updateState(ctx context.Context, client *common.APIClient, zone string, db *iaas.Database) (bool, error) {
	if db.Availability.IsFailed() {
		return true, fmt.Errorf("got unexpected state: Database[%d].Availability is failed", db.ID)
	}
	parameters, err := iaas.NewDatabaseOp(client).GetParameter(ctx, zone, db.ID)
	if err != nil {
		return false, err
	}

	model.UpdateBaseState(db.ID.String(), db.Name, db.Description, db.Tags)
	model.Tags = flattenDatabaseTags(db)
	model.Zone = types.StringValue(zone)
	model.DatabaseType = flattenDatabaseType(db)
	model.DatabaseVersion = types.StringValue(db.Conf.DatabaseVersion)
	model.Plan = types.StringValue(iaastypes.DatabasePlanNameMap[db.PlanID])
	model.Username = types.StringValue(db.CommonSetting.DefaultUser)
	model.Password = types.StringValue(db.CommonSetting.UserPassword)
	if db.ReplicationSetting != nil {
		model.ReplicaUser = types.StringValue(db.CommonSetting.ReplicaUser)
		model.ReplicaPassword = types.StringValue(db.CommonSetting.ReplicaPassword)
	}
	model.NetworkInterface = flattenDatabaseNetworkInterface(db)
	model.Backup = flattenDatabaseBackupSetting(db)
	model.Parameters = convertDatabaseParametersToMap(parameters)

	return false, nil
}

func flattenDatabaseType(db *iaas.Database) types.String {
	return types.StringValue(strings.ToLower(db.Conf.DatabaseName))
}

func flattenDatabaseTags(db *iaas.Database) types.Set {
	var tags iaastypes.Tags
	for _, t := range db.Tags {
		if !(strings.HasPrefix(t, "@MariaDB-") || strings.HasPrefix(t, "@postgres-")) {
			tags = append(tags, t)
		}
	}
	return common.FlattenTags(tags)
}

func flattenDatabaseNetworkInterface(db *iaas.Database) *databaseNetworkInterfaceModel {
	return &databaseNetworkInterfaceModel{
		SwitchID:     types.StringValue(db.SwitchID.String()),
		IPAddress:    types.StringValue(db.IPAddresses[0]),
		Netmask:      types.Int32Value(int32(db.NetworkMaskLen)),
		Gateway:      types.StringValue(db.DefaultRoute),
		Port:         types.Int32Value(int32(db.CommonSetting.ServicePort)),
		SourceRanges: common.StringsToTlist(db.CommonSetting.SourceNetwork),
	}
}

func flattenDatabaseBackupSetting(db *iaas.Database) types.Object {
	v := types.ObjectNull(databaseBackupModel{}.AttributeTypes())
	if db.BackupSetting != nil {
		m := databaseBackupModel{
			Time:     types.StringValue(db.BackupSetting.Time),
			Weekdays: common.FlattenBackupWeekdays(db.BackupSetting.DayOfWeek),
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	}
	return v
}

func convertDatabaseParametersToMap(parameter *iaas.DatabaseParameter) types.Map {
	stringMap := make(map[string]string)
	for k, v := range parameter.Settings {
		switch vv := v.(type) {
		case fmt.Stringer:
			stringMap[k] = vv.String()
		case string:
			stringMap[k] = vv
		default:
			stringMap[k] = fmt.Sprintf("%v", vv)
		}
	}
	dest := make(map[string]string)
	for k, v := range stringMap {
		for _, meta := range parameter.MetaInfo {
			if k == meta.Name {
				dest[meta.Label] = v
			}
		}
	}
	return common.StrMapToTmap(dest)
}
