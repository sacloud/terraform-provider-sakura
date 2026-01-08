// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package vpn_router

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type vpnRouterBaseModel struct {
	common.SakuraBaseModel
	IconID                  types.String                            `tfsdk:"icon_id"`
	Zone                    types.String                            `tfsdk:"zone"`
	Plan                    types.String                            `tfsdk:"plan"`
	Version                 types.Int32                             `tfsdk:"version"`
	PublicNetworkInterface  types.Object                            `tfsdk:"public_network_interface"`
	PublicIP                types.String                            `tfsdk:"public_ip"`
	PublicNetmask           types.Int64                             `tfsdk:"public_netmask"`
	SyslogHost              types.String                            `tfsdk:"syslog_host"`
	InternetConnection      types.Bool                              `tfsdk:"internet_connection"`
	PrivateNetworkInterface []vpnRouterPrivateNetworkInterfaceModel `tfsdk:"private_network_interface"`
	DHCPServer              []vpnRouterDHCPServerModel              `tfsdk:"dhcp_server"`
	DHCPStaticMapping       []vpnRouterDHCPStaticMappingModel       `tfsdk:"dhcp_static_mapping"`
	DNSForwarding           types.Object                            `tfsdk:"dns_forwarding"`
	Firewall                []vpnRouterFirewallModel                `tfsdk:"firewall"`
	L2TP                    types.Object                            `tfsdk:"l2tp"`
	PortForwarding          []vpnRouterPortForwardingModel          `tfsdk:"port_forwarding"`
	PPTP                    types.Object                            `tfsdk:"pptp"`
	WireGuard               types.Object                            `tfsdk:"wire_guard"`
	SiteToSiteVPN           []vpnRouterSiteToSiteVPNModel           `tfsdk:"site_to_site_vpn"`
	SiteToSiteVPNParameter  types.Object                            `tfsdk:"site_to_site_vpn_parameter"`
	StaticNAT               []vpnRouterStaticNATModel               `tfsdk:"static_nat"`
	StaticRoute             []vpnRouterStaticRouteModel             `tfsdk:"static_route"`
	ScheduledMaintenance    types.Object                            `tfsdk:"scheduled_maintenance"`
	//User                    []vpnRouterUserModel                    `tfsdk:"user"`
	MonitoringSuite types.Object `tfsdk:"monitoring_suite"`
}

type vpnRouterPublicNetworkInterfaceModel struct {
	VSwitchID   types.String `tfsdk:"vswitch_id"`
	VIP         types.String `tfsdk:"vip"`
	IPAddresses types.List   `tfsdk:"ip_addresses"`
	VRID        types.Int64  `tfsdk:"vrid"`
	Aliases     types.List   `tfsdk:"aliases"`
}

func (m vpnRouterPublicNetworkInterfaceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"vswitch_id":   types.StringType,
		"vip":          types.StringType,
		"ip_addresses": types.ListType{ElemType: types.StringType},
		"vrid":         types.Int64Type,
		"aliases":      types.ListType{ElemType: types.StringType},
	}
}

type vpnRouterPrivateNetworkInterfaceModel struct {
	Index       types.Int32  `tfsdk:"index"`
	VSwitchID   types.String `tfsdk:"vswitch_id"`
	VIP         types.String `tfsdk:"vip"`
	IPAddresses types.List   `tfsdk:"ip_addresses"`
	Netmask     types.Int32  `tfsdk:"netmask"`
}

type vpnRouterDHCPServerModel struct {
	InterfaceIndex types.Int32  `tfsdk:"interface_index"`
	RangeStart     types.String `tfsdk:"range_start"`
	RangeStop      types.String `tfsdk:"range_stop"`
	DNSServers     types.List   `tfsdk:"dns_servers"`
}

type vpnRouterDHCPStaticMappingModel struct {
	IPAddress  types.String `tfsdk:"ip_address"`
	MACAddress types.String `tfsdk:"mac_address"`
}

type vpnRouterDNSForwardingModel struct {
	InterfaceIndex types.Int32 `tfsdk:"interface_index"`
	DNSServers     types.List  `tfsdk:"dns_servers"`
}

func (m vpnRouterDNSForwardingModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interface_index": types.Int32Type,
		"dns_servers":     types.ListType{ElemType: types.StringType},
	}
}

type vpnRouterFirewallModel struct {
	InterfaceIndex types.Int32                  `tfsdk:"interface_index"`
	Direction      types.String                 `tfsdk:"direction"`
	Expression     []vpnRouterFirewallExprModel `tfsdk:"expression"`
}

type vpnRouterFirewallExprModel struct {
	Protocol           types.String `tfsdk:"protocol"`
	SourceNetwork      types.String `tfsdk:"source_network"`
	SourcePort         types.String `tfsdk:"source_port"`
	DestinationNetwork types.String `tfsdk:"destination_network"`
	DestinationPort    types.String `tfsdk:"destination_port"`
	Allow              types.Bool   `tfsdk:"allow"`
	Logging            types.Bool   `tfsdk:"logging"`
	Description        types.String `tfsdk:"description"`
}

type vpnRouterL2TPModel struct {
	PreSharedSecret types.String `tfsdk:"pre_shared_secret"`
	RangeStart      types.String `tfsdk:"range_start"`
	RangeStop       types.String `tfsdk:"range_stop"`
}

func (m vpnRouterL2TPModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"pre_shared_secret": types.StringType,
		"range_start":       types.StringType,
		"range_stop":        types.StringType,
	}
}

type vpnRouterPortForwardingModel struct {
	Protocol    types.String `tfsdk:"protocol"`
	PrivateIP   types.String `tfsdk:"private_ip"`
	PublicPort  types.Int32  `tfsdk:"public_port"`
	PrivatePort types.Int32  `tfsdk:"private_port"`
	Description types.String `tfsdk:"description"`
}

type vpnRouterPPTPModel struct {
	RangeStart types.String `tfsdk:"range_start"`
	RangeStop  types.String `tfsdk:"range_stop"`
}

func (m vpnRouterPPTPModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"range_start": types.StringType,
		"range_stop":  types.StringType,
	}
}

type vpnRouterWireGuardModel struct {
	IPAddress types.String                  `tfsdk:"ip_address"`
	PublicKey types.String                  `tfsdk:"public_key"`
	Peer      []vpnRouterWireGuardPeerModel `tfsdk:"peer"`
}

func (m vpnRouterWireGuardModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_address": types.StringType,
		"public_key": types.StringType,
		"peer":       types.ListType{ElemType: types.ObjectType{AttrTypes: vpnRouterWireGuardPeerModel{}.AttributeTypes()}},
	}
}

type vpnRouterWireGuardPeerModel struct {
	Name      types.String `tfsdk:"name"`
	IPAddress types.String `tfsdk:"ip_address"`
	PublicKey types.String `tfsdk:"public_key"`
}

func (m vpnRouterWireGuardPeerModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":       types.StringType,
		"ip_address": types.StringType,
		"public_key": types.StringType,
	}
}

type vpnRouterSiteToSiteVPNModel struct {
	Peer            types.String `tfsdk:"peer"`
	RemoteID        types.String `tfsdk:"remote_id"`
	PreSharedSecret types.String `tfsdk:"pre_shared_secret"`
	Routes          types.List   `tfsdk:"routes"`
	LocalPrefix     types.List   `tfsdk:"local_prefix"`
}

type vpnRouterSiteToSiteVPNParameterModel struct {
	IKE            types.Object `tfsdk:"ike"`
	ESP            types.Object `tfsdk:"esp"`
	EncryptionAlgo types.String `tfsdk:"encryption_algo"`
	HashAlgo       types.String `tfsdk:"hash_algo"`
	DHGroup        types.String `tfsdk:"dh_group"`
}

func (m vpnRouterSiteToSiteVPNParameterModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ike":             types.ObjectType{AttrTypes: vpnRouterIKEModel{}.AttributeTypes()},
		"esp":             types.ObjectType{AttrTypes: vpnRouterESPModel{}.AttributeTypes()},
		"encryption_algo": types.StringType,
		"hash_algo":       types.StringType,
		"dh_group":        types.StringType,
	}
}

