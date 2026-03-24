// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/apprun-dedicated-api-go/apis/version"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type healthCheckModel struct {
	Path            types.String `tfsdk:"path"`
	IntervalSeconds types.Int32  `tfsdk:"interval_seconds"`
	TimeoutSeconds  types.Int32  `tfsdk:"timeout_seconds"`
}

type envVarModel struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Secret types.Bool   `tfsdk:"secret"`
}

type exposedPortModel struct {
	TargetPort       types.Int32       `tfsdk:"target_port"`
	LoadBalancerPort types.Int32       `tfsdk:"load_balancer_port"`
	UseLetsEncrypt   types.Bool        `tfsdk:"use_lets_encrypt"`
	Host             types.Set         `tfsdk:"host"`
	HealthCheck      *healthCheckModel `tfsdk:"health_check"`
}

type verModel struct {
	Version           types.Int32        `tfsdk:"version"`
	ApplicationID     types.String       `tfsdk:"application_id"`
	CPU               types.Int64        `tfsdk:"cpu"`
	Memory            types.Int64        `tfsdk:"memory"`
	ScalingMode       types.String       `tfsdk:"scaling_mode"`
	FixedScale        types.Int32        `tfsdk:"fixed_scale"`
	MinScale          types.Int32        `tfsdk:"min_scale"`
	MaxScale          types.Int32        `tfsdk:"max_scale"`
	ScaleInThreshold  types.Int32        `tfsdk:"scale_in_threshold"`
	ScaleOutThreshold types.Int32        `tfsdk:"scale_out_threshold"`
	Image             types.String       `tfsdk:"image"`
	Cmd               types.List         `tfsdk:"cmd"`
	RegistryUsername  types.String       `tfsdk:"registry_username"`
	ActiveNodeCount   types.Int64        `tfsdk:"active_node_count"`
	CreatedAt         types.String       `tfsdk:"created_at"`
	ExposedPorts      []exposedPortModel `tfsdk:"exposed_ports"`
	EnvVars           []envVarModel      `tfsdk:"env_vars"`
}

var healthCheckAttrs = attrTypes{
	"path":             types.StringType,
	"interval_seconds": types.Int32Type,
	"timeout_seconds":  types.Int32Type,
}

var envVarAttrs = attrTypes{
	"key":    types.StringType,
	"value":  types.StringType,
	"secret": types.BoolType,
}

var exposedPortAttrs = attrTypes{
	"target_port":        types.Int32Type,
	"load_balancer_port": types.Int32Type,
	"use_lets_encrypt":   types.BoolType,
	"host":               types.SetType{ElemType: types.StringType},
	"health_check":       types.ObjectType{AttrTypes: healthCheckAttrs},
}
var versionAttrs = attrTypes{
	"version":                  types.Int32Type,
	"application_id":           types.StringType,
	"cpu":                      types.Int64Type,
	"memory":                   types.Int64Type,
	"scaling_mode":             types.StringType,
	"fixed_scale":              types.Int32Type,
	"min_scale":                types.Int32Type,
	"max_scale":                types.Int32Type,
	"scale_in_threshold":       types.Int32Type,
	"scale_out_threshold":      types.Int32Type,
	"image":                    types.StringType,
	"cmd":                      types.ListType{ElemType: types.StringType},
	"registry_username":        types.StringType,
	"registry_password":        types.StringType,
	"registry_password_action": types.StringType,
	"active_node_count":        types.Int64Type,
	"created_at":               types.StringType,
	"exposed_ports":            types.SetType{ElemType: types.ObjectType{AttrTypes: exposedPortAttrs}},
	"env_vars":                 types.SetType{ElemType: types.ObjectType{AttrTypes: envVarAttrs}},
}

func (healthCheckModel) AttributeTypes() attrTypes { return healthCheckAttrs }
func (envVarModel) AttributeTypes() attrTypes      { return envVarAttrs }
func (exposedPortModel) AttributeTypes() attrTypes { return exposedPortAttrs }
func (verModel) AttributeTypes() attrTypes         { return versionAttrs }

func (v *verModel) applicationID() (v1.ApplicationID, error) {
	return intoUUID[v1.ApplicationID](v.ApplicationID)
}

