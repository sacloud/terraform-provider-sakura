// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type evaluationRulesBaseModel struct {
	Rules []evaluationRuleBaseModel `tfsdk:"rules"`
}

type evaluationRuleBaseModel struct {
	ID               types.String                   `tfsdk:"id"`
	Description      types.String                   `tfsdk:"description"`
	Enabled          types.Bool                     `tfsdk:"enabled"`
	IamRolesRequired types.Set                      `tfsdk:"iam_roles_required"`
	Parameters       *evaluationRuleParametersModel `tfsdk:"parameters"`
}

type evaluationRuleParametersModel struct {
	ServicePrincipalID types.String `tfsdk:"service_principal_id"`
	Targets            types.List   `tfsdk:"targets"`
}

func (model *evaluationRulesBaseModel) updateState(erRules []v1.EvaluationRule) {
	rules := make([]evaluationRuleBaseModel, 0, len(erRules))
	for _, erRule := range erRules {
		m := evaluationRuleBaseModel{}
		flattenEvaluationRule(&m, &erRule)
		rules = append(rules, m)
	}
	model.Rules = rules
}

func (model *evaluationRuleBaseModel) updateState(erRule *v1.EvaluationRule) {
	flattenEvaluationRule(model, erRule)
}

func flattenEvaluationRule(model *evaluationRuleBaseModel, rule *v1.EvaluationRule) {
	switch string(rule.Rule.OneOf.Type) {
	case "server-no-public-ip":
		r := rule.Rule.OneOf.ServerNoPublicIP
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            flattenParameterTargets(r.Parameter),
		}
	case "disk-encryption-enabled":
		r := rule.Rule.OneOf.DiskEncryptionEnabled
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            flattenParameterTargets(r.Parameter),
		}
	case "dba-encryption-enabled":
		r := rule.Rule.OneOf.DbaEncryptionEnabled
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            flattenParameterTargets(r.Parameter),
		}
	case "dba-no-public-ip":
		r := rule.Rule.OneOf.DbaNoPublicIP
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            flattenParameterTargets(r.Parameter),
		}
	case "objectstorage-bucket-acl-changed":
		r := rule.Rule.OneOf.ObjectStorageBucketACLChanged
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            types.ListNull(types.StringType),
		}
	case "addon-datalake-no-public-access":
		r := rule.Rule.OneOf.AddonDatalakeNoPublicAccess
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            types.ListNull(types.StringType),
		}
	case "addon-dwh-no-public-access":
		r := rule.Rule.OneOf.AddonDwhNoPublicAccess
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            types.ListNull(types.StringType),
		}
	case "addon-threat-detection-enabled":
		r := rule.Rule.OneOf.AddonThreatDetectionEnabled
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            flattenParameterTargets(r.Parameter),
		}
	case "addon-threat-detections":
		r := rule.Rule.OneOf.AddonThreatDetections
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = nil
	case "addon-vulnerability-detections":
		r := rule.Rule.OneOf.AddonVulnerabilityDetections
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = nil
	case "elb-logging-enabled":
		r := rule.Rule.OneOf.ELBLoggingEnabled
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            types.ListNull(types.StringType),
		}
	case "iam-member-operation-detected":
		r := rule.Rule.OneOf.IAMMemberOperationDetected
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            types.ListNull(types.StringType),
		}
	case "nosql-encryption-enabled":
		r := rule.Rule.OneOf.NoSQLEncryptionEnabled
		model.ID = types.StringValue(string(r.EvaluationRuleId))
		model.Description = types.StringValue(rule.Description.Value)
		model.Enabled = types.BoolValue(rule.IsEnabled)
		model.IamRolesRequired = common.StringsToTset(rule.IamRolesRequired)
		model.Parameters = &evaluationRuleParametersModel{
			ServicePrincipalID: types.StringValue(r.Parameter.Value.ServicePrincipalId.Value),
			Targets:            flattenParameterTargets(r.Parameter),
		}
	default:
		return
	}
}

func flattenParameterTargets(targets v1.OptEvaluationRuleParametersZonedEvaluationTarget) types.List {
	if targets.IsSet() {
		// StringsToTlist内で使われるListValueFromが[]に対してnullを返すため、明示的に空リストの場合の処理を追加
		if len(targets.Value.Zones) == 0 {
			v, _ := types.ListValue(types.StringType, []attr.Value{})
			return v
		} else {
			return common.StringsToTlist(targets.Value.Zones)
		}
	} else {
		return types.ListNull(types.StringType)
	}
}
