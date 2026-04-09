// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package service_endpoint_gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sacloud/iaas-api-go"
	seg "github.com/sacloud/service-endpoint-gateway-api-go"
	v1 "github.com/sacloud/service-endpoint-gateway-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

func isAvailableZoneForVSwitch(ctx context.Context, client *common.APIClient, zone string, vswitchID types.String, diags *diag.Diagnostics) error {
	iaasID := common.ExpandSakuraCloudID(vswitchID)
	swOp := iaas.NewSwitchOp(client)
	_, err := swOp.Read(ctx, zone, iaasID)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			return err
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read vSwitch[%s] : %s", vswitchID.ValueString(), err))
		return err
	}
	return nil
}

func getServiceEndpointGatewayAPIClient(client *common.APIClient, zone string) (*v1.Client, error) {
	apiRoot := fmt.Sprintf("https://secure.sakura.ad.jp/cloud/zone/%s/api/cloud/1.1", zone)
	return seg.NewClientWithAPIRootURL(client.SaClient, apiRoot)
}

func expandSeversIPAddresses(d types.List) []v1.ModelsRemarkServerRemark {
	var servers []v1.ModelsRemarkServerRemark
	if d.IsNull() || d.IsUnknown() {
		return servers
	}
	ipList := common.TlistToStrings(d)
	for _, ip := range ipList {
		servers = append(servers, v1.ModelsRemarkServerRemark{
			IPAddress: ip,
		})
	}
	return servers
}

func expandSEGCreateRequest(d *segResourceModel) v1.ModelsApplianceApplianceCreateRequest {
	return v1.ModelsApplianceApplianceCreateRequest{
		Appliance: v1.ModelsApplianceApplianceCreateBody{
			Remark: v1.ModelsRemarkApplianceCreateRemark{
				Switch: v1.ModelsRemarkSwitchRemark{
					ID: d.VSwitchID.ValueString(),
				},
				Network: v1.ModelsRemarkNetworkRemark{
					NetworkMaskLen: d.NetMask.ValueInt32(),
				},
				Servers: expandSeversIPAddresses(d.ServerIPAddresses),
			},
		},
	}
}

func expandSEGUpdateRequest(d *segResourceModel) v1.ModelsApplianceApplianceUpdateRequest {
	dnsForwardSetting := expandDNSForwardingSettings(d.DNSForwarding)
	return v1.ModelsApplianceApplianceUpdateRequest{
		Appliance: v1.ModelsApplianceApplianceUpdateBody{
			Settings: v1.ModelsSettingsApplianceSettings{
				ServiceEndpointGateway: v1.ModelsSettingsServiceEndpointGatewaySettings{
					EnabledServices: expandEndpointSetting(d.EndpointSetting),
					MonitoringSuite: func(enable types.Bool) v1.OptModelsSettingsMonitoringSuiteSettings {
						if enable.IsNull() || enable.IsUnknown() {
							return v1.OptModelsSettingsMonitoringSuiteSettings{
								Set: false,
							}
						}
						value := v1.ModelsSettingsMonitoringSuiteSettingsEnabledFalse
						if enable.ValueBool() {
							value = v1.ModelsSettingsMonitoringSuiteSettingsEnabledTrue
						}
						return v1.OptModelsSettingsMonitoringSuiteSettings{
							Set: true,
							Value: v1.ModelsSettingsMonitoringSuiteSettings{
								Enabled: value,
							},
						}
					}(d.MonitoringSuiteEnabled),
					DNSForwarding: dnsForwardSetting,
				},
			},
		},
	}
}