type vpnRouterIKEModel struct {
	Lifetime types.Int64  `tfsdk:"lifetime"`
	DPD      types.Object `tfsdk:"dpd"`
}

func (m vpnRouterIKEModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"lifetime": types.Int64Type,
		"dpd":      types.ObjectType{AttrTypes: vpnRouterDPDModel{}.AttributeTypes()},
	}
}

type vpnRouterDPDModel struct {
	Interval types.Int32 `tfsdk:"interval"`
	Timeout  types.Int32 `tfsdk:"timeout"`
}

func (m vpnRouterDPDModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interval": types.Int32Type,
		"timeout":  types.Int32Type,
	}
}

type vpnRouterESPModel struct {
	Lifetime types.Int64 `tfsdk:"lifetime"`
}

func (m vpnRouterESPModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"lifetime": types.Int64Type,
	}
}

type vpnRouterStaticNATModel struct {
	PublicIP    types.String `tfsdk:"public_ip"`
	PrivateIP   types.String `tfsdk:"private_ip"`
	Description types.String `tfsdk:"description"`
}

type vpnRouterStaticRouteModel struct {
	Prefix  types.String `tfsdk:"prefix"`
	NextHop types.String `tfsdk:"next_hop"`
}

type vpnRouterScheduledMaintenanceModel struct {
	DayOfWeek types.String `tfsdk:"day_of_week"`
	Hour      types.Int32  `tfsdk:"hour"`
}

func (m vpnRouterScheduledMaintenanceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"day_of_week": types.StringType,
		"hour":        types.Int32Type,
	}
}

func (model *vpnRouterBaseModel) updateState(ctx context.Context, client *common.APIClient, zone string, vpnRouter *iaas.VPCRouter) (bool, error) {
	if vpnRouter.Availability.IsFailed() {
		return true, fmt.Errorf("got unexpected state: VPCRouter[%d].Availability is failed", vpnRouter.ID)
	}

	model.UpdateBaseState(vpnRouter.ID.String(), vpnRouter.Name, vpnRouter.Description, vpnRouter.Tags)
	model.Zone = types.StringValue(zone)
	model.Plan = types.StringValue(flattenVPNRouterPlan(vpnRouter))
	model.PublicIP = types.StringValue(flattenVPNRouterGlobalAddress(vpnRouter))
	model.PublicNetmask = types.Int64Value(int64(flattenVPNRouterGlobalNetworkMaskLen(vpnRouter)))
	model.PublicNetworkInterface = flattenVPNRouterPublicNetworkInterface(vpnRouter)
	model.InternetConnection = types.BoolValue(vpnRouter.Settings.InternetConnectionEnabled.Bool())
	model.PrivateNetworkInterface = flattenVPNRouterPrivateNetworkInterfaces(vpnRouter)
	model.DHCPServer = flattenVPNRouterDHCPServers(vpnRouter)
	model.DHCPStaticMapping = flattenVPNRouterDHCPStaticMappings(vpnRouter)
	model.DNSForwarding = flattenVPNRouterDNSForwarding(vpnRouter)
	model.Firewall = flattenVPNRouterFirewalls(vpnRouter)
	model.L2TP = flattenVPNRouterL2TP(vpnRouter)
	model.PPTP = flattenVPNRouterPPTP(vpnRouter)
	if vpnRouter.Settings.SyslogHost != "" {
		model.SyslogHost = types.StringValue(vpnRouter.Settings.SyslogHost)
	} else {
		model.SyslogHost = types.StringNull()
	}
	if vpnRouter.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(vpnRouter.IconID.String())
	}

	// get public key from /:id/Status API
	status, err := iaas.NewVPCRouterOp(client).Status(ctx, zone, vpnRouter.ID)
	if err != nil {
		return false, err
	}
	wireGuardPublicKey := ""
	if status != nil && status.WireGuard != nil {
		wireGuardPublicKey = status.WireGuard.PublicKey
	}
	model.WireGuard = flattenVPNRouterWireGuard(vpnRouter, wireGuardPublicKey)
	model.PortForwarding = flattenVPNRouterPortForwardings(vpnRouter)
	model.SiteToSiteVPN = flattenVPNRouterSiteToSiteConfig(vpnRouter)
	model.SiteToSiteVPNParameter = flattenVPNRouterSiteToSiteParameter(vpnRouter)
	model.StaticNAT = flattenVPNRouterStaticNAT(vpnRouter)
	model.StaticRoute = flattenVPNRouterStaticRoutes(vpnRouter)
	model.ScheduledMaintenance = flattenVPNRouterScheduledMaintenance(vpnRouter)
	model.MonitoringSuite = common.FlattenMonitoringSuite(vpnRouter.Settings.MonitoringSuite)

	return false, nil
}

