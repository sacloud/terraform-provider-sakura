// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nosql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sacloud/nosql-api-go"
	v1 "github.com/sacloud/nosql-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type nosqlBaseModel struct {
	common.SakuraBaseModel
	Plan         types.String        `tfsdk:"plan"`
	Zone         types.String        `tfsdk:"zone"`
	Settings     *nosqlSettingsModel `tfsdk:"settings"`
	Remark       *nosqlRemarkModel   `tfsdk:"remark"`
	Instance     types.Object        `tfsdk:"instance"`
	Disk         types.Object        `tfsdk:"disk"`
	Interfaces   types.List          `tfsdk:"interfaces"`
	Availability types.String        `tfsdk:"availability"`
	Generation   types.Int32         `tfsdk:"generation"`
	CreatedAt    types.String        `tfsdk:"created_at"`
}

type nosqlSettingsModel struct {
	SourceNetwork    types.List   `tfsdk:"source_network"`
	ReserveIPAddress types.String `tfsdk:"reserve_ip_address"`
	Backup           types.Object `tfsdk:"backup"`
	Repair           types.Object `tfsdk:"repair"`
}

type nosqlBackupModel struct {
	Connect    types.String `tfsdk:"connect"`
	DaysOfWeek types.Set    `tfsdk:"days_of_week"`
	Time       types.String `tfsdk:"time"`
	Rotate     types.Int32  `tfsdk:"rotate"`
}

func (m nosqlBackupModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"connect":      types.StringType,
		"days_of_week": types.SetType{ElemType: types.StringType},
		"time":         types.StringType,
		"rotate":       types.Int32Type,
	}
}

type nosqlRepairModel struct {
	Incremental types.Object `tfsdk:"incremental"`
	Full        types.Object `tfsdk:"full"`
}

func (m nosqlRepairModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"incremental": types.ObjectType{AttrTypes: nosqlRepairIncrementalModel{}.AttributeTypes()},
		"full":        types.ObjectType{AttrTypes: nosqlRepairFullModel{}.AttributeTypes()},
	}
}

type nosqlRepairIncrementalModel struct {
	DaysOfWeek types.Set    `tfsdk:"days_of_week"`
	Time       types.String `tfsdk:"time"`
}

func (m nosqlRepairIncrementalModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"days_of_week": types.SetType{ElemType: types.StringType},
		"time":         types.StringType,
	}
}

type nosqlRepairFullModel struct {
	Interval  types.Int32  `tfsdk:"interval"`
	DayOfWeek types.String `tfsdk:"day_of_week"`
	Time      types.String `tfsdk:"time"`
}

func (m nosqlRepairFullModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interval":    types.Int32Type,
		"day_of_week": types.StringType,
		"time":        types.StringType,
	}
}

type nosqlRemarkModel struct {
	Nosql   types.Object             `tfsdk:"nosql"`
	Servers types.List               `tfsdk:"servers"`
	Network *nosqlRemarkNetworkModel `tfsdk:"network"`
	ZoneID  types.String             `tfsdk:"zone_id"`
}

type nosqlRemarkNosqlModel struct {
	PrimaryNodes types.Object `tfsdk:"primary_nodes"`
	Engine       types.String `tfsdk:"engine"`
	Version      types.String `tfsdk:"version"`
	DefaultUser  types.String `tfsdk:"default_user"`
	DiskSize     types.Int32  `tfsdk:"disk_size"`
	Memory       types.Int32  `tfsdk:"memory"`
	Nodes        types.Int32  `tfsdk:"nodes"`
	Port         types.Int32  `tfsdk:"port"`
	Virtualcore  types.Int32  `tfsdk:"virtualcore"`
	Zone         types.String `tfsdk:"zone"`
}

func (m nosqlRemarkNosqlModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"primary_nodes": types.ObjectType{AttrTypes: nosqlRemarkNosqlPrimaryNodesModel{}.AttributeTypes()},
		"engine":        types.StringType,
		"version":       types.StringType,
		"default_user":  types.StringType,
		"disk_size":     types.Int32Type,
		"memory":        types.Int32Type,
		"nodes":         types.Int32Type,
		"port":          types.Int32Type,
		"virtualcore":   types.Int32Type,
		"zone":          types.StringType,
	}
}

type nosqlRemarkNosqlPrimaryNodesModel struct {
	ID   types.String `tfsdk:"id"`
	Zone types.String `tfsdk:"zone"`
}

func (m nosqlRemarkNosqlPrimaryNodesModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"zone": types.StringType,
	}
}

type nosqlRemarkNetworkModel struct {
	Gateway types.String `tfsdk:"gateway"`
	Netmask types.Int32  `tfsdk:"netmask"`
}

func (m nosqlRemarkNetworkModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"gateway": types.StringType,
		"netmask": types.Int32Type,
	}
}

type nosqlInstanceModel struct {
	Status         types.String             `tfsdk:"status"`
	StatusChagedAt types.String             `tfsdk:"status_changed_at"`
	Host           *nosqlInstanceHostModel  `tfsdk:"host"`
	Hosts          []nosqlInstanceHostModel `tfsdk:"hosts"`
}

func (m nosqlInstanceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":            types.StringType,
		"status_changed_at": types.StringType,
		"host":              types.ObjectType{AttrTypes: nosqlInstanceHostModel{}.AttributeTypes()},
		"hosts":             types.ListType{ElemType: types.ObjectType{AttrTypes: nosqlInstanceHostModel{}.AttributeTypes()}},
	}
}

type nosqlInstanceHostModel struct {
	Name    types.String `tfsdk:"name"`
	InfoURL types.String `tfsdk:"info_url"`
}

func (m nosqlInstanceHostModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.StringType,
		"info_url": types.StringType,
	}
}

type nosqlDiskModel struct {
	EncryptionKey       *nosqlDiskEncryptionKeyModel `tfsdk:"encryption_key"`
	EncryptionAlgorithm types.String                 `tfsdk:"encryption_algorithm"`
}

func (m nosqlDiskModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"encryption_algorithm": types.StringType,
		"encryption_key":       types.ObjectType{AttrTypes: nosqlDiskEncryptionKeyModel{}.AttributeTypes()},
	}
}

type nosqlDiskEncryptionKeyModel struct {
	KMSKeyID types.String `tfsdk:"kms_key_id"`
}

func (m nosqlDiskEncryptionKeyModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kms_key_id": types.StringType,
	}
}

type nosqlInterfaceModel struct {
	IPAddress     types.String               `tfsdk:"ip_address"`
	UserIPAddress types.String               `tfsdk:"user_ip_address"`
	HostName      types.String               `tfsdk:"hostname"`
	VSwitch       *nosqlInterfaceSwitchModel `tfsdk:"vswitch"`
}

func (m nosqlInterfaceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_address":      types.StringType,
		"user_ip_address": types.StringType,
		"hostname":        types.StringType,
		"vswitch":         types.ObjectType{AttrTypes: nosqlInterfaceSwitchModel{}.AttributeTypes()},
	}
}

type nosqlInterfaceSwitchModel struct {
	ID         types.String                              `tfsdk:"id"`
	Name       types.String                              `tfsdk:"name"`
	Scope      types.String                              `tfsdk:"scope"`
	Subnet     *nosqlInterfaceModelSwitchSubnetModel     `tfsdk:"subnet"`
	UserSubnet *nosqlInterfaceModelSwitchUserSubnetModel `tfsdk:"user_subnet"`
}

func (m nosqlInterfaceSwitchModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"scope":       types.StringType,
		"subnet":      types.ObjectType{AttrTypes: nosqlInterfaceModelSwitchSubnetModel{}.AttributeTypes()},
		"user_subnet": types.ObjectType{AttrTypes: nosqlInterfaceModelSwitchUserSubnetModel{}.AttributeTypes()},
	}
}

type nosqlInterfaceModelSwitchSubnetModel struct {
	NetworkAddress types.String `tfsdk:"network_address"`
	Netmask        types.Int32  `tfsdk:"netmask"`
	Gateway        types.String `tfsdk:"gateway"`
	BandWidthMbps  types.Int32  `tfsdk:"band_width_mbps"`
}

func (m nosqlInterfaceModelSwitchSubnetModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"network_address": types.StringType,
		"netmask":         types.Int32Type,
		"gateway":         types.StringType,
		"band_width_mbps": types.Int32Type,
	}
}

type nosqlInterfaceModelSwitchUserSubnetModel struct {
	Gateway types.String `tfsdk:"gateway"`
	Netmask types.Int32  `tfsdk:"netmask"`
}

func (m nosqlInterfaceModelSwitchUserSubnetModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"gateway": types.StringType,
		"netmask": types.Int32Type,
	}
}

func (m *nosqlBaseModel) updateState(data *v1.GetNosqlAppliance) {
	m.UpdateBaseState(data.ID.Value, data.Name.Value, data.Description.Value, data.Tags.Value)
	m.Plan = types.StringValue(string(nosql.GetPlanFromID(data.Plan.Value.ID.Value)))
	m.Settings = flattenSettings(data)
	m.Remark = flattenRemark(data)
	m.Instance = flattenInstance(data)
	m.Disk = flattenDisk(data)
	m.Interfaces = flattenInterfaces(data)
	m.Availability = types.StringValue(string(data.Availability.Value))
	m.Generation = types.Int32Value(int32(data.Generation.Value))
	m.CreatedAt = types.StringValue(data.CreatedAt.Value.String())
	m.Zone = types.StringValue(data.Remark.Value.Nosql.Value.Zone.Value)
}

