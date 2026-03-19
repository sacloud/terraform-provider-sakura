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

func (portModel) AttributeTypes() attrTypes    { return portAttrs }
func (clusterModel) AttributeTypes() attrTypes { return clusterAttrs }
func (certModel) AttributeTypes() attrTypes    { return certAttrs }
func (appModel) AttributeTypes() attrTypes     { return applicationAttrs }
