// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package vpn_router

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/defaults"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/iaas-service-go/setup"
	"github.com/sacloud/iaas-service-go/vpcrouter/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

// resource.goに入れてもいいが、量が多いのでとりあえず分割

func expandVPNRouterBuilder(model *vpnRouterResourceModel, client *common.APIClient, zone string) *builder.Builder {
	return &builder.Builder{
		Zone:                  zone,
		Name:                  model.Name.ValueString(),
		Description:           model.Description.ValueString(),
		Tags:                  common.TsetToStrings(model.Tags),
		IconID:                common.ExpandSakuraCloudID(model.IconID),
		PlanID:                expandVPNRouterPlanID(model),
		Version:               int(model.Version.ValueInt32()),
		NICSetting:            expandVPNRouterNICSetting(model),
		AdditionalNICSettings: expandVPNRouterAdditionalNICSettings(model),
		RouterSetting:         expandVPNRouterSettings(model),
		SetupOptions: &setup.Options{
			BootAfterBuild:        true,
			NICUpdateWaitDuration: defaults.DefaultNICUpdateWaitDuration,
		},
		Client: iaas.NewVPCRouterOp(client),
	}
}

func expandVPNRouterPlanID(model *vpnRouterResourceModel) iaastypes.ID {
	return iaastypes.VPCRouterPlanIDMap[model.Plan.ValueString()]
}

func expandVPNRouterNICSetting(model *vpnRouterResourceModel) builder.NICSettingHolder {
	planID := expandVPNRouterPlanID(model)
	switch planID {
	case iaastypes.VPCRouterPlans.Standard:
		return &builder.StandardNICSetting{}
	default:
		nic := expandVPNRouterPublicNetworkInterface(model)
		return &builder.PremiumNICSetting{
			SwitchID:         nic.switchID,
			IPAddresses:      nic.ipAddresses,
			VirtualIPAddress: nic.vip,
			IPAliases:        nic.ipAliases,
		}
	}
}

type vpcRouterPublicNetworkInterface struct {
	switchID    iaastypes.ID
	ipAddresses []string
	vip         string
	ipAliases   []string
	vrid        int
}

func expandVPNRouterPublicNetworkInterface(model *vpnRouterResourceModel) *vpcRouterPublicNetworkInterface {
	if model.PublicNetworkInterface.IsNull() || model.PublicNetworkInterface.IsUnknown() {
		return nil
	}

	var d vpnRouterPublicNetworkInterfaceModel
	diags := model.PublicNetworkInterface.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &vpcRouterPublicNetworkInterface{
		switchID:    common.ExpandSakuraCloudID(d.VSwitchID),
		ipAddresses: common.TlistToStringsOrDefault(d.IPAddresses),
		vip:         d.VIP.ValueString(),
		ipAliases:   common.TlistToStringsOrDefault(d.Aliases),
		vrid:        int(d.VRID.ValueInt64()),
	}
}

func expandVPNRouterAdditionalNICSettings(model *vpnRouterResourceModel) []builder.AdditionalNICSettingHolder {
	var results []builder.AdditionalNICSettingHolder
	planID := expandVPNRouterPlanID(model)
	interfaces := model.PrivateNetworkInterface
	for _, iface := range interfaces {
		var nicSetting builder.AdditionalNICSettingHolder
		ipAddresses := common.TlistToStringsOrDefault(iface.IPAddresses)

		switch planID {
		case iaastypes.VPCRouterPlans.Standard:
			nicSetting = &builder.AdditionalStandardNICSetting{
				SwitchID:       common.ExpandSakuraCloudID(iface.VSwitchID),
				IPAddress:      ipAddresses[0],
				NetworkMaskLen: int(iface.Netmask.ValueInt32()),
				Index:          int(iface.Index.ValueInt32()),
			}
		default:
			nicSetting = &builder.AdditionalPremiumNICSetting{
				SwitchID:         common.ExpandSakuraCloudID(iface.VSwitchID),
				NetworkMaskLen:   int(iface.Netmask.ValueInt32()),
				IPAddresses:      ipAddresses,
				VirtualIPAddress: iface.VIP.ValueString(),
				Index:            int(iface.Index.ValueInt32()),
			}
		}
		results = append(results, nicSetting)
	}
	return results
}