func flattenSettings(data *v1.GetNosqlAppliance) *nosqlSettingsModel {
	if !data.Settings.IsSet() {
		return nil
	}
	settings := data.Settings.Value
	model := &nosqlSettingsModel{
		SourceNetwork:    common.StringsToTlist(settings.SourceNetwork),
		ReserveIPAddress: types.StringValue(settings.ReserveIPAddress.Value.String()),
	}

	if backup, ok := settings.Backup.Get(); ok {
		m := &nosqlBackupModel{
			Connect:    types.StringValue(backup.Connect.Value),
			DaysOfWeek: common.StringsToTset(toStrs(backup.DayOfWeek.Value)),
			Time:       types.StringValue(backup.Time.Value),
			Rotate:     types.Int32Value(int32(backup.Rotate.Value)),
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if !diags.HasError() {
			model.Backup = value
		}
	} else {
		model.Backup = types.ObjectNull(nosqlBackupModel{}.AttributeTypes())
	}

	if repair, ok := settings.Repair.Get(); ok {
		repairModel := nosqlRepairModel{}
		if inc, ok := repair.Incremental.Get(); ok {
			m := &nosqlRepairIncrementalModel{
				DaysOfWeek: common.StringsToTset(toStrs(inc.DaysOfWeek)),
				Time:       types.StringValue(inc.Time.Value),
			}
			value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
			if !diags.HasError() {
				repairModel.Incremental = value
			} else {
				tflog.Error(context.Background(), diags[0].Detail())
			}
		} else {
			repairModel.Incremental = types.ObjectNull(nosqlRepairIncrementalModel{}.AttributeTypes())
		}
		if full, ok := repair.Full.Get(); ok {
			m := &nosqlRepairFullModel{
				Interval:  types.Int32Value(int32(full.Interval.Value)),
				DayOfWeek: types.StringValue(string(full.DayOfWeek.Value)),
				Time:      types.StringValue(full.Time.Value),
			}
			value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
			if !diags.HasError() {
				repairModel.Full = value
			}
		} else {
			repairModel.Full = types.ObjectNull(nosqlRepairFullModel{}.AttributeTypes())
		}
		value, diags := types.ObjectValueFrom(context.Background(), repairModel.AttributeTypes(), repairModel)
		if !diags.HasError() {
			model.Repair = value
		}
	} else {
		model.Repair = types.ObjectNull(nosqlRepairModel{}.AttributeTypes())
	}

	return model
}

func toStrs[S ~string](s []S) []string {
	if len(s) == 0 {
		return nil
	}

	t := make([]string, len(s))
	for i, v := range s {
		t[i] = string(v)
	}
	return t
}

func flattenRemark(data *v1.GetNosqlAppliance) *nosqlRemarkModel {
	if !data.Remark.IsSet() {
		return nil
	}

	remark := data.Remark.Value
	servers := make([]string, len(remark.Servers))
	for i, s := range remark.Servers {
		servers[i] = s.UserIPAddress.Value.String()
	}
	model := &nosqlRemarkModel{
		Servers: common.StringsToTlist(servers),
		Network: &nosqlRemarkNetworkModel{
			Gateway: types.StringValue(data.Remark.Value.Network.Value.DefaultRoute.Value),
			Netmask: types.Int32Value(int32(data.Remark.Value.Network.Value.NetworkMaskLen.Value)),
		},
		ZoneID: types.StringValue(remark.Zone.Value.ID.Value),
	}
	database := remark.Nosql.Value
	nosqlSettings := &nosqlRemarkNosqlModel{
		Engine:      types.StringValue(string(database.DatabaseEngine.Value)),
		Version:     types.StringValue(database.DatabaseVersion.Value),
		DefaultUser: types.StringValue(database.DefaultUser.Value),
		DiskSize:    types.Int32Value(int32(database.DiskSize.Value)),
		Memory:      types.Int32Value(int32(database.Memory.Value)),
		Nodes:       types.Int32Value(int32(database.Nodes.Value)),
		Port:        types.Int32Value(int32(database.Port.Value)),
		Virtualcore: types.Int32Value(int32(database.Virtualcore.Value)),
		Zone:        types.StringValue(database.Zone.Value),
	}
	if primaryNodes, ok := database.PrimaryNodes.Get(); ok {
		m := &nosqlRemarkNosqlPrimaryNodesModel{
			ID:   types.StringValue(primaryNodes.Appliance.Value.ID.Value),
			Zone: types.StringValue(primaryNodes.Appliance.Value.Zone.Value.Name.Value),
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if !diags.HasError() {
			nosqlSettings.PrimaryNodes = value
		}
	} else {
		nosqlSettings.PrimaryNodes = types.ObjectNull(nosqlRemarkNosqlPrimaryNodesModel{}.AttributeTypes())
	}
	value, diags := types.ObjectValueFrom(context.Background(), nosqlSettings.AttributeTypes(), nosqlSettings)
	if !diags.HasError() {
		model.Nosql = value
	}

	return model
}

func flattenInstance(data *v1.GetNosqlAppliance) types.Object {
	v := types.ObjectNull(nosqlInstanceModel{}.AttributeTypes())
	if !data.Instance.IsSet() {
		return v
	}

	instance := data.Instance.Value
	model := &nosqlInstanceModel{
		Status:         types.StringValue(instance.Status.Value),
		StatusChagedAt: types.StringValue(instance.StatusChangedAt.Value.String()),
	}

	if host, ok := instance.Host.Get(); ok {
		model.Host = &nosqlInstanceHostModel{
			Name:    types.StringValue(host.Name.Value),
			InfoURL: types.StringValue(host.InfoURL.Value),
		}
	}

	hosts := make([]nosqlInstanceHostModel, len(instance.Hosts))
	for i, h := range instance.Hosts {
		hosts[i] = nosqlInstanceHostModel{
			Name:    types.StringValue(h.Name.Value),
			InfoURL: types.StringValue(h.InfoURL.Value),
		}
	}
	model.Hosts = hosts

	value, diags := types.ObjectValueFrom(context.Background(), model.AttributeTypes(), model)
	if diags.HasError() {
		return v
	}

	return value
}

func flattenDisk(data *v1.GetNosqlAppliance) types.Object {
	v := types.ObjectNull(nosqlDiskModel{}.AttributeTypes())
	if disk, ok := data.Disk.Get(); ok {
		m := &nosqlDiskModel{
			EncryptionKey: &nosqlDiskEncryptionKeyModel{
				KMSKeyID: types.StringValue(disk.EncryptionKey.Value.KMSKeyID.Value),
			},
			EncryptionAlgorithm: types.StringValue(disk.EncryptionAlgorithm.Value),
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	} else {
		return v
	}
}

func flattenInterfaces(data *v1.GetNosqlAppliance) types.List {
	v := types.ListNull(types.ObjectType{AttrTypes: nosqlInterfaceModel{}.AttributeTypes()})
	if len(data.Interfaces) == 0 {
		return v
	}

	interfaces := make([]nosqlInterfaceModel, len(data.Interfaces))
	for i, iface := range data.Interfaces {
		if v, ok := iface.Get(); ok {
			interfaces[i] = nosqlInterfaceModel{
				IPAddress:     types.StringValue(v.IPAddress.Value),
				UserIPAddress: types.StringValue(v.UserIPAddress.Value),
				HostName:      types.StringValue(v.HostName.Value),
				VSwitch:       flattenSwitch(&v.Switch),
			}
		}
	}

	value, diags := types.ListValueFrom(context.Background(), types.ObjectType{AttrTypes: nosqlInterfaceModel{}.AttributeTypes()}, interfaces)
	if diags.HasError() {
		return v
	}

	return value
}

func flattenSwitch(switchInfo *v1.OptGetNosqlApplianceInterfacesItemSwitch) *nosqlInterfaceSwitchModel {
	if !switchInfo.IsSet() {
		return nil
	}

	sw := switchInfo.Value
	model := &nosqlInterfaceSwitchModel{
		ID:    types.StringValue(sw.ID.Value),
		Name:  types.StringValue(sw.Name.Value),
		Scope: types.StringValue(sw.Scope.Value),
	}

	if subnet, ok := sw.Subnet.Get(); ok {
		model.Subnet = &nosqlInterfaceModelSwitchSubnetModel{
			NetworkAddress: types.StringValue(subnet.NetworkAddress.Value),
			Netmask:        types.Int32Value(int32(subnet.NetworkMaskLen.Value)),
			Gateway:        types.StringValue(subnet.DefaultRoute.Value),
			BandWidthMbps:  types.Int32Value(int32(subnet.Internet.Value.BandWidthMbps.Value)),
		}
	}

	if userSubnet, ok := sw.UserSubnet.Get(); ok {
		model.UserSubnet = &nosqlInterfaceModelSwitchUserSubnetModel{
			Netmask: types.Int32Value(int32(userSubnet.NetworkMaskLen.Value)),
			Gateway: types.StringValue(userSubnet.DefaultRoute.Value),
		}
	}

	return model
}
