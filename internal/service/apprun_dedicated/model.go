// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	app "github.com/sacloud/apprun-dedicated-api-go/apis/application"
	"github.com/sacloud/apprun-dedicated-api-go/apis/cluster"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/apprun-dedicated-api-go/apis/version"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type attrTypes = map[string]attr.Type

////////////////////////////////////////////////////////////////

type dataSourceClient struct {
	client *v1.Client
	name   string
}

func dataSourceNamed(name string) dataSourceClient { return dataSourceClient{name: name} }

func (d *dataSourceClient) Configure(_ context.Context, req datasource.ConfigureRequest, res *datasource.ConfigureResponse) {
	client := common.GetApiClientFromProvider(req.ProviderData, &res.Diagnostics)

	if client == nil {
		return
	}

	d.client = client.AppRunDedicatedClient
}

func (d *dataSourceClient) Metadata(_ context.Context, req datasource.MetadataRequest, res *datasource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_apprun_dedicated_" + d.name
}

func (d *dataSourceClient) schemaID() (ret dschema.StringAttribute) {
	ret = common.SchemaDataSourceId(d.name).(dschema.StringAttribute)
	ret.Required = false
	ret.Optional = true
	ret.Computed = true
	ret.Validators = []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		sacloudvalidator.UUIDValidator,
	}

	return
}

func (d *dataSourceClient) schemaName() (ret dschema.StringAttribute) {
	ret = common.SchemaDataSourceName(d.name).(dschema.StringAttribute)
	ret.Required = false
	ret.Optional = true
	ret.Computed = true
	ret.Validators = []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}

	return
}

func (*dataSourceClient) schemaClusterID() dschema.StringAttribute {
	return dschema.StringAttribute{
		Required:    true,
		Description: "The cluster ID that the certificate belongs to",
		Validators:  []validator.String{sacloudvalidator.UUIDValidator},
	}
}

func (d *dataSourceClient) schemaCreatedAt() dschema.StringAttribute {
	return common.SchemaDataSourceCreatedAt(d.name).(dschema.StringAttribute)
}

////////////////////////////////////////////////////////////////

type resourceClient struct {
	client *v1.Client
	name   string
}

func resourceNamed(name string) resourceClient { return resourceClient{name: name} }

func (r *resourceClient) Configure(_ context.Context, req resource.ConfigureRequest, res *resource.ConfigureResponse) {
	client := common.GetApiClientFromProvider(req.ProviderData, &res.Diagnostics)

	if client == nil {
		return
	}

	r.client = client.AppRunDedicatedClient
}

func (r *resourceClient) Metadata(_ context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_apprun_dedicated_" + r.name
}

func (*resourceClient) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, res)
}

func (r *resourceClient) schemaID() rschema.StringAttribute {
	return common.SchemaResourceId(r.name).(rschema.StringAttribute)
}

func (r *resourceClient) schemaName(validators ...validator.String) (ret rschema.StringAttribute) {
	ret = common.SchemaResourceName(r.name).(rschema.StringAttribute)
	ret.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	ret.Validators = append([]validator.String{stringvalidator.LengthBetween(1, 20)}, validators...)

	return
}

func (r *resourceClient) schemaClusterID() (ret rschema.StringAttribute) {
	ret = common.SchemaResourceId("cluster").(rschema.StringAttribute)
	ret.Computed = false
	ret.Required = true
	ret.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	ret.Validators = []validator.String{sacloudvalidator.UUIDValidator}

	return
}

func (r *resourceClient) schemaCreatedAt() rschema.StringAttribute {
	return common.SchemaResourceCreatedAt(r.name).(rschema.StringAttribute)
}

////////////////////////////////////////////////////////////////

type portModel struct {
	Port     types.Int32  `tfsdk:"port"`
	Protocol types.String `tfsdk:"protocol"`
}

var portAttrs = attrTypes{
	"port":     types.Int32Type,
	"protocol": types.StringType,
}

func (p *portModel) updateState(q *v1.ReadLoadBalancerPort) {
	p.Port = types.Int32Value(int32(q.GetPort()))
	p.Protocol = types.StringValue(string(q.GetProtocol()))
}

////////////////////////////////////////////////////////////////

type clusterModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Ports               []portModel  `tfsdk:"ports"`
	ServicePrincipalID  types.String `tfsdk:"service_principal_id"`
	HasLetsEncryptEmail types.Bool   `tfsdk:"has_lets_encrypt_email"`
	CreatedAt           types.String `tfsdk:"created_at"`
}

