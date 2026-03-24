// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

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

func (certModel) AttributeTypes() attrTypes            { return certAttrs }
func (c *certModel) certID() (v1.CertificateID, error) { return intoUUID[v1.CertificateID](c.ID) }
func (c *certModel) clusterID() (v1.ClusterID, error)  { return intoUUID[v1.ClusterID](c.ClusterID) }
