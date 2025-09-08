// Copyright 2016-2025 terraform-provider-sakura authors
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

package internet

import (
	"context"
	"fmt"
	"strings"

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
	SwitchID           types.String `tfsdk:"switch_id"`
	ServerIDs          types.Set    `tfsdk:"server_ids"`
	NetworkAddress     types.String `tfsdk:"network_address"`
	Gateway            types.String `tfsdk:"gateway"`
	MinIPAddress       types.String `tfsdk:"min_ip_address"`
	MaxIPAddress       types.String `tfsdk:"max_ip_address"`
	IPAddresses        types.Set    `tfsdk:"ip_addresses"`
	IPv6Prefix         types.String `tfsdk:"ipv6_prefix"`
	IPv6PrefixLen      types.Int32  `tfsdk:"ipv6_prefix_len"`
	IPv6NetworkAddress types.String `tfsdk:"ipv6_network_address"`
	AssignedTags       types.Set    `tfsdk:"assigned_tags"`
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
	assigned, unassigned := partitionInternetTags(data.Tags)

	model.UpdateBaseState(data.ID.String(), data.Name, data.Description, unassigned)
	model.Netmask = types.Int32Value(int32(data.NetworkMaskLen))
	model.BandWidth = types.Int32Value(int32(data.BandWidthMbps))
	model.SwitchID = types.StringValue(sw.ID.String())
	model.ServerIDs = common.StringsToTset(serverIDs)
	model.NetworkAddress = types.StringValue(sw.Subnets[0].NetworkAddress)
	model.Gateway = types.StringValue(sw.Subnets[0].DefaultRoute)
	model.MinIPAddress = types.StringValue(sw.Subnets[0].AssignedIPAddressMin)
	model.MaxIPAddress = types.StringValue(sw.Subnets[0].AssignedIPAddressMax)
	model.EnableIPv6 = types.BoolValue(enableIPv6)
	model.IPv6Prefix = types.StringValue(ipv6Prefix)
	model.IPv6PrefixLen = types.Int32Value(int32(ipv6PrefixLen))
	model.IPv6NetworkAddress = types.StringValue(ipv6NetworkAddress)
	model.Zone = types.StringValue(zone)
	model.IPAddresses = common.StringsToTset(sw.Subnets[0].GetAssignedIPAddresses())
	model.AssignedTags = common.StringsToTset(assigned)

	return nil
}

func partitionInternetTags(tags []string) (assigned, unassigned []string) {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@previous-id") {
			assigned = append(assigned, tag)
		} else {
			unassigned = append(unassigned, tag)
		}
	}
	return
}