var clusterAttrs = attrTypes{
	"id":                     types.StringType,
	"name":                   types.StringType,
	"ports":                  types.ListType{ElemType: types.ObjectType{AttrTypes: portAttrs}},
	"service_principal_id":   types.StringType,
	"has_lets_encrypt_email": types.BoolType,
	"created_at":             types.Int64Type,
}

func (c *clusterModel) updateState(d *cluster.ClusterDetail) {
	c.ID = uuid2StringValue(d.ClusterID)
	c.Name = types.StringValue(d.Name)
	c.ServicePrincipalID = types.StringValue(d.ServicePrincipalID)
	c.HasLetsEncryptEmail = types.BoolValue(d.HasLetsEncryptEmail)
	c.CreatedAt = intoRFC2822(d.Created)
	c.Ports = common.MapTo(d.Ports, func(p v1.ReadLoadBalancerPort) (q portModel) {
		q.updateState(&p)
		return
	})
}

func (c *clusterModel) clusterID() (v1.ClusterID, error) { return intoUUID[v1.ClusterID](c.ID) }

////////////////////////////////////////////////////////////////

type certModel struct {
	ID        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	Name      types.String `tfsdk:"name"`
	CN        types.String `tfsdk:"common_name"`
	SAN       types.Set    `tfsdk:"subject_alternative_names"`
	NotBefore types.String `tfsdk:"not_before"`
	NotAfter  types.String `tfsdk:"not_after"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

var certAttrs = attrTypes{
	"id":                        types.StringType,
	"cluster_id":                types.StringType,
	"name":                      types.StringType,
	"common_name":               types.StringType,
	"subject_alternative_names": types.SetType{ElemType: types.StringType},
	"not_before":                types.StringType,
	"not_after":                 types.StringType,
	"created_at":                types.StringType,
	"updated_at":                types.StringType,
}

func (c *certModel) updateState(ctx context.Context, d *v1.ReadCertificate, clusterID v1.ClusterID) (ret diag.Diagnostics) {
	c.ID = uuid2StringValue(d.CertificateID)
	c.ClusterID = uuid2StringValue(clusterID)
	c.Name = types.StringValue(d.Name)
	c.CN = types.StringValue(d.CommonName)
	c.NotBefore = intoRFC2822(d.NotBeforeSec)
	c.NotAfter = intoRFC2822(d.NotAfterSec)
	c.CreatedAt = intoRFC2822(d.Created)
	c.UpdatedAt = intoRFC2822(d.Updated)
	c.SAN, ret = types.SetValueFrom(ctx, types.StringType, common.MapTo(d.SubjectAlternativeNames, types.StringValue))

	return
}

func (c *certModel) certID() (v1.CertificateID, error) { return intoUUID[v1.CertificateID](c.ID) }
func (c *certModel) clusterID() (v1.ClusterID, error)  { return intoUUID[v1.ClusterID](c.ClusterID) }

////////////////////////////////////////////////////////////////

type appModel struct {
	ID                     types.String `tfsdk:"id"`
	ClusterID              types.String `tfsdk:"cluster_id"`
	Name                   types.String `tfsdk:"name"`
	ClusterName            types.String `tfsdk:"cluster_name"`
	ActiveVersion          types.Int32  `tfsdk:"active_version"`
	DesiredCount           types.Int32  `tfsdk:"desired_count"`
	ScalingCooldownSeconds types.Int32  `tfsdk:"scaling_cooldown_seconds"`
}

var applicationAttrs = attrTypes{
	"id":                       types.StringType,
	"cluster_id":               types.StringType,
	"name":                     types.StringType,
	"cluster_name":             types.StringType,
	"active_version":           types.Int32Type,
	"desired_count":            types.Int32Type,
	"scaling_cooldown_seconds": types.Int32Type,
}

func (a *appModel) updateState(ctx context.Context, d *app.ApplicationDetail) (ret diag.Diagnostics) {
	a.ID = uuid2StringValue(d.ApplicationID)
	a.ClusterID = uuid2StringValue(d.ClusterID)
	a.Name = types.StringValue(d.Name)
	a.ClusterName = types.StringValue(d.ClusterName)
	a.ActiveVersion = types.Int32PointerValue(d.ActiveVersion)
	a.DesiredCount = types.Int32PointerValue(d.DesiredCount)
	a.ScalingCooldownSeconds = types.Int32Value(d.ScalingCooldownSeconds)

	return
}

func (a *appModel) applicationID() (v1.ApplicationID, error) { return intoUUID[v1.ApplicationID](a.ID) }
func (a *appModel) clusterID() (v1.ClusterID, error)         { return intoUUID[v1.ClusterID](a.ClusterID) }

////////////////////////////////////////////////////////////////

type healthCheckModel struct {
	Path            types.String `tfsdk:"path"`
	IntervalSeconds types.Int32  `tfsdk:"interval_seconds"`
	TimeoutSeconds  types.Int32  `tfsdk:"timeout_seconds"`
}

var healthCheckAttrs = attrTypes{
	"path":             types.StringType,
	"interval_seconds": types.Int32Type,
	"timeout_seconds":  types.Int32Type,
}

func (h *healthCheckModel) updateState(d *v1.HealthCheck) {
	h.Path = types.StringValue(d.GetPath())
	h.IntervalSeconds = types.Int32Value(d.GetIntervalSeconds())
	h.TimeoutSeconds = types.Int32Value(d.GetTimeoutSeconds())
}

func (h healthCheckModel) intoCreate() (ret v1.HealthCheck) {
	ret.Path = h.Path.ValueString()
	ret.IntervalSeconds = h.IntervalSeconds.ValueInt32()
	ret.TimeoutSeconds = h.TimeoutSeconds.ValueInt32()

	return
}

////////////////////////////////////////////////////////////////

type envVarModel struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Secret types.Bool   `tfsdk:"secret"`
}

var envVarAttrs = attrTypes{
	"key":    types.StringType,
	"value":  types.StringType,
	"secret": types.BoolType,
}

func (e *envVarModel) updateState(d v1.ReadEnvironmentVariable) {
	e.Key = types.StringValue(d.GetKey())
	e.Value = types.StringValue("")
	if !d.Value.Null {
		e.Value = types.StringValue(d.Value.Value)
	}
	e.Secret = types.BoolValue(d.GetSecret())
}

func (e envVarModel) intoCreate() (ret version.EnvironmentVariable) {
	ret.Key = e.Key.ValueString()
	ret.Value = e.Value.ValueStringPointer()
	ret.Secret = e.Secret.ValueBool()

	return
}

////////////////////////////////////////////////////////////////

type exposedPortModel struct {
	TargetPort       types.Int32       `tfsdk:"target_port"`
	LoadBalancerPort types.Int32       `tfsdk:"load_balancer_port"`
	UseLetsEncrypt   types.Bool        `tfsdk:"use_lets_encrypt"`
	Host             types.Set         `tfsdk:"host"`
	HealthCheck      *healthCheckModel `tfsdk:"health_check"`
}

var exposedPortAttrs = attrTypes{
	"target_port":        types.Int32Type,
	"load_balancer_port": types.Int32Type,
	"use_lets_encrypt":   types.BoolType,
	"host":               types.SetType{ElemType: types.StringType},
	"health_check":       types.ObjectType{AttrTypes: healthCheckAttrs},
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

func (p exposedPortModel) intoCreate() (ret version.ExposedPort) {
	ret.TargetPort = v1.Port(p.TargetPort.ValueInt32())
	ret.LoadBalancerPort = saclient.Ptr(v1.Port(p.LoadBalancerPort.ValueInt32()))
	ret.UseLetsEncrypt = p.UseLetsEncrypt.ValueBool()
	ret.Host = common.TsetToStrings(p.Host)
	ret.HealthCheck = saclient.Ptr(p.HealthCheck.intoCreate())

	return
}

////////////////////////////////////////////////////////////////

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

	return
}

func (v *verModel) applicationID() (v1.ApplicationID, error) {
	return intoUUID[v1.ApplicationID](v.ApplicationID)
}
func (v *verModel) versionNumber() v1.ApplicationVersionNumber {
	return v1.ApplicationVersionNumber(v.Version.ValueInt32())
}

////////////////////////////////////////////////////////////////

func (portModel) AttributeTypes() attrTypes        { return portAttrs }
func (clusterModel) AttributeTypes() attrTypes     { return clusterAttrs }
func (certModel) AttributeTypes() attrTypes        { return certAttrs }
func (appModel) AttributeTypes() attrTypes         { return applicationAttrs }
func (verModel) AttributeTypes() attrTypes         { return versionAttrs }
func (healthCheckModel) AttributeTypes() attrTypes { return healthCheckAttrs }
func (exposedPortModel) AttributeTypes() attrTypes { return exposedPortAttrs }
func (envVarModel) AttributeTypes() attrTypes      { return envVarAttrs }
