// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
)

type alertBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ProjectID   types.String `tfsdk:"project_id"` // さくらではプロジェクトという命名で統一するようになっているため、account_idではなくproject_idとする
	ResourceID  types.String `tfsdk:"resource_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (model *alertBaseModel) updateState(alert *v1.AlertProject) {
	model.ID = types.StringValue(strconv.FormatInt(alert.ID, 10))
	model.Name = types.StringValue(alert.Name.Value)
	model.Description = types.StringValue(alert.Description.Value)
	model.ProjectID = types.StringValue(alert.AccountID)
	model.ResourceID = types.StringValue(strconv.FormatInt(alert.ResourceID.Value, 10))
	model.CreatedAt = types.StringValue(alert.CreatedAt.String())
}

type alertRuleBaseModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	AlertID                   types.String `tfsdk:"alert_id"` // アラートプロジェクトとサービス全体で利用するプロジェクトでわかりにくいため、alert_idとする
	MetricStorageID           types.String `tfsdk:"metric_storage_id"`
	Query                     types.String `tfsdk:"query"`
	Format                    types.String `tfsdk:"format"`
	Template                  types.String `tfsdk:"template"`
	Open                      types.Bool   `tfsdk:"open"`
	EnabledWarning            types.Bool   `tfsdk:"enabled_warning"`
	EnabledCritical           types.Bool   `tfsdk:"enabled_critical"`
	ThresholdWarning          types.String `tfsdk:"threshold_warning"`
	ThresholdCritical         types.String `tfsdk:"threshold_critical"`
	ThresholdDurationWarning  types.Int64  `tfsdk:"threshold_duration_warning"`
	ThresholdDurationCritical types.Int64  `tfsdk:"threshold_duration_critical"`
}

func (model *alertRuleBaseModel) updateState(alertRule *v1.AlertRule) {
	model.ID = types.StringValue(alertRule.UID.String())
	model.Name = types.StringValue(alertRule.Name.Value)
	model.AlertID = types.StringValue(strconv.FormatInt(alertRule.ProjectID.Value, 10))
	model.MetricStorageID = types.StringValue(strconv.FormatInt(alertRule.MetricsStorageID.Value, 10))
	model.Query = types.StringValue(alertRule.Query)
	// FormatとTemplateはAPIのレスポンスでは空文字列になることがあるため、空文字列の場合はNullとする
	if alertRule.Format.Value != "" {
		model.Format = types.StringValue(alertRule.Format.Value)
	} else {
		model.Format = types.StringNull()
	}
	if alertRule.Template.Value != "" {
		model.Template = types.StringValue(alertRule.Template.Value)
	} else {
		model.Template = types.StringNull()
	}
	model.Open = types.BoolValue(alertRule.Open)
	model.EnabledWarning = types.BoolValue(alertRule.EnabledWarning.Value)
	model.EnabledCritical = types.BoolValue(alertRule.EnabledCritical.Value)
	model.ThresholdWarning = types.StringValue(alertRule.ThresholdWarning.Value)
	model.ThresholdCritical = types.StringValue(alertRule.ThresholdCritical.Value)
	model.ThresholdDurationWarning = types.Int64Value(alertRule.ThresholdDurationWarning.Value)
	model.ThresholdDurationCritical = types.Int64Value(alertRule.ThresholdDurationCritical.Value)
}

type alertNotificationTargetBaseModel struct {
	ID          types.String `tfsdk:"id"`
	AlertID     types.String `tfsdk:"alert_id"`
	ServiceType types.String `tfsdk:"service_type"`
	URL         types.String `tfsdk:"url"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
}

func (model *alertNotificationTargetBaseModel) updateState(target *v1.NotificationTarget) {
	model.ID = types.StringValue(target.UID.String())
	model.AlertID = types.StringValue(strconv.FormatInt(target.ProjectID.Value, 10))
	model.ServiceType = flattenAlertNotificationTargetServiceType(string(target.ServiceType))
	// URLはAPIのレスポンスでは空文字列になることがあるため、空文字列の場合はNullとする
	if target.URL.Value != "" {
		model.URL = types.StringValue(target.URL.Value)
	} else {
		model.URL = types.StringNull()
	}
	// 現在のOpenAPIでは構造が指定されていないため空のJSONとなるが、将来的に構造が追加される可能性があるため、JSON文字列として保存する
	j, _ := json.Marshal(target.Config)
	model.Config = types.StringValue(string(j))
	model.Description = types.StringValue(target.Description.Value)
}

