// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

	var certID *v1.CertificateID
	var clusterID *v1.ClusterID

	if state.ID.IsNull() {
		// Lookup by name
		clusterID, certID = d.byName(ctx, req, res, &state)
	} else {
		// Lookup by ID
		clusterID, certID = d.byID(ctx, req, res, &state)
	}

	if certID == nil || clusterID == nil {
		return
	}

	api := d.api(*clusterID)
	detail, err := api.Read(ctx, *certID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated certificate: %s", err))
		return
	}

	if detail == nil {
		common.FilterNoResultErr(&res.Diagnostics)
		return
	}

	state.updateState(ctx, detail, *clusterID)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (d *certDataSource) byID(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse, state *certDataSourceModel) (*v1.ClusterID, *v1.CertificateID) {
	certID, err := state.certID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid ID", fmt.Sprintf("failed to parse certificate ID: %s", err))
	}

	clusterID, err := state.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
	}

	return &clusterID, &certID
}

func (d *certDataSource) byName(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse, state *certDataSourceModel) (*v1.ClusterID, *v1.CertificateID) {
	clusterID, err := state.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return nil, nil
	}

	api := d.api(clusterID)
	certs, err := listed(func(cursor *v1.CertificateID) ([]v1.ReadCertificate, *v1.CertificateID, error) {
		return api.List(ctx, 10, cursor)
	})

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list AppRun Dedicated certificates: %s", err))
		return nil, nil
	}

	name := state.Name.ValueString()
	for _, i := range certs {
		if i.Name == name {
			return &clusterID, &i.CertificateID
		}
	}

	res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("certificate with name %q not found", name))
	return nil, nil
}

func (d *certDataSource) api(clusterID v1.ClusterID) *cert.CertificateOp {
	return cert.NewCertificateOp(d.client, clusterID)
}
