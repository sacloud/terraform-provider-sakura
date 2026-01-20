// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package nlb

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type nlbBaseModel struct {
	common.SakuraBaseModel
	Zone             types.String              `tfsdk:"zone"`
	IconID           types.String              `tfsdk:"icon_id"`
	Plan             types.String              `tfsdk:"plan"`
	NetworkInterface *nlbNetworkInterfaceModel `tfsdk:"network_interface"`
	VIP              []nlbVIPModel             `tfsdk:"vip"`
}

type nlbNetworkInterfaceModel struct {
	VSwitchID   types.String `tfsdk:"vswitch_id"`
	VRID        types.Int64  `tfsdk:"vrid"`
	IPAddresses types.List   `tfsdk:"ip_addresses"`
	Netmask     types.Int32  `tfsdk:"netmask"`
	Gateway     types.String `tfsdk:"gateway"`
}

type nlbServerModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	Protocol  types.String `tfsdk:"protocol"`
	Path      types.String `tfsdk:"path"`
	Status    types.Int32  `tfsdk:"status"`
	Enabled   types.Bool   `tfsdk:"enabled"`
}

type nlbVIPModel struct {
	VIP         types.String     `tfsdk:"vip"`
	Port        types.Int32      `tfsdk:"port"`
	DelayLoop   types.Int32      `tfsdk:"delay_loop"`
	SorryServer types.String     `tfsdk:"sorry_server"`
	Description types.String     `tfsdk:"description"`
	Server      []nlbServerModel `tfsdk:"server"`
}

func (model *nlbBaseModel) updateState(lb *iaas.LoadBalancer, zone string) {
	model.UpdateBaseState(lb.ID.String(), lb.Name, lb.Description, lb.Tags)
	model.Zone = types.StringValue(zone)
	if lb.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(lb.IconID.String())
	}
	model.Plan = types.StringValue(flattenLoadBalancerPlanID(lb))
	model.NetworkInterface = &nlbNetworkInterfaceModel{
		VSwitchID:   types.StringValue(lb.SwitchID.String()),
		VRID:        types.Int64Value(int64(lb.VRID)),
		IPAddresses: common.StringsToTlist(lb.IPAddresses),
		Netmask:     types.Int32Value(int32(lb.NetworkMaskLen)),
	}
	if lb.DefaultRoute != "" {
		model.NetworkInterface.Gateway = types.StringValue(lb.DefaultRoute)
	} else {
		model.NetworkInterface.Gateway = types.StringNull()
	}
	model.VIP = flattenVIPs(lb)
}

func flattenLoadBalancerPlanID(lb *iaas.LoadBalancer) string {
	switch lb.PlanID {
	case iaastypes.LoadBalancerPlans.Standard:
		return "standard"
	case iaastypes.LoadBalancerPlans.HighSpec:
		return "highspec"
	}
	return ""
}

func flattenVIPs(lb *iaas.LoadBalancer) []nlbVIPModel {
	if lb == nil || len(lb.VirtualIPAddresses) == 0 {
		return nil
	}

	var results []nlbVIPModel
	for _, vip := range lb.VirtualIPAddresses {
		var servers []nlbServerModel
		for _, s := range vip.Servers {
			server := nlbServerModel{
				IPAddress: types.StringValue(s.IPAddress),
				Protocol:  types.StringValue(string(s.HealthCheck.Protocol)),
				Enabled:   types.BoolValue(s.Enabled.Bool()),
			}
			// http/https以外ではPath/Statusは空になる
			if s.HealthCheck.Path == "" {
				server.Path = types.StringNull()
			} else {
				server.Path = types.StringValue(s.HealthCheck.Path)
			}
			if s.HealthCheck.ResponseCode.Int() == 0 {
				server.Status = types.Int32Null()
			} else {
				server.Status = types.Int32Value(int32(s.HealthCheck.ResponseCode))
			}
			servers = append(servers, server)
		}
		vipModel := nlbVIPModel{
			VIP:         types.StringValue(vip.VirtualIPAddress),
			Port:        types.Int32Value(int32(vip.Port)),
			DelayLoop:   types.Int32Value(int32(vip.DelayLoop)),
			Description: types.StringValue(vip.Description),
			Server:      servers,
		}
		if vip.SorryServer == "" {
			vipModel.SorryServer = types.StringNull()
		} else {
			vipModel.SorryServer = types.StringValue(vip.SorryServer)
		}
		results = append(results, vipModel)
	}
	return results
}