func expandVPNRouterSettings(model *vpnRouterResourceModel) *builder.RouterSetting {
	nic := expandVPNRouterPublicNetworkInterface(model)
	vrid := 0
	if nic != nil {
		vrid = nic.vrid
	}
	return &builder.RouterSetting{
		VRID:                      vrid,
		InternetConnectionEnabled: iaastypes.StringFlag(model.InternetConnection.ValueBool()),
		StaticNAT:                 expandVPNRouterStaticNATList(model),
		PortForwarding:            expandVPNRouterPortForwardingList(model),
		Firewall:                  expandVPNRouterFirewallList(model),
		DHCPServer:                expandVPNRouterDHCPServerList(model),
		DHCPStaticMapping:         expandVPNRouterDHCPStaticMappingList(model),
		DNSForwarding:             expandVPNRouterDNSForwarding(model),
		PPTPServer:                expandVPNRouterPPTP(model),
		L2TPIPsecServer:           expandVPNRouterL2TP(model),
		RemoteAccessUsers:         expandVPNRouterUserList(model),
		WireGuard:                 expandVPNRouterWireGuard(model),
		SiteToSiteIPsecVPN:        expandVPNRouterSiteToSite(model),
		StaticRoute:               expandVPNRouterStaticRouteList(model),
		SyslogHost:                model.SyslogHost.ValueString(),
		ScheduledMaintenance:      expandVPNRouterScheduledMaintenance(model),
		MonitoringSuite:           common.ExpandMonitoringSuite(model.MonitoringSuite),
	}
}

func expandVPNRouterStaticNATList(model *vpnRouterResourceModel) []*iaas.VPCRouterStaticNAT {
	if values := model.StaticNAT; len(values) > 0 {
		var results []*iaas.VPCRouterStaticNAT
		for _, v := range values {
			results = append(results, expandVPNRouterStaticNAT(&v))
		}
		return results
	}
	return nil
}

func expandVPNRouterStaticNAT(model *vpnRouterStaticNATModel) *iaas.VPCRouterStaticNAT {
	return &iaas.VPCRouterStaticNAT{
		GlobalAddress:  model.PublicIP.ValueString(),
		PrivateAddress: model.PrivateIP.ValueString(),
		Description:    model.Description.ValueString(),
	}
}

func expandVPNRouterDHCPServerList(model *vpnRouterResourceModel) []*iaas.VPCRouterDHCPServer {
	if values := model.DHCPServer; len(values) > 0 {
		var results []*iaas.VPCRouterDHCPServer
		for _, v := range values {
			results = append(results, expandVPNRouterDHCPServer(&v))
		}
		return results
	}
	return nil
}

func expandVPNRouterDHCPServer(model *vpnRouterDHCPServerModel) *iaas.VPCRouterDHCPServer {
	return &iaas.VPCRouterDHCPServer{
		Interface:  fmt.Sprintf("eth%d", model.InterfaceIndex.ValueInt32()),
		RangeStart: model.RangeStart.ValueString(),
		RangeStop:  model.RangeStop.ValueString(),
		DNSServers: common.TlistToStrings(model.DNSServers),
	}
}

func vpcRouterInterfaceNameToIndex(ifName string) int {
	strIndex := strings.ReplaceAll(ifName, "eth", "")
	index, err := strconv.Atoi(strIndex)
	if err != nil {
		return -1
	}
	return index
}

func expandVPNRouterDHCPStaticMappingList(model *vpnRouterResourceModel) []*iaas.VPCRouterDHCPStaticMapping {
	if values := model.DHCPStaticMapping; len(values) > 0 {
		var results []*iaas.VPCRouterDHCPStaticMapping
		for _, v := range values {
			results = append(results, expandVPNRouterDHCPStaticMapping(&v))
		}
		return results
	}
	return nil
}

