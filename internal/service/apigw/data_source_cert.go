// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwCertDataSource struct {
	client *v1.Client
}

func NewApigwCertDataSource() datasource.DataSource {
	return &apigwCertDataSource{}
}

var (
	_ datasource.DataSource              = &apigwCertDataSource{}
	_ datasource.DataSourceWithConfigure = &apigwCertDataSource{}
)

func (r *apigwCertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_cert"
}

func (r *apigwCertDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwCertDataSourceModel struct {
	apigwCertBaseModel
	RSA   *apigwCertCertDataSourceModel `tfsdk:"rsa"`
	ECDSA *apigwCertCertDataSourceModel `tfsdk:"ecdsa"`
}

type apigwCertCertDataSourceModel struct {
	ExpiredAt types.String `tfsdk:"expired_at"`
}

func (r *apigwCertDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaDataSourceId("API Gateway Certificate"),
			"name":       common.SchemaDataSourceName("API Gateway Certificate"),
			"created_at": schemaDataSourceAPIGWCreatedAt("API Gateway Certificate"),
			"updated_at": schemaDataSourceAPIGWUpdatedAt("API Gateway Certificate"),
			"rsa": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "RSA settings of the API Gateway Certificate",
				Attributes: map[string]schema.Attribute{
					"expired_at": schema.StringAttribute{
						Computed:    true,
						Description: "The expiration timestamp",
					},
				},
			},
			"ecdsa": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "ECDSA settings of the API Gateway Certificate",
				Attributes: map[string]schema.Attribute{
					"expired_at": schema.StringAttribute{
						Computed:    true,
						Description: "The expiration timestamp",
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing API Gateway Certificate.",
	}
}

func (r *apigwCertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apigwCertDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert := getAPIGWCert(ctx, r.client, data.ID.ValueString(), data.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if cert == nil {
		return
	}

	data.updateState(cert)
	flattenCertsDataSource(&data, cert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenCertsDataSource(model *apigwCertDataSourceModel, cert *v1.Certificate) {
	if cert.Rsa.IsSet() {
		model.RSA = &apigwCertCertDataSourceModel{
			ExpiredAt: types.StringValue(cert.Rsa.Value.ExpiredAt.Value.String()),
		}
	}
	if cert.Ecdsa.IsSet() {
		model.ECDSA = &apigwCertCertDataSourceModel{
			ExpiredAt: types.StringValue(cert.Ecdsa.Value.ExpiredAt.Value.String()),
		}
	}
}
