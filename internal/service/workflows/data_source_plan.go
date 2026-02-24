// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowPlanDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &workflowPlanDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowPlanDataSource{}
)

func NewPlanDataSource() datasource.DataSource {
	return &workflowPlanDataSource{}
}

func (d *workflowPlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows_plan"
}

func (d *workflowPlanDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.WorkflowsClient
}

type workflowPlanModel struct {
	ID                  types.String  `tfsdk:"id"`
	Name                types.String  `tfsdk:"name"`
	Grade               types.Float64 `tfsdk:"grade"`
	BasePrice           types.Float64 `tfsdk:"base_price"`
	IncludedSteps       types.Float64 `tfsdk:"included_steps"`
	OverageStepUnit     types.Float64 `tfsdk:"overage_step_unit"`
	OveragePricePerUnit types.Float64 `tfsdk:"overage_price_per_unit"`
}

func (model *workflowPlanModel) updateState(data v1.ListPlansOKPlansItem) {
	model.ID = types.StringValue(fmt.Sprintf("%.0f", data.ID))
	model.Name = types.StringValue(data.Name)
	model.Grade = types.Float64Value(data.Grade)
	model.BasePrice = types.Float64Value(data.BasePrice)
	model.IncludedSteps = types.Float64Value(data.IncludedSteps)
	model.OverageStepUnit = types.Float64Value(data.OverageStepUnit)
	model.OveragePricePerUnit = types.Float64Value(data.OveragePricePerUnit)
}

func (d *workflowPlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resourceName := "Workflows Plan"

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaDataSourceId(resourceName),
			"name": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The name of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"grade": schema.Float64Attribute{
				Computed:    true,
				Description: desc.Sprintf("The grade of the %s.", resourceName),
			},
			"base_price": schema.Float64Attribute{
				Computed:    true,
				Description: desc.Sprintf("The base price of the %s.", resourceName),
			},
			"included_steps": schema.Float64Attribute{
				Computed:    true,
				Description: desc.Sprintf("The included steps of the %s.", resourceName),
			},
			"overage_step_unit": schema.Float64Attribute{
				Computed:    true,
				Description: desc.Sprintf("The overage step unit of the %s.", resourceName),
			},
			"overage_price_per_unit": schema.Float64Attribute{
				Computed:    true,
				Description: desc.Sprintf("The overage price per unit of the %s.", resourceName),
			},
		},
		MarkdownDescription: "Get information about a Workflow Plan.",
	}
}

func (d *workflowPlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workflowPlanModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subscriptionOp := workflows.NewSubscriptionOp(d.client)
	plans, err := subscriptionOp.ListPlans(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Workflow plans: %s", err))
		return
	}

	availablePlanNames := make([]string, 0, len(plans.Plans))
	for _, p := range plans.Plans {
		availablePlanNames = append(availablePlanNames, p.Name)
	}

	var plan *v1.ListPlansOKPlansItem
	for _, p := range plans.Plans {
		if strings.Contains(p.Name, data.Name.ValueString()) {
			if plan != nil {
				resp.Diagnostics.AddError("Read: Search Error", fmt.Sprintf("multiple Workflow plans found with name containing '%s'. available plan names: %s", data.Name.ValueString(), strings.Join(availablePlanNames, ", ")))
				return
			}
			plan = &p
		}
	}

	if plan == nil {
		resp.Diagnostics.AddError("Read: Search Error", fmt.Sprintf("Workflow plan with name '%s' not found. available plan names: %s", data.Name.ValueString(), strings.Join(availablePlanNames, ", ")))
		return
	}

	data.updateState(*plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
