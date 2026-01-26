// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaas "github.com/sacloud/iaas-api-go"
	seccon "github.com/sacloud/security-control-api-go"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type evaluationRuleResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &evaluationRuleResource{}
	_ resource.ResourceWithConfigure   = &evaluationRuleResource{}
	_ resource.ResourceWithImportState = &evaluationRuleResource{}
)

func NewEvaluationRuleResource() resource.Resource {
	return &evaluationRuleResource{}
}

func (r *evaluationRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_control_evaluation_rule"
}

func (r *evaluationRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SecurityControlClient
}

type evaluationRuleResourceModel struct {
	evaluationRuleBaseModel
	NoActionOnDelete types.Bool     `tfsdk:"no_action_on_delete"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func (r *evaluationRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Evaluation Rule",
				Validators: []validator.String{
					stringvalidator.OneOf(common.MapTo(v1.EvaluationRuleIDServerNoPublicIP.AllValues(), common.ToString)...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the Evaluation Rule is enabled",
			},
			"parameters": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The parameters of the Evaluation Rule",
				Attributes: map[string]schema.Attribute{
					"service_principal_id": schema.StringAttribute{
						Optional:    true,
						Description: "The Service Principal ID associated with the Evaluation Rule",
					},
					"targets": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "The list of targets for the Evaluation Rule",
					},
				},
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the Evaluation Rule",
			},
			"iam_roles_required": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The set of IAM roles required for the Evaluation Rule",
			},
			"no_action_on_delete": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "No action when the resource is deleted.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Security Control's Evaluation Rule.",
	}
}

func (r *evaluationRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *evaluationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan evaluationRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	erOp := seccon.NewEvaluationRulesOp(r.client)
	erReq := seccon.SetupEvaluationRuleInput(expandInputParams(&plan))
	script, err := erOp.Update(ctx, plan.ID.ValueString(), erReq)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update Security Control's Evaluation Rule[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(script)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *evaluationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state evaluationRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	evaluationRule := getEvaluationRule(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if evaluationRule == nil || resp.Diagnostics.HasError() {
		return
	}

	state.updateState(evaluationRule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *evaluationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan evaluationRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	erOp := seccon.NewEvaluationRulesOp(r.client)
	erReq := seccon.SetupEvaluationRuleInput(expandInputParams(&plan))
	script, err := erOp.Update(ctx, plan.ID.ValueString(), erReq)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Security Control's Evaluation Rule[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(script)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *evaluationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state evaluationRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.NoActionOnDelete.ValueBool() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	erOp := seccon.NewEvaluationRulesOp(r.client)
	erReq := seccon.SetupEvaluationRuleInput(expandInputParams(&state))
	erReq.IsEnabled = false
	_, err := erOp.Update(ctx, state.ID.ValueString(), erReq)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to disable Security Control's Evaluation Rule[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getEvaluationRule(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.EvaluationRule {
	erOp := seccon.NewEvaluationRulesOp(client)
	evaluationRule, err := erOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Security Control's EvaluationRule[%s]: %s", id, err))
		return nil
	}

	return evaluationRule
}

func expandInputParams(model *evaluationRuleResourceModel) *seccon.EvaluationRuleInputParams {
	params := &seccon.EvaluationRuleInputParams{
		ID:      model.ID.ValueString(),
		Enabled: model.Enabled.ValueBool(),
	}

	if model.Parameters != nil {
		if utils.IsKnown(model.Parameters.ServicePrincipalID) {
			params.ServicePrincipalID = model.Parameters.ServicePrincipalID.ValueString()
		}
		if utils.IsKnown(model.Parameters.Targets) {
			params.Targets = common.TlistToStringsOrDefault(model.Parameters.Targets)
		}
	}

	return params
}
