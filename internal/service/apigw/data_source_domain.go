// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwDomainDataSource struct {
	client *v1.Client
}

func NewApigwDomainDataSource() datasource.DataSource {
	return &apigwDomainDataSource{}
}

var (
	_ datasource.DataSource              = &apigwDomainDataSource{}
	_ datasource.DataSourceWithConfigure = &apigwDomainDataSource{}
)

func (r *apigwDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_domain"
}

func (r *apigwDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwDomainDataSourceModel struct {
	apigwDomainBaseModel
}

func (r *apigwDomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaDataSourceId("API Gateway Domain"),
			"name":       common.SchemaDataSourceName("API Gateway Domain"),
			"created_at": schemaDataSourceAPIGWCreatedAt("API Gateway Domain"),
			"updated_at": schemaDataSourceAPIGWUpdatedAt("API Gateway Domain"),
			"certificate_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the API Gateway Certificate",
			},
			"certificate_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the API Gateway Certificate",
			},
		},
		MarkdownDescription: "Get information about an existing API Gateway Domain.",
	}
}

func (r *apigwDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apigwDomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := getAPIGWDomain(ctx, r.client, data.ID.ValueString(), data.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if domain == nil {
		return
	}

	data.updateState(domain)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
