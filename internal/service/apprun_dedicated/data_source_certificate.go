// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cert "github.com/sacloud/apprun-dedicated-api-go/apis/certificate"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type certDataSource struct{ dataSourceClient }
type certDataSourceModel struct{ certModel }

var (
	_ datasource.DataSource              = &certDataSource{}
	_ datasource.DataSourceWithConfigure = &certDataSource{}
)

func NewCertDataSource() datasource.DataSource {
	return &certDataSource{dataSourceNamed("certificate")}
}

func (d *certDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	id := d.schemaID()

	name := d.schemaName()

	clusterID := d.schemaClusterID()

	commonName := schema.StringAttribute{
		Computed:    true,
		Description: "The common name of the certificate",
	}

	// SANはAPI上はリストで表現されている
	// が、X.509とRFC6125によると順序はない
	// Terraform上はSetであると考えるべきだろう
	subjectAlternativeNames := schema.SetAttribute{
		Computed:    true,
		ElementType: types.StringType,
		Description: "The subject alternative names of the certificate",
	}

	notBefore := schema.StringAttribute{
		Computed:    true,
		Description: "The certificate validity start time",
	}

	notAfter := schema.StringAttribute{
		Computed:    true,
		Description: "The certificate validity end time",
	}

	createdAt := schema.StringAttribute{
		Computed:    true,
		Description: "The creation timestamp of the certificate",
	}

	updatedAt := schema.StringAttribute{
		Computed:    true,
		Description: "The update timestamp of the certificate",
	}

	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated certificate",
		Attributes: map[string]schema.Attribute{
			"id":                        id,
			"name":                      name,
			"cluster_id":                clusterID,
			"common_name":               commonName,
			"subject_alternative_names": subjectAlternativeNames,
			"not_before":                notBefore,
			"not_after":                 notAfter,
			"created_at":                createdAt,
			"updated_at":                updatedAt,
		},
	}
}

func (d *certDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state certDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	var certID *certID
	var clusterID *clusterID
	var ds diag.Diagnostics

	if state.ID.IsNull() {
		clusterID, certID, ds = state.byName(ctx, d)
	} else {
		clusterID, certID, ds = state.byId(ctx, d)
	}
	res.Diagnostics.Append(ds...)

	if ds.HasError() {
		return
	}

	detail, err := d.api(*clusterID).Read(ctx, *certID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated certificate: %s", err))
		return
	}

	if detail == nil {
		common.FilterNoResultErr(&res.Diagnostics)
		return
	}

	res.Diagnostics.Append(state.updateState(ctx, detail, *clusterID)...)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (state *certDataSourceModel) byId(context.Context, *certDataSource) (*clusterID, *certID, diag.Diagnostics) {
	var d diag.Diagnostics
	certID, err := state.certID()

	if err != nil {
		d.AddError("Read: Invalid ID", fmt.Sprintf("failed to parse certificate ID: %s", err))
	}

	clusterID, err := state.clusterID()

	if err != nil {
		d.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
	}

	return &clusterID, &certID, d
}

func (state *certDataSourceModel) byName(ctx context.Context, r *certDataSource) (_ *clusterID, _ *certID, d diag.Diagnostics) {
	clusterID, err := state.clusterID()

	if err != nil {
		d.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse certificate ID: %s", err))
		return
	}

	api := r.api(clusterID)
	certs, err := listed(func(cursor *certID) ([]v1.ReadCertificate, *certID, error) {
		return api.List(ctx, 10, cursor)
	})

	if err != nil {
		d.AddError("Read: API Error", fmt.Sprintf("failed to list AppRun Dedicated certificates: %s", err))
		return
	}

	name := state.Name.ValueString()
	for _, i := range certs {
		if i.Name == name {
			return &clusterID, &i.CertificateID, d
		}
	}

	d.AddError("Read: API Error", fmt.Sprintf("certificate with name %q not found", name))
	return
}

func (d *certDataSource) api(clusterID clusterID) *cert.CertificateOp {
	return cert.NewCertificateOp(d.client, clusterID)
}
