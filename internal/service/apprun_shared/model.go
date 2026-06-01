// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_shared

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
)

type apprunSharedBaseModel struct {
	ID             types.String                   `tfsdk:"id"`
	ResourceID     types.String                   `tfsdk:"resource_id"`
	Name           types.String                   `tfsdk:"name"`
	TimeoutSeconds types.Int32                    `tfsdk:"timeout_seconds"`
	Port           types.Int32                    `tfsdk:"port"`
	MinScale       types.Int32                    `tfsdk:"min_scale"`
	MaxScale       types.Int32                    `tfsdk:"max_scale"`
	Components     []*apprunSharedComponentModel  `tfsdk:"components"`
	PacketFilter   *apprunSharedPacketFilterModel `tfsdk:"packet_filter"`
	Status         types.String                   `tfsdk:"status"`
	PublicURL      types.String                   `tfsdk:"public_url"`
}

type apprunSharedComponentModel struct {
	Name         types.String                            `tfsdk:"name"`
	MaxCpu       types.String                            `tfsdk:"max_cpu"`
	MaxMemory    types.String                            `tfsdk:"max_memory"`
	DeploySource *apprunSharedComponentDeploySourceModel `tfsdk:"deploy_source"`
	Env          types.Set                               `tfsdk:"env"`
	Probe        types.Object                            `tfsdk:"probe"`
}

type apprunSharedComponentDeploySourceModel struct {
	ContainerRegistry *apprunSharedComponentContainerRegistryModel `tfsdk:"container_registry"`
}

