// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwPlanDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &apigwPlanDataSource{}
	_ datasource.DataSourceWithConfigure = &apigwPlanDataSource{}
)

func NewApigwPlanDataSource() datasource.DataSource {
	return &apigwPlanDataSource{}
}

func (d *apigwPlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_plan"
}

func (d *apigwPlanDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.ApigwClient
}

type apigwPlanDataSourceModel struct {
	ID              types.String           `tfsdk:"id"`
	CreatedAt       types.String           `tfsdk:"created_at"`
	UpdatedAt       types.String           `tfsdk:"updated_at"`
	Name            types.String           `tfsdk:"name"`
	Price           types.String           `tfsdk:"price"`
	Description     types.String           `tfsdk:"description"`
	MaxServices     types.Int32            `tfsdk:"max_services"`
	MaxRequests     types.Int64            `tfsdk:"max_requests"`
	MaxRequestsUnit types.String           `tfsdk:"max_requests_unit"`
	Overage         *apigwPlanOverageModel `tfsdk:"overage"`
}

type apigwPlanOverageModel struct {
	UnitRequests types.Int64  `tfsdk:"unit_requests"`
	UnitPrice    types.String `tfsdk:"unit_price"`
}

func (d *apigwPlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaDataSourceId("API Gateway Plan"),
			"name":       common.SchemaDataSourceName("API Gateway Plan"),
			"created_at": schemaDataSourceAPIGWCreatedAt("API Gateway Plan"),
			"updated_at": schemaDataSourceAPIGWUpdatedAt("API Gateway Plan"),
			"price": schema.StringAttribute{
				Computed:    true,
				Description: "The price of the plan",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the plan",
			},
			"max_services": schema.Int32Attribute{
				Computed:    true,
				Description: "The maximum number of services",
			},
			"max_requests": schema.Int64Attribute{
				Computed:    true,
				Description: "The maximum number of requests",
			},
			"max_requests_unit": schema.StringAttribute{
				Computed:    true,
				Description: "The unit for max requests",
			},
			"overage": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Overage information",
				Attributes: map[string]schema.Attribute{
					"unit_requests": schema.Int64Attribute{
						Computed:    true,
						Description: "Unit requests for overage",
					},
					"unit_price": schema.StringAttribute{
						Computed:    true,
						Description: "Unit price for overage",
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing API Gateway Plan.",
	}
}

func (d *apigwPlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apigwPlanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subsOp := apigw.NewSubscriptionOp(d.client)
	list, err := subsOp.ListPlans(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list API Gateway plans: %s", err))
		return
	}

	var plan *v1.Plan
	for _, p := range list {
		if strings.Contains(p.Name.Value, data.Name.ValueString()) {
			plan = &p
			break
		}
	}
	if plan == nil {
		resp.Diagnostics.AddError("Read: Search Error", fmt.Sprintf("API Gateway plan with name '%s' not found", data.Name.ValueString()))
		return
	}

	data.ID = types.StringValue(plan.ID.Value.String())
	data.CreatedAt = types.StringValue(plan.CreatedAt.Value.String())
	data.UpdatedAt = types.StringValue(plan.UpdatedAt.Value.String())
	data.Price = types.StringValue(plan.Price.Value)
	data.Description = types.StringValue(plan.Description.Value)
	data.MaxServices = types.Int32Value(int32(plan.MaxServices.Value))
	data.MaxRequests = types.Int64Value(int64(plan.MaxRequests.Value))
	data.MaxRequestsUnit = types.StringValue(string(plan.MaxRequestsUnit.Value))
	if o, ok := plan.Overage.Get(); ok {
		data.Overage = &apigwPlanOverageModel{
			UnitRequests: types.Int64Value(int64(o.UnitRequests.Value)),
			UnitPrice:    types.StringValue(o.UnitPrice.Value),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
