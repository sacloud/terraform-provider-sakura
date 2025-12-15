// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package internet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type internetBaseModel struct {
	common.SakuraBaseModel
	Zone               types.String `tfsdk:"zone"`
	IconID             types.String `tfsdk:"icon_id"`
	Netmask            types.Int32  `tfsdk:"netmask"`
	BandWidth          types.Int32  `tfsdk:"band_width"`
	EnableIPv6         types.Bool   `tfsdk:"enable_ipv6"`
	VSwitchID          types.String `tfsdk:"vswitch_id"`
	ServerIDs          types.List   `tfsdk:"server_ids"`
	NetworkAddress     types.String `tfsdk:"network_address"`
	Gateway            types.String `tfsdk:"gateway"`
	MinIPAddress       types.String `tfsdk:"min_ip_address"`
	MaxIPAddress       types.String `tfsdk:"max_ip_address"`
	IPAddresses        types.List   `tfsdk:"ip_addresses"`
	IPv6Prefix         types.String `tfsdk:"ipv6_prefix"`
	IPv6PrefixLen      types.Int32  `tfsdk:"ipv6_prefix_len"`
	IPv6NetworkAddress types.String `tfsdk:"ipv6_network_address"`
}

func (model *internetBaseModel) updateState(ctx context.Context, client *common.APIClient, zone string, data *iaas.Internet) error {
	swOp := iaas.NewSwitchOp(client)
	sw, err := swOp.Read(ctx, zone, data.Switch.ID)
	if err != nil {
		return fmt.Errorf("could not read SakuraCloud Switch[%s]: %s", data.Switch.ID, err)
	}

	var serverIDs []string
	if sw.ServerCount > 0 {
		servers, err := swOp.GetServers(ctx, zone, sw.ID)
		if err != nil {
			return fmt.Errorf("could not find SakuraCloud Servers of Switch[%s]: %s", sw.ID.String(), err)
		}
		for _, s := range servers.Servers {
			serverIDs = append(serverIDs, s.ID.String())
		}
	}

	var enableIPv6 bool
	var ipv6Prefix, ipv6NetworkAddress string
	var ipv6PrefixLen int
	if len(data.Switch.IPv6Nets) > 0 {
		enableIPv6 = true
		ipv6Prefix = data.Switch.IPv6Nets[0].IPv6Prefix
		ipv6PrefixLen = data.Switch.IPv6Nets[0].IPv6PrefixLen
		ipv6NetworkAddress = fmt.Sprintf("%s/%d", ipv6Prefix, ipv6PrefixLen)
	}
	model.UpdateBaseState(data.ID.String(), data.Name, data.Description, data.Tags)
	model.Netmask = types.Int32Value(int32(data.NetworkMaskLen))
	model.BandWidth = types.Int32Value(int32(data.BandWidthMbps))
	model.VSwitchID = types.StringValue(sw.ID.String())
	model.ServerIDs = common.StringsToTlist(serverIDs)
	model.NetworkAddress = types.StringValue(sw.Subnets[0].NetworkAddress)
	model.Gateway = types.StringValue(sw.Subnets[0].DefaultRoute)
	model.MinIPAddress = types.StringValue(sw.Subnets[0].AssignedIPAddressMin)
	model.MaxIPAddress = types.StringValue(sw.Subnets[0].AssignedIPAddressMax)
	model.EnableIPv6 = types.BoolValue(enableIPv6)
	model.IPv6Prefix = types.StringValue(ipv6Prefix)
	model.IPv6PrefixLen = types.Int32Value(int32(ipv6PrefixLen))
	model.IPv6NetworkAddress = types.StringValue(ipv6NetworkAddress)
	model.Zone = types.StringValue(zone)
	model.IPAddresses = common.StringsToTlist(sw.Subnets[0].GetAssignedIPAddresses())
	if data.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(data.IconID.String())
	}

	return nil
}
