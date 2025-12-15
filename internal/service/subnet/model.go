// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package subnet

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	iaas "github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type subnetBaseModel struct {
	ID             types.String `tfsdk:"id"`
	InternetID     types.String `tfsdk:"internet_id"`
	VSwitchID      types.String `tfsdk:"vswitch_id"`
	Netmask        types.Int32  `tfsdk:"netmask"`
	NextHop        types.String `tfsdk:"next_hop"`
	NetworkAddress types.String `tfsdk:"network_address"`
	MinIPAddress   types.String `tfsdk:"min_ip_address"`
	MaxIPAddress   types.String `tfsdk:"max_ip_address"`
	IPAddresses    types.List   `tfsdk:"ip_addresses"`
	Zone           types.String `tfsdk:"zone"`
}

func (m *subnetBaseModel) updateState(zone string, data *iaas.Subnet) error {
	if data.SwitchID.IsEmpty() {
		return fmt.Errorf("failed to read Subnet[%s]: %s", data.ID, "switch is nil")
	}
	if data.InternetID.IsEmpty() {
		return fmt.Errorf("failed to read Subnet[%s]: %s", data.ID, "internet is nil")
	}

	var addrs []string
	for _, ip := range data.IPAddresses {
		addrs = append(addrs, ip.IPAddress)
	}

	m.ID = types.StringValue(data.ID.String())
	m.InternetID = types.StringValue(data.InternetID.String())
	m.VSwitchID = types.StringValue(data.SwitchID.String())
	m.Netmask = types.Int32Value(int32(data.NetworkMaskLen))
	m.NextHop = types.StringValue(data.NextHop)
	m.NetworkAddress = types.StringValue(data.NetworkAddress)
	m.MinIPAddress = types.StringValue(data.IPAddresses[0].IPAddress)
	m.MaxIPAddress = types.StringValue(data.IPAddresses[len(data.IPAddresses)-1].IPAddress)
	m.IPAddresses = common.StringsToTlist(addrs)
	m.Zone = types.StringValue(zone)

	return nil
}
