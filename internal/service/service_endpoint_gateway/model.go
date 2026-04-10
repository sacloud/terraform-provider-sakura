// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package service_endpoint_gateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/service-endpoint-gateway-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type segEndpointSettingModel struct {
	ObjectStorageEndpoints        types.List `tfsdk:"object_storage_endpoints"`
	MonitoringSuiteEndpoints      types.List `tfsdk:"monitoring_suite_endpoints"`
	ContainerRegistryEndpoints    types.List `tfsdk:"container_registry_endpoints"`
	AIEngineEndpoints             types.List `tfsdk:"ai_engine_endpoints"`
	AppRunDedicatedControlEnabled types.Bool `tfsdk:"app_run_dedicated_control_enabled"`
}

type segDNSForwardingModel struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	PrivateHostedZone types.String `tfsdk:"private_hosted_zone"`
	UpstreamDNS1      types.String `tfsdk:"upstream_dns_1"`
	UpstreamDNS2      types.String `tfsdk:"upstream_dns_2"`
}

type segBaseModel struct {
	ID                     types.String `tfsdk:"id"`
	Zone                   types.String `tfsdk:"zone"`
	VSwitchID              types.String `tfsdk:"vswitch_id"`
	ServerIPAddresses      types.List   `tfsdk:"server_ip_addresses"`
	NetMask                types.Int32  `tfsdk:"netmask"`
	EndpointSetting        types.Object `tfsdk:"endpoint_setting"`
	MonitoringSuiteEnabled types.Bool   `tfsdk:"monitoring_suite_enable"`
	DNSForwarding          types.Object `tfsdk:"dns_forwarding"`
}

func (model *segBaseModel) updateState(appliance *v1.ModelsApplianceAppliance) error {
	if appliance.Availability != v1.ModelsApplianceApplianceAvailabilityAvailable {
		return fmt.Errorf("got unexpected state: Appliance[%s].Availability is failed", appliance.ID)
	}

	model.ID = types.StringValue(appliance.ID)
	model.Zone = types.StringValue(appliance.Switch.Zone.Name)
	model.VSwitchID = types.StringValue(appliance.Switch.ID)
	model.ServerIPAddresses = flattenServerIPAddreses(appliance.Remark.Value.Servers)
	model.NetMask = types.Int32Value(appliance.Remark.Value.Network.NetworkMaskLen)
	model.EndpointSetting = flattenEndpointSetting(appliance.Settings)
	model.MonitoringSuiteEnabled = flattenMonitoringSuiteEnabled(appliance.Settings)
	model.DNSForwarding = flattenDNSForwarding(appliance.Settings)
	return nil
}

func flattenEndpointSetting(setting v1.NilModelsSettingsApplianceSettings) types.Object {
	if setting.IsNull() {
		return types.ObjectNull(segEndpointSettingModel{}.AttributeTypes())
	}

	settingModel := segEndpointSettingModel{
		ObjectStorageEndpoints:        types.ListNull(types.StringType),
		MonitoringSuiteEndpoints:      types.ListNull(types.StringType),
		ContainerRegistryEndpoints:    types.ListNull(types.StringType),
		AIEngineEndpoints:             types.ListNull(types.StringType),
		AppRunDedicatedControlEnabled: types.BoolNull(),
	}

	enableServiceSettings := setting.Value.ServiceEndpointGateway.EnabledServices
	for _, setting := range enableServiceSettings {
		switch setting.Type {
		case v1.ModelsSettingsEnabledServiceTypeObjectStorage:
			settingModel.ObjectStorageEndpoints = common.StringsToTlist(setting.Config.Endpoints)
		case v1.ModelsSettingsEnabledServiceTypeMonitoringSuite:
			settingModel.MonitoringSuiteEndpoints = common.StringsToTlist(setting.Config.Endpoints)
		case v1.ModelsSettingsEnabledServiceTypeContainerRegistry:
			settingModel.ContainerRegistryEndpoints = common.StringsToTlist(setting.Config.Endpoints)
		case v1.ModelsSettingsEnabledServiceTypeAIEngine:
			settingModel.AIEngineEndpoints = common.StringsToTlist(setting.Config.Endpoints)
		case v1.ModelsSettingsEnabledServiceTypeAppRunDedicatedControlPlane:
			settingModel.AppRunDedicatedControlEnabled = types.BoolValue(setting.Config.Mode.Value == v1.ModelsSettingsServiceConfigModeManaged)
		}
	}

	value, diags := types.ObjectValueFrom(context.Background(), settingModel.AttributeTypes(), settingModel)
	if diags.HasError() {
		return types.ObjectNull(segEndpointSettingModel{}.AttributeTypes())
	}

	return value
}

func flattenMonitoringSuiteEnabled(setting v1.NilModelsSettingsApplianceSettings) types.Bool {
	if setting.IsNull() {
		return types.BoolNull()
	}

	monitoringSuite := setting.Value.ServiceEndpointGateway.MonitoringSuite
	if !monitoringSuite.Set {
		return types.BoolNull()
	}

	return types.BoolValue(monitoringSuite.Value.Enabled == v1.ModelsSettingsMonitoringSuiteSettingsEnabledTrue)
}

func flattenDNSForwarding(setting v1.NilModelsSettingsApplianceSettings) types.Object {
	if setting.IsNull() {
		return types.ObjectNull(segDNSForwardingModel{}.AttributeTypes())
	}

	if !setting.Value.ServiceEndpointGateway.DNSForwarding.Set {
		return types.ObjectNull(segDNSForwardingModel{}.AttributeTypes())
	}

	dnsForwarding := setting.Value.ServiceEndpointGateway.DNSForwarding.Value
	model := segDNSForwardingModel{
		Enabled:           types.BoolValue(dnsForwarding.Enabled == v1.ModelsSettingsDNSForwardingSettingsEnabledTrue),
		PrivateHostedZone: types.StringValue(dnsForwarding.PrivateHostedZone),
		UpstreamDNS1:      types.StringValue(dnsForwarding.UpstreamDNS1),
		UpstreamDNS2:      types.StringValue(dnsForwarding.UpstreamDNS2),
	}

	value, diags := types.ObjectValueFrom(context.Background(), model.AttributeTypes(), model)
	if diags.HasError() {
		return types.ObjectNull(segDNSForwardingModel{}.AttributeTypes())
	}

	return value
}

func flattenServerIPAddreses(servers []v1.ModelsRemarkServerRemark) types.List {
	serverlist := make([]string, 0, len(servers))
	for _, server := range servers {
		if server.IPAddress != "" {
			serverlist = append(serverlist, server.IPAddress)
		}
	}

	return common.StringsToTlist(serverlist)
}

// segEndpointSettingAttrTypes returns the attribute types for segEndpointSettingModel
func (dns segEndpointSettingModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"object_storage_endpoints":          types.ListType{ElemType: types.StringType},
		"monitoring_suite_endpoints":        types.ListType{ElemType: types.StringType},
		"container_registry_endpoints":      types.ListType{ElemType: types.StringType},
		"ai_engine_endpoints":               types.ListType{ElemType: types.StringType},
		"app_run_dedicated_control_enabled": types.BoolType,
	}
}

// segDNSForwardingAttrTypes returns the attribute types for segDNSForwardingModel
func (dns segDNSForwardingModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":             types.BoolType,
		"private_hosted_zone": types.StringType,
		"upstream_dns_1":      types.StringType,
		"upstream_dns_2":      types.StringType,
	}
}