func flattenAlertNotificationTargetServiceType(st string) types.String {
	switch st {
	case "SAKURA_SIMPLE_NOTICE":
		return types.StringValue("simple_notification")
	case "SAKURA_EVENT_BUS":
		return types.StringValue("eventbus")
	default:
		return types.StringNull()
	}
}

type alertNotificationRoutingBaseModel struct {
	ID                    types.String                              `tfsdk:"id"`
	AlertID               types.String                              `tfsdk:"alert_id"`
	NotificationTargetID  types.String                              `tfsdk:"notification_target_id"`
	MatchLabels           []alertNotificationRoutingMatchLabelModel `tfsdk:"match_labels"`
	ResendIntervalMinutes types.Int32                               `tfsdk:"resend_interval_minutes"`
	Order                 types.Int32                               `tfsdk:"order"`
}

type alertNotificationRoutingMatchLabelModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (model *alertNotificationRoutingBaseModel) updateState(routing *v1.NotificationRouting) {
	model.ID = types.StringValue(routing.UID.String())
	model.AlertID = types.StringValue(strconv.FormatInt(routing.ProjectID.Value, 10))
	model.NotificationTargetID = types.StringValue(routing.NotificationTarget.UID.String())
	model.ResendIntervalMinutes = types.Int32Value(int32(routing.ResendIntervalMinutes.Value))
	model.Order = types.Int32Value(int32(routing.Order))

	matchLabels := make([]alertNotificationRoutingMatchLabelModel, len(routing.MatchLabels))
	for i, ml := range routing.MatchLabels {
		matchLabels[i] = alertNotificationRoutingMatchLabelModel{
			Name:  types.StringValue(ml.Name),
			Value: types.StringValue(ml.Value),
		}
	}
	model.MatchLabels = matchLabels
}

type alertLogMeasureRuleBaseModel struct {
	ID              types.String                  `tfsdk:"id"`
	Name            types.String                  `tfsdk:"name"`
	Description     types.String                  `tfsdk:"description"`
	AlertID         types.String                  `tfsdk:"alert_id"`
	LogStorageID    types.String                  `tfsdk:"log_storage_id"`
	MetricStorageID types.String                  `tfsdk:"metric_storage_id"`
	Rule            *alertLogMeasureRuleRuleModel `tfsdk:"rule"`
	CreatedAt       types.String                  `tfsdk:"created_at"`
	UpdatedAt       types.String                  `tfsdk:"updated_at"`
}

type alertLogMeasureRuleRuleModel struct {
	Version types.String                   `tfsdk:"version"`
	Query   *alertLogMeasureRuleQueryModel `tfsdk:"query"`
}

type alertLogMeasureRuleQueryModel struct {
	Matchers jsontypes.Normalized `tfsdk:"matchers"`
}

func (model *alertLogMeasureRuleBaseModel) updateState(rule *v1.LogMeasureRule) {
	model.ID = types.StringValue(rule.UID.String())
	model.Name = types.StringValue(rule.Name.Value)
	model.Description = types.StringValue(rule.Description.Value)
	model.AlertID = types.StringValue(strconv.Itoa(int(rule.GetProjectID().Value)))
	model.LogStorageID = types.StringValue(strconv.Itoa(int(rule.LogStorage.ID)))
	model.MetricStorageID = types.StringValue(strconv.Itoa(int(rule.MetricsStorage.ID)))
	model.CreatedAt = types.StringValue(rule.CreatedAt.String())
	model.UpdatedAt = types.StringValue(rule.UpdatedAt.String())

	matchers, err := json.Marshal(rule.Rule.Query.Matchers)
	if err != nil {
		matchers = []byte("{}")
	}
	model.Rule = &alertLogMeasureRuleRuleModel{
		Version: types.StringValue(string(rule.Rule.Version)),
		Query: &alertLogMeasureRuleQueryModel{
			Matchers: jsontypes.NewNormalizedValue(string(matchers)),
		},
	}
}
