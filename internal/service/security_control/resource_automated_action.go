// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	iaas "github.com/sacloud/iaas-api-go"
	seccon "github.com/sacloud/security-control-api-go"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type automatedActionResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &automatedActionResource{}
	_ resource.ResourceWithConfigure   = &automatedActionResource{}
	_ resource.ResourceWithImportState = &automatedActionResource{}
)

func NewAutomatedActionResource() resource.Resource {
	return &automatedActionResource{}
}

func (r *automatedActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_control_automated_action"
}

func (r *automatedActionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SecurityControlClient
}

type automatedActionResourceModel struct {
	automatedActionBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *automatedActionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Automated Action"),
			"name":        common.SchemaResourceName("Automated Action"),
			"description": common.SchemaResourceDescription("Automated Action"),
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the Automated Action is enabled",
			},
			"action": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The settings for Automated Action",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:    true,
						Description: desc.Sprintf("The triggered type of Automated Action. This must be one of [%s]", []string{"simple_notification", "workflows"}),
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"simple_notification", "workflows"}...),
						},
					},
					"parameters": schema.SingleNestedAttribute{
						Required:    true,
						Description: "The parameters for Automated Action",
						Attributes: map[string]schema.Attribute{
							"service_principal_id": schema.StringAttribute{
								Required:    true,
								Description: "The Service Principal ID associated with the Automated Action",
							},
							"target_id": schema.StringAttribute{
								Required:    true,
								Description: "The id of target resource for the Automated Action",
							},
							"revision": schema.Int64Attribute{
								Optional:    true,
								Description: "The revision number of workflow to be executed",
							},
							"revision_alias": schema.StringAttribute{
								Optional:    true,
								Description: "The revision alias of workflow to be executed",
							},
							"args": schema.StringAttribute{
								Optional:    true,
								Description: "The json formatted arguments to be passed to the workflow",
								Validators: []validator.String{
									sacloudvalidator.StringFuncValidator(func(s string) error {
										if json.Valid([]byte(s)) {
											return nil
										} else {
											return fmt.Errorf("invalid JSON format: %s", s)
										}
									}),
								},
							},
							"name": schema.StringAttribute{
								Optional:    true,
								Description: "The name of the workflow execution",
							},
						},
					},
				},
			},
			"execution_condition": schema.StringAttribute{
				Required:    true,
				Description: "The CEL expression that defines the condition for Automated Action trigger",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp of the Automated Action",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Security Control's Automated Action.",
	}
}

func (r *automatedActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *automatedActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan automatedActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	aaOp := seccon.NewAutomatedActionsOp(r.client)
	aa, err := aaOp.Create(ctx, expandAutomatedActionInput(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Security Control's Automated Action: %s", err))
		return
	}

	plan.updateState(aa)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *automatedActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state automatedActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aa := getAutomatedAction(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if aa == nil || resp.Diagnostics.HasError() {
		return
	}

	state.updateState(aa)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *automatedActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan automatedActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	aaOp := seccon.NewAutomatedActionsOp(r.client)
	aa, err := aaOp.Update(ctx, plan.ID.ValueString(), expandAutomatedActionInput(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Security Control's Automated Action[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(aa)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *automatedActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state automatedActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	aaOp := seccon.NewAutomatedActionsOp(r.client)
	err := aaOp.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Security Control's Automated Action[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getAutomatedAction(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.AutomatedActionOutput {
	erOp := seccon.NewAutomatedActionsOp(client)
	aa, err := erOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Security Control's Automated Action[%s]: %s", id, err))
		return nil
	}

	return aa
}

func expandAutomatedActionInput(model *automatedActionResourceModel) *v1.AutomatedActionInput {
	return &v1.AutomatedActionInput{
		Name:               model.Name.ValueString(),
		Description:        v1.NewOptString(model.Description.ValueString()),
		IsEnabled:          model.Enabled.ValueBool(),
		ExecutionCondition: model.ExecutionCondition.ValueString(),
		Action:             expandActionDefinition(model.Action),
	}
}

func expandActionDefinitionType(actionType string) v1.ActionDefinitionSumType {
	switch actionType {
	case "simple_notification":
		return v1.ActionDefinitionSimpleNotificationActionDefinitionSum
	case "workflows":
		return v1.ActionDefinitionWorkflowsActionDefinitionSum
	default:
		panic("unsupported action type")
	}
}

func expandActionDefinition(model *automatedActionActionModel) v1.ActionDefinition {
	actionType := model.Type.ValueString()
	action := v1.ActionDefinitionSum{
		Type: expandActionDefinitionType(actionType),
	}

	switch actionType {
	case "simple_notification":
		action.ActionDefinitionSimpleNotification = v1.ActionDefinitionSimpleNotification{
			ActionType: v1.ActionDefinitionSimpleNotificationActionTypeSimpleNotification,
			ActionParameter: v1.SakuraSimpleNotification{
				ServicePrincipalId:  model.Parameters.ServicePrincipalID.ValueString(),
				NotificationGroupId: model.Parameters.TargetID.ValueString(),
			},
		}
	case "workflows":
		params := v1.SakuraWorkflows{
			ServicePrincipalId: model.Parameters.ServicePrincipalID.ValueString(),
			WorkflowId:         model.Parameters.TargetID.ValueString(),
		}
		if !model.Parameters.Revision.IsNull() {
			params.RevisionId = v1.NewOptInt(int(model.Parameters.Revision.ValueInt64()))
		}
		if !model.Parameters.RevisionAlias.IsNull() {
			params.RevisionAlias = v1.NewOptString(model.Parameters.RevisionAlias.ValueString())
		}
		if !model.Parameters.Args.IsNull() {
			params.Args = v1.NewOptString(model.Parameters.Args.ValueString())
		}
		if !model.Parameters.Name.IsNull() {
			params.Name = v1.NewOptString(model.Parameters.Name.ValueString())
		}
		action.ActionDefinitionWorkflows = v1.ActionDefinitionWorkflows{
			ActionType:      v1.ActionDefinitionWorkflowsActionTypeWorkflows,
			ActionParameter: params,
		}
	default:
		panic("unsupported action type")
	}

	return v1.ActionDefinition{
		OneOf: action,
	}
}