func flattenVPNRouterPlan(vpcRouter *iaas.VPCRouter) string {
	return iaastypes.VPCRouterPlanNameMap[vpcRouter.PlanID]
}

func flattenVPNRouterPublicNetworkInterface(vpcRouter *iaas.VPCRouter) types.Object {
	v := types.ObjectNull(vpnRouterPublicNetworkInterfaceModel{}.AttributeTypes())
	if vpcRouter.PlanID == iaastypes.VPCRouterPlans.Standard {
		return v
	}
	m := vpnRouterPublicNetworkInterfaceModel{
		VSwitchID:   types.StringValue(flattenVPNRouterSwitchID(vpcRouter)),
		VIP:         types.StringValue(flattenVPNRouterVIP(vpcRouter)),
		IPAddresses: common.StringsToTlist(flattenVPNRouterIPAddresses(vpcRouter)),
		VRID:        types.Int64Value(int64(flattenVPNRouterVRID(vpcRouter))),
	}
	aliases := flattenVPNRouterIPAliases(vpcRouter)
	if len(aliases) > 0 {
		m.Aliases = common.StringsToTlist(aliases)
	}
	value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
	if diags.HasError() {
		return v
	}
	return value
}

func flattenVPNRouterPrivateNetworkInterfaces(vpcRouter *iaas.VPCRouter) []vpnRouterPrivateNetworkInterfaceModel {
	var interfaces []vpnRouterPrivateNetworkInterfaceModel
	if len(vpcRouter.Interfaces) > 0 {
		for _, iface := range vpcRouter.Settings.Interfaces {
			if iface.Index == 0 {
				continue
			}
			// find nic from data.Interfaces
			var nic *iaas.VPCRouterInterface
			for _, n := range vpcRouter.Interfaces {
				if iface.Index == n.Index {
					nic = n
					break
				}
			}

			if nic != nil {
				v := vpnRouterPrivateNetworkInterfaceModel{
					VSwitchID:   types.StringValue(nic.SwitchID.String()),
					IPAddresses: common.StringsToTlist(iface.IPAddress),
					Netmask:     types.Int32Value(int32(iface.NetworkMaskLen)),
					Index:       types.Int32Value(int32(iface.Index)),
				}
				if iface.VirtualIPAddress != "" {
					v.VIP = types.StringValue(iface.VirtualIPAddress)
				}
				interfaces = append(interfaces, v)
			}
		}
	}
	return interfaces
}

func flattenVPNRouterGlobalAddress(vpcRouter *iaas.VPCRouter) string {
	if vpcRouter.PlanID == iaastypes.VPCRouterPlans.Standard {
		return vpcRouter.Interfaces[0].IPAddress
	}
	return vpcRouter.Settings.Interfaces[0].VirtualIPAddress
}

func flattenVPNRouterGlobalNetworkMaskLen(vpcRouter *iaas.VPCRouter) int {
	return vpcRouter.Interfaces[0].SubnetNetworkMaskLen
}

func flattenVPNRouterSwitchID(vpcRouter *iaas.VPCRouter) string {
	if vpcRouter.PlanID != iaastypes.VPCRouterPlans.Standard {
		return vpcRouter.Interfaces[0].SwitchID.String()
	}
	return ""
}

func flattenVPNRouterVIP(vpcRouter *iaas.VPCRouter) string {
	if vpcRouter.PlanID != iaastypes.VPCRouterPlans.Standard {
		return vpcRouter.Settings.Interfaces[0].VirtualIPAddress
	}
	return ""
}

