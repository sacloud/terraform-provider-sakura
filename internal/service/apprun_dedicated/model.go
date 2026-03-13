// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/apprun-dedicated-api-go/apis/cluster"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type attrTypes = map[string]attr.Type

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

func (portModel) AttributeTypes() attrTypes        { return portAttrs }
func (clusterModel) AttributeTypes() attrTypes     { return clusterAttrs }
func (certModel) AttributeTypes() attrTypes        { return certAttrs }
