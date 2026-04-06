// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	lb "github.com/sacloud/apprun-dedicated-api-go/apis/loadbalancer"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type lbID = v1.LoadBalancerID

type lbAddrModel struct {
	Address types.String `tfsdk:"address"`
	Vip     types.Bool   `tfsdk:"vip"`
}

type lbifAddrModel struct {
	InterfaceIndex types.Int32   `tfsdk:"interface_index"`
	Addresses      []lbAddrModel `tfsdk:"addresses"`
}

type lbNodeModel struct {
	ID                 types.String    `tfsdk:"id"`
	ResourceID         types.String    `tfsdk:"resource_id"`
	Status             types.String    `tfsdk:"status"`
	ArchiveVersion     types.String    `tfsdk:"archive_version"`
	CreateErrorMessage types.String    `tfsdk:"create_error_message"`
	Created            types.String    `tfsdk:"created"`
	Interfaces         []lbifAddrModel `tfsdk:"interfaces"`
}

type lbifModel struct {
	InterfaceIndex  types.Int32  `tfsdk:"interface_index"`
	Upstream        types.String `tfsdk:"upstream"`
	IpPool          []rangeModel `tfsdk:"ip_pool"`
	NetmaskLen      types.Int32  `tfsdk:"netmask"`
	DefaultGateway  types.String `tfsdk:"default_gateway"`
	Vip             types.String `tfsdk:"vip"`
	VirtualRouterID types.Int32  `tfsdk:"virtual_router_id"`
	PacketFilterID  types.String `tfsdk:"packet_filter_id"`
}

type lbModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ServiceClassPath types.String `tfsdk:"service_class_path"`
	NameServers      types.List   `tfsdk:"name_servers"`
	Interfaces       []lbifModel  `tfsdk:"interfaces"`
	Created          types.String `tfsdk:"created"`
	Deleting         types.Bool   `tfsdk:"deleting"`
}

var lbAddrAttrs = attrTypes{
	"address": types.StringType,
	"vip":     types.BoolType,
}

var lbifAddrAttrs = attrTypes{
	"interface_index": types.Int32Type,
	"addresses":       types.SetType{ElemType: types.ObjectType{AttrTypes: lbAddrAttrs}},
}

var lbNodeAttrs = attrTypes{
	"id":                   types.StringType,
	"resource_id":          types.StringType,
	"status":               types.StringType,
	"archive_version":      types.StringType,
	"create_error_message": types.StringType,
	"created":              types.StringType,
	"interfaces":           types.ListType{ElemType: types.ObjectType{AttrTypes: lbifAddrAttrs}},
}

var lbifAttrs = attrTypes{
	"interface_index":   types.Int32Type,
	"upstream":          types.StringType,
	"ip_pool":           types.SetType{ElemType: types.ObjectType{AttrTypes: rangeAttrs}},
	"netmask":           types.Int32Type,
	"default_gateway":   types.StringType,
	"vip":               types.StringType,
	"virtual_router_id": types.Int32Type,
	"packet_filter_id":  types.StringType,
}

var lbAttrs = attrTypes{
	"id":                 types.StringType,
	"name":               types.StringType,
	"service_class_path": types.StringType,
	"name_servers":       types.ListType{ElemType: types.StringType},
	"interfaces":         types.SetType{ElemType: types.ObjectType{AttrTypes: lbifAttrs}},
	"created":            types.StringType,
	"deleting":           types.BoolType,
}

func (lbAddrModel) AttributeTypes() attrTypes   { return lbAddrAttrs }
func (lbifAddrModel) AttributeTypes() attrTypes { return lbifAddrAttrs }
func (lbNodeModel) AttributeTypes() attrTypes   { return lbNodeAttrs }
func (lbifModel) AttributeTypes() attrTypes     { return lbifAttrs }
func (lbModel) AttributeTypes() attrTypes       { return lbAttrs }
func (m *lbModel) lbID() (lbID, error)          { return intoUUID[lbID](m.ID) }

func (a *lbAddrModel) updateState(src lb.NodeInterfaceAddress) {
	a.Address = types.StringValue(src.Address)
	a.Vip = types.BoolValue(src.Vip)
}