func expandVPNRouterDHCPStaticMapping(model *vpnRouterDHCPStaticMappingModel) *iaas.VPCRouterDHCPStaticMapping {
	return &iaas.VPCRouterDHCPStaticMapping{
		IPAddress:  model.IPAddress.ValueString(),
		MACAddress: model.MACAddress.ValueString(),
	}
}

func expandVPNRouterDNSForwarding(model *vpnRouterResourceModel) *iaas.VPCRouterDNSForwarding {
	if model.DNSForwarding.IsNull() || model.DNSForwarding.IsUnknown() {
		return nil
	}

	var d vpnRouterDNSForwardingModel
	diags := model.DNSForwarding.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &iaas.VPCRouterDNSForwarding{
		Interface:  fmt.Sprintf("eth%d", d.InterfaceIndex.ValueInt32()),
		DNSServers: common.TlistToStrings(d.DNSServers),
	}
}

func expandVPNRouterFirewallList(model *vpnRouterResourceModel) []*iaas.VPCRouterFirewall {
	if values := model.Firewall; len(values) > 0 {
		var results []*iaas.VPCRouterFirewall
		for _, v := range values {
			results = append(results, expandVPNRouterFirewall(&v))
		}

		// インデックスごとにSend/Receiveをまとめる
		// results: {Index: 0, Send: []Rules{...}, Receive: nil} , {Index: 0, Send: nil, Receive: []Rules{...}}
		// merged: {Index: 0, Send: []Rules{...}, Receive: []Rules{...}}
		var merged []*iaas.VPCRouterFirewall
		for i := 0; i < 8; i++ {
			firewall := &iaas.VPCRouterFirewall{
				Index: i,
			}
			for _, f := range results {
				if f.Index == i {
					if len(f.Send) > 0 {
						firewall.Send = f.Send
					}
					if len(f.Receive) > 0 {
						firewall.Receive = f.Receive
					}
				}
			}
			merged = append(merged, firewall)
		}
		return merged
	}
	return nil
}

func expandVPNRouterFirewall(model *vpnRouterFirewallModel) *iaas.VPCRouterFirewall {
	index := model.InterfaceIndex.ValueInt32()
	direction := model.Direction.ValueString()
	f := &iaas.VPCRouterFirewall{
		Index: int(index),
	}
	if direction == "send" {
		f.Send = expandVPNRouterFirewallRuleList(model)
	}
	if direction == "receive" {
		f.Receive = expandVPNRouterFirewallRuleList(model)
	}
	return f
}

func expandVPNRouterFirewallRuleList(model *vpnRouterFirewallModel) []*iaas.VPCRouterFirewallRule {
	if values := model.Expression; len(values) > 0 {
		var results []*iaas.VPCRouterFirewallRule
		for _, v := range values {
			results = append(results, expandVPNRouterFirewallRule(&v))
		}
		return results
	}
	return nil
}

func expandVPNRouterFirewallRule(model *vpnRouterFirewallExprModel) *iaas.VPCRouterFirewallRule {
	allow := model.Allow.ValueBool()
	action := iaastypes.Actions.Allow
	if !allow {
		action = iaastypes.Actions.Deny
	}

	return &iaas.VPCRouterFirewallRule{
		Protocol:           iaastypes.Protocol(model.Protocol.ValueString()),
		SourceNetwork:      iaastypes.VPCFirewallNetwork(model.SourceNetwork.ValueString()),
		SourcePort:         iaastypes.VPCFirewallPort(model.SourcePort.ValueString()),
		DestinationNetwork: iaastypes.VPCFirewallNetwork(model.DestinationNetwork.ValueString()),
		DestinationPort:    iaastypes.VPCFirewallPort(model.DestinationPort.ValueString()),
		Action:             action,
		Logging:            iaastypes.StringFlag(model.Logging.ValueBool()),
		Description:        model.Description.ValueString(),
	}
}

func expandVPNRouterPPTP(model *vpnRouterResourceModel) *iaas.VPCRouterPPTPServer {
	if model.PPTP.IsNull() || model.PPTP.IsUnknown() {
		return nil
	}

	var d vpnRouterPPTPModel
	diags := model.PPTP.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &iaas.VPCRouterPPTPServer{
		RangeStart: d.RangeStart.ValueString(),
		RangeStop:  d.RangeStop.ValueString(),
	}
}