type apprunSharedComponentContainerRegistryModel struct {
	Image             types.String `tfsdk:"image"`
	Server            types.String `tfsdk:"server"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int32  `tfsdk:"password_wo_version"`
}

type apprunSharedComponentEnvModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (m apprunSharedComponentEnvModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}

type apprunSharedProbeModel struct {
	HttpGet *apprunSharedProbeHttpGetModel `tfsdk:"http_get"`
}

func (m apprunSharedProbeModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"http_get": types.ObjectType{AttrTypes: apprunSharedProbeHttpGetModel{}.AttributeTypes()},
	}
}

type apprunSharedProbeHttpGetModel struct {
	Path    types.String `tfsdk:"path"`
	Port    types.Int32  `tfsdk:"port"`
	Headers types.Set    `tfsdk:"headers"`
}

func (m apprunSharedProbeHttpGetModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"path":    types.StringType,
		"port":    types.Int32Type,
		"headers": types.SetType{ElemType: types.ObjectType{AttrTypes: apprunSharedProbeHttpGetHeaderModel{}.AttributeTypes()}},
	}
}

type apprunSharedProbeHttpGetHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (m apprunSharedProbeHttpGetHeaderModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	}
}

type apprunSharedPacketFilterModel struct {
	Enabled  types.Bool                               `tfsdk:"enabled"`
	Settings []*apprunSharedPacketFilterSettingsModel `tfsdk:"settings"`
}

func (m apprunSharedPacketFilterModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":  types.BoolType,
		"settings": types.ListType{ElemType: types.ObjectType{AttrTypes: apprunSharedPacketFilterSettingsModel{}.AttributeTypes()}},
	}
}

type apprunSharedPacketFilterSettingsModel struct {
	FromIP             types.String `tfsdk:"from_ip"`
	FromIPPrefixLength types.Int32  `tfsdk:"from_ip_prefix_length"`
}

func (m apprunSharedPacketFilterSettingsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"from_ip":               types.StringType,
		"from_ip_prefix_length": types.Int32Type,
	}
}

func stringValueFromOptString(value v1.OptString) types.String {
	if v, ok := value.Get(); ok {
		return types.StringValue(v)
	}
	return types.StringNull()
}

func (model *apprunSharedBaseModel) updateState(application *v1.HandlerGetApplication, pf *v1.HandlerGetPacketFilter) {
	model.ID = types.StringValue(application.ID)
	model.ResourceID = types.StringValue(application.ResourceID)
	model.Name = types.StringValue(application.Name)
	model.TimeoutSeconds = types.Int32Value(int32(application.TimeoutSeconds))
	model.Port = types.Int32Value(int32(application.Port))
	model.MinScale = types.Int32Value(int32(application.MinScale))
	model.MaxScale = types.Int32Value(int32(application.MaxScale))
	model.Components = flattenApprunApplicationComponents(model, application, true)
	model.PacketFilter = flattenApprunPacketFilter(pf)
	model.Status = types.StringValue(string(application.Status))
	model.PublicURL = types.StringValue(application.PublicURL)
}

func flattenApprunApplicationComponents(model *apprunSharedBaseModel, application *v1.HandlerGetApplication, includePassword bool) []*apprunSharedComponentModel {
	var results []*apprunSharedComponentModel

	for _, c := range application.Components {
		containerRegistry := &apprunSharedComponentContainerRegistryModel{}
		if cr, ok := c.DeploySource.ContainerRegistry.Get(); ok {
			containerRegistry.Image = types.StringValue(cr.Image)
			containerRegistry.Server = stringValueFromOptString(cr.Server)
			containerRegistry.Username = stringValueFromOptString(cr.Username)
		}

		result := &apprunSharedComponentModel{
			Name:      types.StringValue(c.Name),
			MaxCpu:    types.StringValue(c.MaxCPU),
			MaxMemory: types.StringValue(c.MaxMemory),
			DeploySource: &apprunSharedComponentDeploySourceModel{
				ContainerRegistry: containerRegistry,
			},
			Env:   flattenApprunApplicationEnvs(&c),
			Probe: flattenApprunApplicationProbe(&c),
		}

		if includePassword {
			// NOTE:
			// v1.Applicationはcontainer_registryのpasswordが含まれないため、そのままだとtfstateに空文字列がセットされてしまう。
			// この場合resourceにpasswordの定義があると、resourceを変更していなくてもterraform planでdiffが出てしまう。
			// この対策として、passwordのみschema.ResourceDataからデータを参照してセットするようにする。
			for _, exComponent := range model.Components {
				if exComponent.Name.ValueString() == c.Name && exComponent.DeploySource.ContainerRegistry != nil {
					if exComponent.DeploySource.ContainerRegistry.PasswordWOVersion.ValueInt32() > 0 {
						result.DeploySource.ContainerRegistry.PasswordWOVersion = types.Int32Value(exComponent.DeploySource.ContainerRegistry.PasswordWOVersion.ValueInt32())
						break
					}
					if exComponent.DeploySource.ContainerRegistry.Password.ValueString() != "" {
						result.DeploySource.ContainerRegistry.Password = types.StringValue(exComponent.DeploySource.ContainerRegistry.Password.ValueString())
					}
					break
				}
			}
		}

		results = append(results, result)
	}
	return results
}

func flattenApprunApplicationEnvs(component *v1.HandlerGetApplicationComponentsItem) types.Set {
	if len(component.Env) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: apprunSharedComponentEnvModel{}.AttributeTypes()})
	}

	var results []apprunSharedComponentEnvModel
	for _, e := range component.Env {
		results = append(results, apprunSharedComponentEnvModel{
			Key:   stringValueFromOptString(e.Key),
			Value: stringValueFromOptString(e.Value),
		})
	}

	return toTSet(apprunSharedComponentEnvModel{}.AttributeTypes(), results)
}

func toTSet(elemType map[string]attr.Type, elem any) types.Set {
	r, _ := types.SetValueFrom(context.Background(), types.ObjectType{AttrTypes: elemType}, elem)
	return r
}

func flattenApprunApplicationProbe(component *v1.HandlerGetApplicationComponentsItem) types.Object {
	v := types.ObjectNull(apprunSharedProbeModel{}.AttributeTypes())
	probe, ok := component.Probe.Get()
	if ok {
		httpGet, ok := probe.HTTPGet.Get()
		if !ok {
			return v
		}
		m := apprunSharedProbeModel{
			HttpGet: &apprunSharedProbeHttpGetModel{
				Path:    types.StringValue(httpGet.Path),
				Port:    types.Int32Value(int32(httpGet.Port)),
				Headers: flattenApprunApplicationProbeHttpGetHeaders(&httpGet),
			},
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	}
	return v
}

func flattenApprunApplicationProbeHttpGetHeaders(httpGet *v1.HandlerGetApplicationComponentsItemProbeHTTPGet) types.Set {
	if len(httpGet.Headers) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: apprunSharedProbeHttpGetHeaderModel{}.AttributeTypes()})
	}

	var results []apprunSharedProbeHttpGetHeaderModel
	for _, h := range httpGet.Headers {
		results = append(results, apprunSharedProbeHttpGetHeaderModel{
			Name:  stringValueFromOptString(h.Name),
			Value: stringValueFromOptString(h.Value),
		})
	}

	return toTSet(apprunSharedProbeHttpGetHeaderModel{}.AttributeTypes(), results)
}

func flattenApprunPacketFilter(packetFilter *v1.HandlerGetPacketFilter) *apprunSharedPacketFilterModel {
	if packetFilter == nil || (!packetFilter.IsEnabled && len(packetFilter.Settings) == 0) {
		return nil
	}

	return &apprunSharedPacketFilterModel{
		Enabled:  types.BoolValue(packetFilter.IsEnabled),
		Settings: flattenApprunPacketFilterSettings(packetFilter.Settings),
	}
}

func flattenApprunPacketFilterSettings(settings []v1.HandlerGetPacketFilterSettingsItem) []*apprunSharedPacketFilterSettingsModel {
	var results []*apprunSharedPacketFilterSettingsModel
	for _, s := range settings {
		results = append(results, &apprunSharedPacketFilterSettingsModel{
			FromIP:             types.StringValue(s.FromIP),
			FromIPPrefixLength: types.Int32Value(int32(s.FromIPPrefixLength)),
		})
	}
	return results
}
