// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type apigwSubscriptionDataSource struct {
	client *v1.Client
}

func NewApigwSubscriptionDataSource() datasource.DataSource {
	return &apigwSubscriptionDataSource{}
}

var (
	_ datasource.DataSource              = &apigwSubscriptionDataSource{}
	_ datasource.DataSourceWithConfigure = &apigwSubscriptionDataSource{}
)

func (r *apigwSubscriptionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_subscription"
}

func (r *apigwSubscriptionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwSubscriptionDataSourceModel struct {
	apigwSubscriptionBaseModel
}

func (r *apigwSubscriptionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("API Gateway Subscription"),
			"name":       common.SchemaDataSourceName("API Gateway Subscription"),
			"created_at": schemaDataSourceAPIGWCreatedAt("API Gateway Subscription"),
			"updated_at": schemaDataSourceAPIGWUpdatedAt("API Gateway Subscription"),
			"plan_id": schema.StringAttribute{
				Computed:    true,
				Description: "Plan ID of the API Gateway Subscription",
			},
			"resource_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Resource ID of the API Gateway Subscription",
			},
			"monthly_request": schema.Int64Attribute{
				Computed:    true,
				Description: "Monthly request count of the API Gateway Subscription",
			},
			"service": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Service information of the API Gateway Subscription",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "ID of the API Gateway Service associated with the API Gateway Subscription",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "Name of the API Gateway Service associated with the API Gateway Subscription",
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing API Gateway Subscription.",
	}
}

func (r *apigwSubscriptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apigwSubscriptionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id string
	if utils.IsKnown(data.Name) {
		id = getAPIGWSubscriptionId(ctx, r.client, data.Name.ValueString(), &resp.Diagnostics)
		if id == "" {
			resp.Diagnostics.AddError("Read: Search Error", "API Gateway Subscription with the specified name was not found")
			return
		}
	} else {
		id = data.ID.ValueString()
	}

	sub := getAPIGWSubscriptionFromList(ctx, r.client, id, &resp.State, &resp.Diagnostics)
	if sub == nil {
		return
	}

	data.updateState(sub)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