func flattenVPNRouterIPAddresses(vpcRouter *iaas.VPCRouter) []string {
	if vpcRouter.PlanID != iaastypes.VPCRouterPlans.Standard {
		return vpcRouter.Settings.Interfaces[0].IPAddress
	}
	return []string{}
}

func flattenVPNRouterIPAliases(vpcRouter *iaas.VPCRouter) []string {
	if vpcRouter.PlanID != iaastypes.VPCRouterPlans.Standard {
		return vpcRouter.Settings.Interfaces[0].IPAliases
	}
	return []string{}
}

func flattenVPNRouterVRID(vpcRouter *iaas.VPCRouter) int {
	if vpcRouter.PlanID != iaastypes.VPCRouterPlans.Standard {
		return vpcRouter.Settings.VRID
	}
	return 0
}

func flattenVPNRouterStaticNAT(vpcRouter *iaas.VPCRouter) []vpnRouterStaticNATModel {
	var staticNATs []vpnRouterStaticNATModel
	for _, s := range vpcRouter.Settings.StaticNAT {
		staticNATs = append(staticNATs, vpnRouterStaticNATModel{
			PublicIP:    types.StringValue(s.GlobalAddress),
			PrivateIP:   types.StringValue(s.PrivateAddress),
			Description: types.StringValue(s.Description),
		})
	}
	return staticNATs
}

func flattenVPNRouterDHCPServers(vpcRouter *iaas.VPCRouter) []vpnRouterDHCPServerModel {
	var dhcpServers []vpnRouterDHCPServerModel
	for _, d := range vpcRouter.Settings.DHCPServer {
		s := vpnRouterDHCPServerModel{
			RangeStart:     types.StringValue(d.RangeStart),
			RangeStop:      types.StringValue(d.RangeStop),
			InterfaceIndex: types.Int32Value(int32(vpcRouterInterfaceNameToIndex(d.Interface))),
		}
		if len(d.DNSServers) > 0 {
			s.DNSServers = common.StringsToTlist(d.DNSServers)
		}
		dhcpServers = append(dhcpServers, s)
	}
	return dhcpServers
}

func flattenVPNRouterDHCPStaticMappings(vpcRouter *iaas.VPCRouter) []vpnRouterDHCPStaticMappingModel {
	var staticMappings []vpnRouterDHCPStaticMappingModel
	for _, d := range vpcRouter.Settings.DHCPStaticMapping {
		staticMappings = append(staticMappings, vpnRouterDHCPStaticMappingModel{
			IPAddress:  types.StringValue(d.IPAddress),
			MACAddress: types.StringValue(d.MACAddress),
		})
	}
	return staticMappings
}