func (v *verModel) versionNumber() v1.ApplicationVersionNumber {
	return v1.ApplicationVersionNumber(v.Version.ValueInt32())
}

func (h healthCheckModel) intoCreate() (ret v1.HealthCheck) {
	ret.Path = h.Path.ValueString()
	ret.IntervalSeconds = h.IntervalSeconds.ValueInt32()
	ret.TimeoutSeconds = h.TimeoutSeconds.ValueInt32()

	return
}

func (e envVarModel) intoCreate() (ret version.EnvironmentVariable) {
	ret.Key = e.Key.ValueString()
	ret.Value = e.Value.ValueStringPointer()
	ret.Secret = e.Secret.ValueBool()

	return
}
func (p exposedPortModel) intoCreate() (ret version.ExposedPort) {
	ret.TargetPort = v1.Port(p.TargetPort.ValueInt32())
	ret.LoadBalancerPort = saclient.Ptr(v1.Port(p.LoadBalancerPort.ValueInt32()))
	ret.UseLetsEncrypt = p.UseLetsEncrypt.ValueBool()
	ret.Host = common.TsetToStrings(p.Host)
	ret.HealthCheck = saclient.Ptr(p.HealthCheck.intoCreate())

	return
}

func (h *healthCheckModel) updateState(d *v1.HealthCheck) {
	h.Path = types.StringValue(d.GetPath())
	h.IntervalSeconds = types.Int32Value(d.GetIntervalSeconds())
	h.TimeoutSeconds = types.Int32Value(d.GetTimeoutSeconds())
}

func (e *envVarModel) updateState(d version.EnvironmentVariable) {
	e.Key = types.StringValue(d.Key)
	e.Value = types.StringPointerValue(d.Value)
	e.Secret = types.BoolValue(d.Secret)
}

func (*exposedPortModel) int32(i16 *v1.Port) types.Int32 {
	var i32 int32

	if i16 == nil {
		return types.Int32Null()
	}

	i32 = int32(*i16)

	return types.Int32Value(i32)
}

func (p *exposedPortModel) updateState(ctx context.Context, d *version.ExposedPort) {
	p.TargetPort = p.int32(&d.TargetPort)
	p.LoadBalancerPort = p.int32(d.LoadBalancerPort)
	p.UseLetsEncrypt = types.BoolValue(d.UseLetsEncrypt)
	p.Host = common.StringsToTset(d.Host)
	p.HealthCheck = new(healthCheckModel)
	p.HealthCheck.updateState(d.HealthCheck)
}

func (v *verModel) updateState(ctx context.Context, d *version.VersionDetail, aid v1.ApplicationID) (ret diag.Diagnostics) {
	v.Version = types.Int32Value(common.ToInt32(d.Version))
	v.ApplicationID = uuid2StringValue(aid)
	v.CPU = types.Int64Value(d.CPU)
	v.Memory = types.Int64Value(d.Memory)
	v.ScalingMode = types.StringValue(common.ToString(d.ScalingMode))
	v.FixedScale = types.Int32PointerValue(d.FixedScale)
	v.MinScale = types.Int32PointerValue(d.MinScale)
	v.MaxScale = types.Int32PointerValue(d.MaxScale)
	v.ScaleInThreshold = types.Int32PointerValue(d.ScaleInThreshold)
	v.ScaleOutThreshold = types.Int32PointerValue(d.ScaleOutThreshold)
	v.Image = types.StringValue(d.Image)
	v.RegistryUsername = types.StringPointerValue(d.RegistryUsername)
	// password is write only
	v.ActiveNodeCount = types.Int64Value(d.ActiveNodeCount)
	v.CreatedAt = intoRFC2822(d.Created)

	v.Cmd, ret = types.ListValueFrom(ctx, types.StringType, common.MapTo(d.Cmd, types.StringValue))

	if ret.HasError() {
		return
	}

	v.ExposedPorts = common.MapTo(d.ExposedPorts, func(p version.ExposedPort) (ret exposedPortModel) {
		ret.updateState(ctx, &p)
		return
	})

	v.EnvVars = common.MapTo(d.EnvVars, func(src version.EnvironmentVariable) (dst envVarModel) {
		dst.updateState(src)
		return
	})

	return
}
