// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/apprun-dedicated-api-go/apis/cluster"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type portModel struct {
	Port     types.Int32  `tfsdk:"port"`
	Protocol types.String `tfsdk:"protocol"`
}

type clusterModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Ports               []portModel  `tfsdk:"ports"`
	ServicePrincipalID  types.String `tfsdk:"service_principal_id"`
	HasLetsEncryptEmail types.Bool   `tfsdk:"has_lets_encrypt_email"`
	CreatedAt           types.String `tfsdk:"created_at"`
}

var portAttrs = attrTypes{
	"port":     types.Int32Type,
	"protocol": types.StringType,
}

var clusterAttrs = attrTypes{
	"id":                     types.StringType,
	"name":                   types.StringType,
	"ports":                  types.ListType{ElemType: types.ObjectType{AttrTypes: portAttrs}},
	"service_principal_id":   types.StringType,
	"has_lets_encrypt_email": types.BoolType,
	"created_at":             types.Int64Type,
}

func (p *portModel) updateState(q *v1.ReadLoadBalancerPort) {
	p.Port = types.Int32Value(common.ToInt32(q.GetPort()))
	p.Protocol = types.StringValue(common.ToString(q.GetProtocol()))
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

func (portModel) AttributeTypes() attrTypes              { return portAttrs }
func (clusterModel) AttributeTypes() attrTypes           { return clusterAttrs }
func (c *clusterModel) clusterID() (v1.ClusterID, error) { return intoUUID[v1.ClusterID](c.ID) }