func flattenVPNRouterDNSForwarding(vpcRouter *iaas.VPCRouter) types.Object {
	v := types.ObjectNull(vpnRouterDNSForwardingModel{}.AttributeTypes())
	if vpcRouter.Settings.DNSForwarding != nil {
		m := vpnRouterDNSForwardingModel{
			InterfaceIndex: types.Int32Value(int32(vpcRouterInterfaceNameToIndex(vpcRouter.Settings.DNSForwarding.Interface))),
		}
		if len(vpcRouter.Settings.DNSForwarding.DNSServers) > 0 {
			m.DNSServers = common.StringsToTlist(vpcRouter.Settings.DNSForwarding.DNSServers)
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	}
	return v
}

func flattenVPNRouterFirewalls(vpcRouter *iaas.VPCRouter) []vpnRouterFirewallModel {
	var firewallRules []vpnRouterFirewallModel
	for i, configs := range vpcRouter.Settings.Firewall {
		directionRules := map[string][]*iaas.VPCRouterFirewallRule{
			"send":    configs.Send,
			"receive": configs.Receive,
		}

		for direction, rules := range directionRules {
			if len(rules) == 0 {
				continue
			}
			var expressions []vpnRouterFirewallExprModel
			for _, rule := range rules {
				expression := vpnRouterFirewallExprModel{
					SourceNetwork:      types.StringValue(rule.SourceNetwork.String()),
					SourcePort:         types.StringValue(rule.SourcePort.String()),
					DestinationNetwork: types.StringValue(rule.DestinationNetwork.String()),
					DestinationPort:    types.StringValue(rule.DestinationPort.String()),
					Allow:              types.BoolValue(rule.Action.IsAllow()),
					Protocol:           types.StringValue(rule.Protocol.String()),
					Logging:            types.BoolValue(rule.Logging.Bool()),
					Description:        types.StringValue(rule.Description),
				}
				expressions = append(expressions, expression)
			}
			firewallRules = append(firewallRules, vpnRouterFirewallModel{
				InterfaceIndex: types.Int32Value(int32(i)),
				Direction:      types.StringValue(direction),
				Expression:     expressions,
			})
		}
	}
	return firewallRules
}

func flattenVPNRouterPPTP(vpcRouter *iaas.VPCRouter) types.Object {
	v := types.ObjectNull(vpnRouterPPTPModel{}.AttributeTypes())
	if vpcRouter.Settings.PPTPServerEnabled.Bool() {
		m := vpnRouterPPTPModel{
			RangeStart: types.StringValue(vpcRouter.Settings.PPTPServer.RangeStart),
			RangeStop:  types.StringValue(vpcRouter.Settings.PPTPServer.RangeStop),
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	}
	return v
}

func flattenVPNRouterL2TP(vpcRouter *iaas.VPCRouter) types.Object {
	v := types.ObjectNull(vpnRouterL2TPModel{}.AttributeTypes())
	if vpcRouter.Settings.L2TPIPsecServerEnabled.Bool() {
		m := vpnRouterL2TPModel{
			PreSharedSecret: types.StringValue(vpcRouter.Settings.L2TPIPsecServer.PreSharedSecret),
			RangeStart:      types.StringValue(vpcRouter.Settings.L2TPIPsecServer.RangeStart),
			RangeStop:       types.StringValue(vpcRouter.Settings.L2TPIPsecServer.RangeStop),
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	}
	return v
}

func flattenVPNRouterWireGuard(vpcRouter *iaas.VPCRouter, publicKey string) types.Object {
	v := types.ObjectNull(vpnRouterWireGuardModel{}.AttributeTypes())
	if vpcRouter.Settings.WireGuardEnabled.Bool() {
		var peers []vpnRouterWireGuardPeerModel
		for _, peer := range vpcRouter.Settings.WireGuard.Peers {
			peers = append(peers, vpnRouterWireGuardPeerModel{
				Name:      types.StringValue(peer.Name),
				IPAddress: types.StringValue(peer.IPAddress),
				PublicKey: types.StringValue(peer.PublicKey),
			})
		}

		m := vpnRouterWireGuardModel{
			IPAddress: types.StringValue(vpcRouter.Settings.WireGuard.IPAddress),
			PublicKey: types.StringValue(publicKey),
			Peer:      peers,
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	}
	return v
}

func flattenVPNRouterPortForwardings(vpcRouter *iaas.VPCRouter) []vpnRouterPortForwardingModel {
	var portForwardings []vpnRouterPortForwardingModel
	for _, p := range vpcRouter.Settings.PortForwarding {
		globalPort := p.GlobalPort.Int()
		privatePort := p.PrivatePort.Int()
		portForwardings = append(portForwardings, vpnRouterPortForwardingModel{
			Protocol:    types.StringValue(string(p.Protocol)),
			PrivateIP:   types.StringValue(p.PrivateAddress),
			PublicPort:  types.Int32Value(int32(globalPort)),
			PrivatePort: types.Int32Value(int32(privatePort)),
			Description: types.StringValue(p.Description),
		})
	}
	return portForwardings
}

func flattenVPNRouterSiteToSiteConfig(vpcRouter *iaas.VPCRouter) []vpnRouterSiteToSiteVPNModel {
	var s2sSettings []vpnRouterSiteToSiteVPNModel
	if vpcRouter.Settings.SiteToSiteIPsecVPN != nil {
		for _, s := range vpcRouter.Settings.SiteToSiteIPsecVPN.Config {
			s2sSettings = append(s2sSettings, vpnRouterSiteToSiteVPNModel{
				Peer:            types.StringValue(s.Peer),
				RemoteID:        types.StringValue(s.RemoteID),
				PreSharedSecret: types.StringValue(s.PreSharedSecret),
				Routes:          common.StringsToTlist(s.Routes),
				LocalPrefix:     common.StringsToTlist(s.LocalPrefix),
			})
		}
	}
	return s2sSettings
}

func flattenVPNRouterSiteToSiteParameter(vpcRouter *iaas.VPCRouter) types.Object {
	v := types.ObjectNull(vpnRouterSiteToSiteVPNParameterModel{}.AttributeTypes())
	if vpcRouter.Settings.SiteToSiteIPsecVPN != nil {
		ctx := context.Background()
		m := vpnRouterSiteToSiteVPNParameterModel{
			EncryptionAlgo: types.StringValue(vpcRouter.Settings.SiteToSiteIPsecVPN.EncryptionAlgo),
			HashAlgo:       types.StringValue(vpcRouter.Settings.SiteToSiteIPsecVPN.HashAlgo),
			DHGroup:        types.StringValue(vpcRouter.Settings.SiteToSiteIPsecVPN.DHGroup),
		}
		if vpcRouter.Settings.SiteToSiteIPsecVPN.IKE != nil {
			ike := vpnRouterIKEModel{
				Lifetime: types.Int64Value(int64(vpcRouter.Settings.SiteToSiteIPsecVPN.IKE.Lifetime)),
			}
			if vpcRouter.Settings.SiteToSiteIPsecVPN.IKE.DPD != nil {
				dpd := vpnRouterDPDModel{
					Interval: types.Int32Value(int32(vpcRouter.Settings.SiteToSiteIPsecVPN.IKE.DPD.Interval)),
					Timeout:  types.Int32Value(int32(vpcRouter.Settings.SiteToSiteIPsecVPN.IKE.DPD.Timeout)),
				}
				dpdValue, diags := types.ObjectValueFrom(ctx, dpd.AttributeTypes(), dpd)
				if diags.HasError() {
					return v
				}
				ike.DPD = dpdValue
			}
			ikeValue, diags := types.ObjectValueFrom(ctx, ike.AttributeTypes(), ike)
			if diags.HasError() {
				return v
			}
			m.IKE = ikeValue
		}
		if vpcRouter.Settings.SiteToSiteIPsecVPN.ESP != nil {
			esp := vpnRouterESPModel{
				Lifetime: types.Int64Value(int64(vpcRouter.Settings.SiteToSiteIPsecVPN.ESP.Lifetime)),
			}
			espValue, diags := types.ObjectValueFrom(ctx, esp.AttributeTypes(), esp)
			if diags.HasError() {
				return v
			}
			m.ESP = espValue
		}
		value, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	}
	return v
}

func flattenVPNRouterStaticRoutes(vpcRouter *iaas.VPCRouter) []vpnRouterStaticRouteModel {
	var staticRoutes []vpnRouterStaticRouteModel
	for _, s := range vpcRouter.Settings.StaticRoute {
		staticRoutes = append(staticRoutes, vpnRouterStaticRouteModel{
			Prefix:  types.StringValue(s.Prefix),
			NextHop: types.StringValue(s.NextHop),
		})
	}
	return staticRoutes
}

func flattenVPNRouterScheduledMaintenance(vpcRouter *iaas.VPCRouter) types.Object {
	v := types.ObjectNull(vpnRouterScheduledMaintenanceModel{}.AttributeTypes())
	if vpcRouter.Settings != nil && vpcRouter.Settings.ScheduledMaintenance != nil {
		model := vpnRouterScheduledMaintenanceModel{
			DayOfWeek: types.StringValue(iaastypes.DayOfTheWeekFromInt(vpcRouter.Settings.ScheduledMaintenance.DayOfWeek).String()),
			Hour:      types.Int32Value(int32(vpcRouter.Settings.ScheduledMaintenance.Hour)),
		}
		tflog.Info(context.Background(), fmt.Sprintf("ScheduledMaintenance: %#v", model))
		value, diags := types.ObjectValueFrom(context.Background(), model.AttributeTypes(), model)
		if diags.HasError() {
			tflog.Info(context.Background(), fmt.Sprintf("ScheduledMaintenance error: %#v", diags.Errors()))
			return v
		}
		return value
	}
	return v
}