func expandVPNRouterL2TP(model *vpnRouterResourceModel) *iaas.VPCRouterL2TPIPsecServer {
	if model.L2TP.IsNull() || model.L2TP.IsUnknown() {
		return nil
	}

	var d vpnRouterL2TPModel
	diags := model.L2TP.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &iaas.VPCRouterL2TPIPsecServer{
		RangeStart:      d.RangeStart.ValueString(),
		RangeStop:       d.RangeStop.ValueString(),
		PreSharedSecret: d.PreSharedSecret.ValueString(),
	}
}

func expandVPNRouterWireGuard(model *vpnRouterResourceModel) *iaas.VPCRouterWireGuard {
	if model.WireGuard.IsNull() || model.WireGuard.IsUnknown() {
		return nil
	}

	var d vpnRouterWireGuardModel
	diags := model.WireGuard.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	var peers []*iaas.VPCRouterWireGuardPeer
	if peerValues := d.Peer; len(peerValues) > 0 {
		for _, pv := range peerValues {
			peers = append(peers, &iaas.VPCRouterWireGuardPeer{
				Name:      pv.Name.ValueString(),
				IPAddress: pv.IPAddress.ValueString(),
				PublicKey: pv.PublicKey.ValueString(),
			})
		}
	}

	return &iaas.VPCRouterWireGuard{
		IPAddress: d.IPAddress.ValueString(),
		Peers:     peers,
	}
}

func expandVPNRouterPortForwardingList(model *vpnRouterResourceModel) []*iaas.VPCRouterPortForwarding {
	if values := model.PortForwarding; len(values) > 0 {
		var results []*iaas.VPCRouterPortForwarding
		for _, v := range values {
			results = append(results, expandVPNRouterPortForwarding(&v))
		}
		return results
	}
	return nil
}

func expandVPNRouterPortForwarding(model *vpnRouterPortForwardingModel) *iaas.VPCRouterPortForwarding {
	return &iaas.VPCRouterPortForwarding{
		Protocol:       iaastypes.EVPCRouterPortForwardingProtocol(model.Protocol.ValueString()),
		GlobalPort:     iaastypes.StringNumber(model.PublicPort.ValueInt32()),
		PrivateAddress: model.PrivateIP.ValueString(),
		PrivatePort:    iaastypes.StringNumber(model.PrivatePort.ValueInt32()),
		Description:    model.Description.ValueString(),
	}
}

func expandVPNRouterSiteToSite(model *vpnRouterResourceModel) *iaas.VPCRouterSiteToSiteIPsecVPN {
	siteToSiteVPN := &iaas.VPCRouterSiteToSiteIPsecVPN{}
	if values := model.SiteToSiteVPN; len(values) > 0 {
		for _, v := range values {
			siteToSiteVPN.Config = append(siteToSiteVPN.Config, expandVPNRouterSiteToSiteConfig(&v))
		}
	}

	if model.SiteToSiteVPNParameter.IsNull() || model.SiteToSiteVPNParameter.IsUnknown() {
		return siteToSiteVPN
	}

	var d vpnRouterSiteToSiteVPNParameterModel
	diags := model.SiteToSiteVPNParameter.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return siteToSiteVPN
	}

	siteToSiteVPN.IKE = expandVPNRouterSiteToSiteParameterIKE(&d)
	siteToSiteVPN.ESP = expandVPNRouterSiteToSiteParameterESP(&d)
	siteToSiteVPN.EncryptionAlgo = d.EncryptionAlgo.ValueString()
	siteToSiteVPN.HashAlgo = d.HashAlgo.ValueString()
	siteToSiteVPN.DHGroup = d.DHGroup.ValueString()

	return siteToSiteVPN
}

