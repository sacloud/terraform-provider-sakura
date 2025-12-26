// Copyright 2016-2025 The sacloud/terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package local_router

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type localRouterBaseModel struct {
	common.SakuraBaseModel
	IconID           types.String                      `tfsdk:"icon_id"`
	Switch           *localRouterSwitchModel           `tfsdk:"switch"`
	NetworkInterface *localRouterNetworkInterfaceModel `tfsdk:"network_interface"`
	Peer             []localRouterPeerModel            `tfsdk:"peer"`
	StaticRoute      []localRouterStaticRouteModel     `tfsdk:"static_route"`
	SecretKeys       types.List                        `tfsdk:"secret_keys"`
}

type localRouterSwitchModel struct {
	Code     types.String `tfsdk:"code"`
	Category types.String `tfsdk:"category"`
	Zone     types.String `tfsdk:"zone"`
}

type localRouterNetworkInterfaceModel struct {
	VIP         types.String `tfsdk:"vip"`
	IPAddresses types.List   `tfsdk:"ip_addresses"`
	Netmask     types.Int32  `tfsdk:"netmask"`
	VRID        types.Int64  `tfsdk:"vrid"`
}

type localRouterPeerModel struct {
	PeerID      types.String `tfsdk:"peer_id"`
	SecretKey   types.String `tfsdk:"secret_key"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
}

type localRouterStaticRouteModel struct {
	Prefix  types.String `tfsdk:"prefix"`
	NextHop types.String `tfsdk:"next_hop"`
}

func (model *localRouterBaseModel) updateState(lr *iaas.LocalRouter) {
	model.UpdateBaseState(lr.ID.String(), lr.Name, lr.Description, lr.Tags)
	if lr.IconID != iaastypes.ID(0) {
		model.IconID = types.StringValue(lr.IconID.String())
	} else {
		model.IconID = types.StringNull()
	}

	if lr.Switch != nil {
		model.Switch = &localRouterSwitchModel{
			Code:     types.StringValue(lr.Switch.Code),
			Category: types.StringValue(lr.Switch.Category),
			Zone:     types.StringValue(lr.Switch.ZoneID),
		}
	}

	if lr.Interface != nil {
		model.NetworkInterface = &localRouterNetworkInterfaceModel{
			VIP:         types.StringValue(lr.Interface.VirtualIPAddress),
			IPAddresses: common.StringsToTlist(lr.Interface.IPAddress),
			Netmask:     types.Int32Value(int32(lr.Interface.NetworkMaskLen)),
			VRID:        types.Int64Value(int64(lr.Interface.VRID)),
		}
	}

	if len(lr.Peers) > 0 {
		var peers []localRouterPeerModel
		for _, p := range lr.Peers {
			peers = append(peers, localRouterPeerModel{
				PeerID:      types.StringValue(p.ID.String()),
				SecretKey:   types.StringValue(p.SecretKey),
				Enabled:     types.BoolValue(p.Enabled),
				Description: types.StringValue(p.Description),
			})
		}
		model.Peer = peers
	}

	if len(lr.StaticRoutes) > 0 {
		var routes []localRouterStaticRouteModel
		for _, r := range lr.StaticRoutes {
			routes = append(routes, localRouterStaticRouteModel{
				Prefix:  types.StringValue(r.Prefix),
				NextHop: types.StringValue(r.NextHop),
			})
		}
		model.StaticRoute = routes
	}

	model.SecretKeys = common.StringsToTlist(lr.SecretKeys)
}
