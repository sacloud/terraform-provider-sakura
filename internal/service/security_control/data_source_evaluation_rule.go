// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	seccon "github.com/sacloud/security-control-api-go"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type evaluationRuleDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &evaluationRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &evaluationRuleDataSource{}
)

func NewEvaluationRuleDataSource() datasource.DataSource {
	return &evaluationRuleDataSource{}
}

func (d *evaluationRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_control_evaluation_rule"
}

func (d *evaluationRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.SecurityControlClient
}

type evaluationRuleDataSourceModel struct {
	evaluationRuleBaseModel
}

func (d *evaluationRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Evaluation Rule",
				Validators: []validator.String{
					stringvalidator.OneOf(common.MapTo(v1.EvaluationRuleIDServerNoPublicIP.AllValues(), common.ToString)...),
				},
			},
			"description": common.SchemaDataSourceDescription("Security Control's Evaluation Rule"),
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the Evaluation Rule is enabled",
			},
			"iam_roles_required": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The set of IAM roles required for the Evaluation Rule",
			},
			"parameters": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The parameters of the Evaluation Rule",
				Attributes: map[string]schema.Attribute{
					"service_principal_id": schema.StringAttribute{
						Computed:    true,
						Description: "The Service Principal ID associated with the Evaluation Rule",
					},
					"targets": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "The list of targets for the Evaluation Rule",
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing Security Control Evaluation Rule.",
	}
}

func (d *evaluationRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data evaluationRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	erOp := seccon.NewEvaluationRulesOp(d.client)
	result, err := erOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Security Control Evaluation Rules: %s", err))
		return
	}

	data.updateState(result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
