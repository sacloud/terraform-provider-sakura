// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	asg "github.com/sacloud/apprun-dedicated-api-go/apis/autoscalinggroup"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type asgID = v1.AutoScalingGroupID

type rangeModel struct {
	Start types.String `tfsdk:"start"`
	End   types.String `tfsdk:"end"`
}

type nodeInterfaceModel struct {
	InterfaceIndex types.Int32  `tfsdk:"interface_index"`
	Upstream       types.String `tfsdk:"upstream"`
	IpPool         []rangeModel `tfsdk:"ip_pool"`
	NetmaskLen     types.Int32  `tfsdk:"netmask"`
	DefaultGateway types.String `tfsdk:"default_gateway"`
	PacketFilterID types.String `tfsdk:"packet_filter_id"`
	ConnectsToLB   types.Bool   `tfsdk:"connects_to_lb"`
}

type asgModel struct {
	ID                     types.String         `tfsdk:"id"`
	ClusterID              types.String         `tfsdk:"cluster_id"`
	Name                   types.String         `tfsdk:"name"`
	Zone                   types.String         `tfsdk:"zone"`
	NameServers            types.List           `tfsdk:"name_servers"`
	WorkerServiceClassPath types.String         `tfsdk:"worker_service_class_path"`
	MinNodes               types.Int32          `tfsdk:"min_nodes"`
	MaxNodes               types.Int32          `tfsdk:"max_nodes"`
	CurrentNodes           types.Int32          `tfsdk:"current_nodes"`
	Deleting               types.Bool           `tfsdk:"deleting"`
	Interfaces             []nodeInterfaceModel `tfsdk:"interfaces"`
}

var rangeAttrs = attrTypes{
	"start": types.StringType,
	"end":   types.StringType,
}

var asgInterfaceAttrs = attrTypes{
	"interface_index":  types.Int32Type,
	"upstream":         types.StringType,
	"ip_pool":          types.SetType{ElemType: types.ObjectType{AttrTypes: rangeAttrs}},
	"netmask":          types.Int32Type,
	"default_gateway":  types.StringType,
	"packet_filter_id": types.StringType,
	"connects_to_lb":   types.BoolType,
}

var asgAttrs = attrTypes{
	"id":                        types.StringType,
	"name":                      types.StringType,
	"zone":                      types.StringType,
	"name_servers":              types.ListType{ElemType: types.StringType},
	"worker_service_class_path": types.StringType,
	"min_nodes":                 types.Int32Type,
	"max_nodes":                 types.Int32Type,
	"current_nodes":             types.Int32Type,
	"deleting":                  types.BoolType,
	"interfaces":                types.SetType{ElemType: types.ObjectType{AttrTypes: asgInterfaceAttrs}},
}

func (rangeModel) AttributeTypes() attrTypes         { return rangeAttrs }
func (nodeInterfaceModel) AttributeTypes() attrTypes { return asgInterfaceAttrs }
func (asgModel) AttributeTypes() attrTypes           { return asgAttrs }
func (a *asgModel) clusterID() (clusterID, error)    { return intoUUID[clusterID](a.ClusterID) }

func (i rangeModel) intoCreate() (ret v1.IpRange) {
	ret.Start = v1.IPv4(i.Start.ValueString())
	ret.End = v1.IPv4(i.End.ValueString())
	return
}

func (i nodeInterfaceModel) intoCreate() (ret asg.NodeInterface, diag diag.Diagnostics) {
	ret.Upstream = i.Upstream.ValueString()
	ret.IpPool = common.MapTo(i.IpPool, rangeModel.intoCreate)
	ret.DefaultGateway = i.DefaultGateway.ValueStringPointer()
	ret.PacketFilterID = i.PacketFilterID.ValueStringPointer()
	ret.ConnectsToLB = i.ConnectsToLB.ValueBool()

	n, d := intoInt16(i.InterfaceIndex.ValueInt32Pointer())
	diag.Append(d...)
	if n != nil {
		ret.InterfaceIndex = *n
	}

	n, d = intoInt16(i.NetmaskLen.ValueInt32Pointer())
	diag.Append(d...)
	ret.NetmaskLen = n

	return
}

func (a *asgModel) intoCreate() (ret asg.CreateParams, diag diag.Diagnostics) {
	ret.Name = a.Name.ValueString()
	ret.Zone = a.Zone.ValueString()
	ret.NameServers = common.MapTo(common.TlistToStrings(a.NameServers), func(s string) v1.IPv4 {
		return v1.IPv4(s)
	})
	ret.WorkerServiceClassPath = a.WorkerServiceClassPath.ValueString()
	ret.MinNodes = a.MinNodes.ValueInt32()
	ret.MaxNodes = a.MaxNodes.ValueInt32()
	ret.Interfaces = common.MapTo(a.Interfaces, func(i nodeInterfaceModel) asg.NodeInterface {
		j, d := i.intoCreate()
		diag.Append(d...)
		return j
	})

	return
}

func (i *rangeModel) updateState(ip v1.IpRange) {
	i.Start = types.StringValue(common.ToString(ip.Start))
	i.End = types.StringValue(common.ToString(ip.End))
}

func (i *nodeInterfaceModel) updateState(d asg.NodeInterface) {
	i.InterfaceIndex = types.Int32Value(common.ToInt32(d.InterfaceIndex))
	i.Upstream = types.StringValue(d.Upstream)
	i.NetmaskLen = intoInt32(d.NetmaskLen)
	i.DefaultGateway = types.StringPointerValue(d.DefaultGateway)
	i.PacketFilterID = types.StringPointerValue(d.PacketFilterID)
	i.ConnectsToLB = types.BoolValue(d.ConnectsToLB)
	i.IpPool = common.MapTo(d.IpPool, stateUpdater[v1.IpRange, rangeModel])
}

func (a *asgModel) updateState(ctx context.Context, d *asg.AutoScalingGroupDetail, cid clusterID) (ret diag.Diagnostics) {
	a.ID = uuid2StringValue(d.AutoScalingGroupID)
	a.ClusterID = uuid2StringValue(cid)
	a.Name = types.StringValue(d.Name)
	a.Zone = types.StringValue(d.Zone)
	a.WorkerServiceClassPath = types.StringValue(d.WorkerServiceClassPath)
	a.MinNodes = types.Int32Value(d.MinNodes)
	a.MaxNodes = types.Int32Value(d.MaxNodes)
	a.CurrentNodes = types.Int32Value(d.CurrentNodes)
	a.Deleting = types.BoolValue(d.Deleting)
	a.Interfaces = common.MapTo(d.Interfaces, stateUpdater[asg.NodeInterface, nodeInterfaceModel])
	a.NameServers, ret = types.ListValueFrom(ctx, types.StringType, common.MapTo(d.NameServers, common.ToString))
	return
}
