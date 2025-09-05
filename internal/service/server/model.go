// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

type serverBaseModel struct {
	common.SakuraBaseModel
	IconID           types.String                  `tfsdk:"icon_id"`
	Zone             types.String                  `tfsdk:"zone"`
	Core             types.Int64                   `tfsdk:"core"`
	Memory           types.Int64                   `tfsdk:"memory"`
	GPU              types.Int64                   `tfsdk:"gpu"`
	CPUModel         types.String                  `tfsdk:"cpu_model"`
	Commitment       types.String                  `tfsdk:"commitment"`
	Disks            types.Set                     `tfsdk:"disks"`
	InterfaceDriver  types.String                  `tfsdk:"interface_driver"`
	NetworkInterface []serverNetworkInterfaceModel `tfsdk:"network_interface"`
	CDROMID          types.String                  `tfsdk:"cdrom_id"`
	PrivateHostID    types.String                  `tfsdk:"private_host_id"`
	PrivateHostName  types.String                  `tfsdk:"private_host_name"`
	IPAddress        types.String                  `tfsdk:"ip_address"` // iptypes.IPAddress `tfsdk:"ip_address"`
	Gateway          types.String                  `tfsdk:"gateway"`
	NetworkAddress   types.String                  `tfsdk:"network_address"`
	Netmask          types.Int32                   `tfsdk:"netmask"`
	Hostname         types.String                  `tfsdk:"hostname"`
	DNSServers       types.Set                     `tfsdk:"dns_servers"`
}

type serverNetworkInterfaceModel struct {
	Upstream       types.String `tfsdk:"upstream"`
	UserIPAddress  types.String `tfsdk:"user_ip_address"` // iptypes.IPv4Address `tfsdk:"user_ip_address"`
	PacketFilterID types.String `tfsdk:"packet_filter_id"`
	MACAddress     types.String `tfsdk:"mac_address"`
}

func (model *serverBaseModel) updateState(server *iaas.Server, zone string) {
	// FrameworkはConnInfo周りの機能を提供しないため削除
	ip, gateway, nwMaskLen, nwAddress := flattenServerNetworkInfo(server)
	model.UpdateBaseState(server.ID.String(), server.Name, server.Description, server.Tags)
	model.Core = types.Int64Value(int64(server.CPU))
	model.Memory = types.Int64Value(int64(server.GetMemoryGB()))
	model.CPUModel = types.StringValue(server.ServerPlanCPUModel)
	model.Commitment = types.StringValue(server.ServerPlanCommitment.String())
	model.Disks = common.StringsToTset(flattenServerConnectedDiskIDs(server))
	model.InterfaceDriver = types.StringValue(server.InterfaceDriver.String())
	model.PrivateHostName = types.StringValue(server.PrivateHostName)
	model.NetworkInterface = flattenServerNICs(server)
	model.IPAddress = types.StringValue(ip) // iptypes.NewIPAddressValue(ip)
	model.Gateway = types.StringValue(gateway)
	model.NetworkAddress = types.StringValue(nwAddress)
	model.Netmask = types.Int32Value(int32(nwMaskLen))
	model.Hostname = types.StringValue(server.HostName)
	model.DNSServers = common.StringsToTset(server.Zone.Region.NameServers)
	model.Zone = types.StringValue(zone)
}

func flattenServerNICs(server *iaas.Server) []serverNetworkInterfaceModel {
	var results []serverNetworkInterfaceModel
	for _, nic := range server.Interfaces {
		var upstream string
		switch {
		case nic.SwitchID.IsEmpty():
			upstream = "disconnect"
		case nic.SwitchScope == iaastypes.Scopes.Shared:
			upstream = "shared"
		default:
			upstream = nic.SwitchID.String()
		}
		results = append(results, serverNetworkInterfaceModel{
			Upstream:       types.StringValue(upstream),
			PacketFilterID: types.StringValue(nic.PacketFilterID.String()),
			MACAddress:     types.StringValue(strings.ToLower(nic.MACAddress)),
			UserIPAddress:  types.StringValue(nic.UserIPAddress), // iptypes.NewIPv4AddressValue(nic.UserIPAddress),
		})
	}
	return results
}

func flattenServerConnectedDiskIDs(server *iaas.Server) []string {
	var ids []string
	for _, disk := range server.Disks {
		ids = append(ids, disk.ID.String())
	}
	return ids
}

func flattenServerNetworkInfo(server *iaas.Server) (ip, gateway string, nwMaskLen int, nwAddress string) {
	if len(server.Interfaces) > 0 && !server.Interfaces[0].SwitchID.IsEmpty() {
		nic := server.Interfaces[0]
		if nic.SwitchScope == iaastypes.Scopes.Shared {
			ip = nic.IPAddress
		} else {
			ip = nic.UserIPAddress
		}
		gateway = nic.UserSubnetDefaultRoute
		nwMaskLen = nic.UserSubnetNetworkMaskLen
		nwAddress = nic.SubnetNetworkAddress // null if connected switch(not router)
	}
	return
}