func expandVPNRouterSiteToSiteConfig(model *vpnRouterSiteToSiteVPNModel) *iaas.VPCRouterSiteToSiteIPsecVPNConfig {
	return &iaas.VPCRouterSiteToSiteIPsecVPNConfig{
		Peer:            model.Peer.ValueString(),
		RemoteID:        model.RemoteID.ValueString(),
		PreSharedSecret: model.PreSharedSecret.ValueString(),
		Routes:          common.TlistToStringsOrDefault(model.Routes),
		LocalPrefix:     common.TlistToStringsOrDefault(model.LocalPrefix),
	}
}

func expandVPNRouterSiteToSiteParameterIKE(model *vpnRouterSiteToSiteVPNParameterModel) *iaas.VPCRouterSiteToSiteIPsecVPNIKE {
	if model.IKE.IsNull() || model.IKE.IsUnknown() {
		return nil
	}

	var d vpnRouterIKEModel
	diags := model.IKE.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &iaas.VPCRouterSiteToSiteIPsecVPNIKE{
		Lifetime: int(d.Lifetime.ValueInt64()),
		DPD:      expandVPNRouterSiteToSiteParameterIKEDPD(&d),
	}
}

func expandVPNRouterSiteToSiteParameterIKEDPD(model *vpnRouterIKEModel) *iaas.VPCRouterSiteToSiteIPsecVPNIKEDPD {
	if model.DPD.IsNull() || model.DPD.IsUnknown() {
		return nil
	}

	var d vpnRouterDPDModel
	diags := model.DPD.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &iaas.VPCRouterSiteToSiteIPsecVPNIKEDPD{
		Interval: int(d.Interval.ValueInt32()),
		Timeout:  int(d.Timeout.ValueInt32()),
	}
}

func expandVPNRouterSiteToSiteParameterESP(model *vpnRouterSiteToSiteVPNParameterModel) *iaas.VPCRouterSiteToSiteIPsecVPNESP {
	if model.ESP.IsNull() || model.ESP.IsUnknown() {
		return nil
	}

	var d vpnRouterESPModel
	diags := model.ESP.As(context.Background(), &d, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &iaas.VPCRouterSiteToSiteIPsecVPNESP{
		Lifetime: int(d.Lifetime.ValueInt64()),
	}
}

func expandVPNRouterStaticRouteList(model *vpnRouterResourceModel) []*iaas.VPCRouterStaticRoute {
	if values := model.StaticRoute; len(values) > 0 {
		var results []*iaas.VPCRouterStaticRoute
		for _, v := range values {
			results = append(results, expandVPNRouterStaticRoute(&v))
		}
		return results
	}
	return nil
}

func expandVPNRouterStaticRoute(model *vpnRouterStaticRouteModel) *iaas.VPCRouterStaticRoute {
	return &iaas.VPCRouterStaticRoute{
		Prefix:  model.Prefix.ValueString(),
		NextHop: model.NextHop.ValueString(),
	}
}

func expandVPNRouterUserList(model *vpnRouterResourceModel) []*iaas.VPCRouterRemoteAccessUser {
	if values := model.User; len(values) > 0 {
		var results []*iaas.VPCRouterRemoteAccessUser
		for _, v := range values {
			results = append(results, expandVPNRouterUser(&v))
		}
		return results
	}
	return nil
}

func expandVPNRouterUser(model *vpnRouterUserModel) *iaas.VPCRouterRemoteAccessUser {
	return &iaas.VPCRouterRemoteAccessUser{
		UserName: model.Name.ValueString(),
		Password: model.Password.ValueString(),
	}
}

func expandVPNRouterScheduledMaintenance(model *vpnRouterResourceModel) *iaas.VPCRouterScheduledMaintenance {
	if model.ScheduledMaintenance.IsNull() || model.ScheduledMaintenance.IsUnknown() {
		return nil
	}

	var m vpnRouterScheduledMaintenanceModel
	diags := model.ScheduledMaintenance.As(context.Background(), &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	dayOfWeek := m.DayOfWeek.ValueString()
	hour := int(m.Hour.ValueInt32())
	return &iaas.VPCRouterScheduledMaintenance{
		DayOfWeek: iaastypes.DayOfTheWeekFromString(dayOfWeek).Int(),
		Hour:      hour,
	}
}