func expandEndpointSetting(d types.Object) []v1.ModelsSettingsEnabledService {
	if d.IsNull() || d.IsUnknown() {
		return []v1.ModelsSettingsEnabledService{}
	}
	var model segEndpointSettingModel
	diags := d.As(context.Background(), &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return []v1.ModelsSettingsEnabledService{}
	}

	var settings []v1.ModelsSettingsEnabledService
	if !model.ObjectStorageEndpoints.IsNull() {
		settings = append(settings, v1.ModelsSettingsEnabledService{
			Type: v1.ModelsSettingsEnabledServiceTypeObjectStorage,
			Config: v1.ModelsSettingsServiceConfig{
				Endpoints: common.TlistToStrings(model.ObjectStorageEndpoints),
			},
		})
	}
	if !model.MonitoringSuiteEndpoints.IsNull() {
		settings = append(settings, v1.ModelsSettingsEnabledService{
			Type: v1.ModelsSettingsEnabledServiceTypeMonitoringSuite,
			Config: v1.ModelsSettingsServiceConfig{
				Endpoints: common.TlistToStrings(model.MonitoringSuiteEndpoints),
			},
		})
	}
	if !model.ContainerRegistryEndpoints.IsNull() {
		settings = append(settings, v1.ModelsSettingsEnabledService{
			Type: v1.ModelsSettingsEnabledServiceTypeContainerRegistry,
			Config: v1.ModelsSettingsServiceConfig{
				Endpoints: common.TlistToStrings(model.ContainerRegistryEndpoints),
			},
		})
	}
	if !model.AIEngineEndpoints.IsNull() {
		settings = append(settings, v1.ModelsSettingsEnabledService{
			Type: v1.ModelsSettingsEnabledServiceTypeAIEngine,
			Config: v1.ModelsSettingsServiceConfig{
				Endpoints: common.TlistToStrings(model.AIEngineEndpoints),
			},
		})
	}
	if !model.AppRunDedicatedControlEnabled.IsNull() {
		settings = append(settings, v1.ModelsSettingsEnabledService{
			Type: v1.ModelsSettingsEnabledServiceTypeAppRunDedicatedControlPlane,
			Config: v1.ModelsSettingsServiceConfig{
				Mode: v1.OptModelsSettingsServiceConfigMode{
					Set: true,
					Value: func(enabled types.Bool) v1.ModelsSettingsServiceConfigMode {
						if enabled.ValueBool() {
							return v1.ModelsSettingsServiceConfigModeManaged
						}
						return v1.ModelsSettingsServiceConfigModeEmpty
					}(model.AppRunDedicatedControlEnabled),
				},
			},
		})
	}
	return settings
}

func expandDNSForwardingSettings(d types.Object) v1.OptModelsSettingsDNSForwardingSettings {
	if d.IsNull() || d.IsUnknown() {
		return v1.OptModelsSettingsDNSForwardingSettings{
			Set: false,
		}
	}
	var model segDNSForwardingModel
	diags := d.As(context.Background(), &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return v1.OptModelsSettingsDNSForwardingSettings{
			Set: false,
		}
	}
	return v1.OptModelsSettingsDNSForwardingSettings{
		Set: true,
		Value: v1.ModelsSettingsDNSForwardingSettings{
			Enabled: func(enable types.Bool) v1.ModelsSettingsDNSForwardingSettingsEnabled {
				if enable.ValueBool() {
					return v1.ModelsSettingsDNSForwardingSettingsEnabledTrue
				}
				return v1.ModelsSettingsDNSForwardingSettingsEnabledFalse
			}(model.Enabled),
			PrivateHostedZone: model.PrivateHostedZone.ValueString(),
			UpstreamDNS1:      model.UpstreamDNS1.ValueString(),
			UpstreamDNS2:      model.UpstreamDNS2.ValueString(),
		},
	}
}

func waitForInstanceStatus(ctx context.Context, api seg.ServiceEndpointGatewayAPI, id string, status v1.ModelsInstanceInstanceStatus) error {
	withTimeout, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		ok, err := checkInstanceStatus(ctx, api, id, status)
		if err != nil {
			return err
		}
		if ok {
			return nil // desired status reached
		}
		select {
		case <-withTimeout.Done():
			return errors.New("timeout waiting for condition")
		case <-ticker.C:
			// retry
		}
	}
}

func checkInstanceStatus(ctx context.Context, api seg.ServiceEndpointGatewayAPI, id string,
	requestStatus v1.ModelsInstanceInstanceStatus) (bool, error) {
	resp, err := api.Read(ctx, id)
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, nil
	}
	currentStatus, set := resp.Appliance.Instance.Status.Get()
	if !set {
		return false, nil
	}
	return currentStatus == requestStatus, nil
}