func (i *lbifAddrModel) updateState(src lb.NodeInterface) {
	i.InterfaceIndex = types.Int32Value(common.ToInt32(src.InterfaceIndex))
	i.Addresses = common.MapTo(src.Addresses, stateUpdater[lb.NodeInterfaceAddress, lbAddrModel])
}

func (n *lbNodeModel) updateState(detail lb.LoadBalancerNodeDetail) {
	n.ID = uuid2StringValue(detail.LoadBalancerNodeID)
	n.ResourceID = types.StringPointerValue(detail.ResourceID)
	n.Status = types.StringValue(common.ToString(detail.Status))
	n.ArchiveVersion = types.StringPointerValue(detail.ArchiveVersion)
	n.CreateErrorMessage = types.StringPointerValue(detail.CreateErrorMessage)
	n.Created = intoRFC2822(detail.Created)
	n.Interfaces = common.MapTo(detail.Interfaces, stateUpdater[lb.NodeInterface, lbifAddrModel])
}

func (i *lbifModel) updateState(src lb.LoadBalancerInterface) {
	i.InterfaceIndex = types.Int32Value(common.ToInt32(src.InterfaceIndex))
	i.Upstream = types.StringValue(src.Upstream)
	i.NetmaskLen = intoInt32(src.NetmaskLen)
	i.DefaultGateway = types.StringPointerValue(src.DefaultGateway)
	i.Vip = types.StringPointerValue(src.Vip)
	i.VirtualRouterID = intoInt32(src.VirtualRouterID)
	i.PacketFilterID = types.StringPointerValue(src.PacketFilterID)
	i.IpPool = common.MapTo(src.IpPool, stateUpdater[v1.IpRange, rangeModel])
}

func (m *lbModel) updateState(ctx context.Context, d *lb.LoadBalancerDetail) (ret diag.Diagnostics) {
	m.ID = uuid2StringValue(d.LoadBalancerID)
	m.Name = types.StringValue(d.Name)
	m.ServiceClassPath = types.StringValue(d.ServiceClassPath)
	m.Created = intoRFC2822(d.Created)
	m.Deleting = types.BoolValue(d.Deleting)
	m.Interfaces = common.MapTo(d.Interfaces, stateUpdater[lb.LoadBalancerInterface, lbifModel])
	m.NameServers, ret = types.ListValueFrom(ctx, types.StringType, common.MapTo(d.NameServers, common.ToString))

	return
}

func (i *lbifModel) intoCreate() (ret lb.LoadBalancerInterface, diag diag.Diagnostics) {
	ret.Upstream = i.Upstream.ValueString()
	ret.IpPool = common.MapTo(i.IpPool, rangeModel.intoCreate)
	ret.DefaultGateway = i.DefaultGateway.ValueStringPointer()
	ret.Vip = i.Vip.ValueStringPointer()
	ret.PacketFilterID = i.PacketFilterID.ValueStringPointer()

	n, d := intoInt16(saclient.Ptr(i.InterfaceIndex.ValueInt32()))
	diag.Append(d...)
	if n != nil {
		ret.InterfaceIndex = *n
	}

	n, d = intoInt16(i.NetmaskLen.ValueInt32Pointer())
	diag.Append(d...)
	ret.NetmaskLen = n

	n, d = intoInt16(i.VirtualRouterID.ValueInt32Pointer())
	diag.Append(d...)
	ret.VirtualRouterID = n

	return
}

func (m lbModel) intoCreate() (ret lb.CreateParams, diag diag.Diagnostics) {
	ret.Name = m.Name.ValueString()
	ret.ServiceClassPath = m.ServiceClassPath.ValueString()
	ret.NameServers = common.MapTo(common.TlistToStrings(m.NameServers), func(i string) v1.IPv4 { return v1.IPv4(i) })
	ret.Interfaces = common.MapTo(m.Interfaces, func(i lbifModel) lb.LoadBalancerInterface {
		j, k := i.intoCreate()
		diag.Append(k...)
		return j
	})

	return
}
