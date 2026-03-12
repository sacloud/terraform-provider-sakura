// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	c.ID = types.StringValue(uuid.UUID(d.ClusterID).String())
	c.Name = types.StringValue(d.Name)
	c.ServicePrincipalID = types.StringValue(d.ServicePrincipalID)
	c.HasLetsEncryptEmail = types.BoolValue(d.HasLetsEncryptEmail)
	c.CreatedAt = types.StringValue(time.Unix(common.ToInt64(d.Created), 0).Format(time.RFC822))

	ports := make([]portModel, len(d.Ports))
	for i, j := range d.Ports {
		ports[i].updateState(&j)
	}
	c.Ports = ports
}

func (c *clusterModel) clusterID() (ret v1.ClusterID, err error) {
	u, err := uuid.Parse(c.ID.ValueString())

	if err == nil {
		ret = v1.ClusterID(u)
	}

	return
}

////////////////////////////////////////////////////////////////

func (portModel) AttributeTypes() attrTypes    { return portAttrs }
func (clusterModel) AttributeTypes() attrTypes { return clusterAttrs }
