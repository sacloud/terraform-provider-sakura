// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
)

type automatedActionBaseModel struct {
	ID                 types.String                `tfsdk:"id"`
	Name               types.String                `tfsdk:"name"`
	Description        types.String                `tfsdk:"description"`
	Enabled            types.Bool                  `tfsdk:"enabled"`
	Action             *automatedActionActionModel `tfsdk:"action"`
	ExecutionCondition types.String                `tfsdk:"execution_condition"`
	CreatedAt          types.String                `tfsdk:"created_at"`
}

type automatedActionActionModel struct {
	Type       types.String                    `tfsdk:"type"`
	Parameters *automatedActionParametersModel `tfsdk:"parameters"`
}

type automatedActionParametersModel struct {
	ServicePrincipalID types.String `tfsdk:"service_principal_id"`
	TargetID           types.String `tfsdk:"target_id"`
	Revision           types.Int64  `tfsdk:"revision"`
	RevisionAlias      types.String `tfsdk:"revision_alias"`
	Args               types.String `tfsdk:"args"`
	Name               types.String `tfsdk:"name"`
}

func (model *automatedActionBaseModel) updateState(aa *v1.AutomatedActionOutput) {
	model.ID = types.StringValue(string(aa.AutomatedActionId))
	model.Name = types.StringValue(aa.Name)
	model.Description = types.StringValue(aa.Description.Value)
	model.Enabled = types.BoolValue(aa.IsEnabled)
	model.ExecutionCondition = types.StringValue(aa.ExecutionCondition)
	model.CreatedAt = types.StringValue(aa.CreatedAt.String())
	model.Action = flattenAutomatedActionActionModel(&aa.Action.OneOf)
}

func flattenAutomatedActionActionModel(action *v1.ActionDefinitionSum) *automatedActionActionModel {
	actionModel := &automatedActionActionModel{
		Type: flattenActionDefinitionType(action.Type),
	}
	switch action.Type {
	case v1.ActionDefinitionSimpleNotificationActionDefinitionSum:
		sn := &action.ActionDefinitionSimpleNotification
		actionModel.Parameters = &automatedActionParametersModel{
			ServicePrincipalID: types.StringValue(sn.ActionParameter.ServicePrincipalId),
			TargetID:           types.StringValue(sn.ActionParameter.NotificationGroupId),
		}
	case v1.ActionDefinitionWorkflowsActionDefinitionSum:
		wf := &action.ActionDefinitionWorkflows
		parametersModel := &automatedActionParametersModel{
			ServicePrincipalID: types.StringValue(wf.ActionParameter.ServicePrincipalId),
			TargetID:           types.StringValue(wf.ActionParameter.WorkflowId),
		}
		if wf.ActionParameter.RevisionId.IsSet() {
			parametersModel.Revision = types.Int64Value(int64(wf.ActionParameter.RevisionId.Value))
		}
		if wf.ActionParameter.RevisionAlias.IsSet() {
			parametersModel.RevisionAlias = types.StringValue(wf.ActionParameter.RevisionAlias.Value)
		}
		if wf.ActionParameter.Args.IsSet() {
			parametersModel.Args = types.StringValue(wf.ActionParameter.Args.Value)
		}
		if wf.ActionParameter.Name.IsSet() {
			parametersModel.Name = types.StringValue(wf.ActionParameter.Name.Value)
		}

		actionModel.Parameters = parametersModel
	default:
		return nil
	}

	return actionModel
}

func flattenActionDefinitionType(actionType v1.ActionDefinitionSumType) types.String {
	switch actionType {
	case v1.ActionDefinitionSimpleNotificationActionDefinitionSum:
		return types.StringValue("simple_notification")
	case v1.ActionDefinitionWorkflowsActionDefinitionSum:
		return types.StringValue("workflows")
	default:
		panic("unsupported action type")
	}
}
