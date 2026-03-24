// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type metricStorageBaseModel struct {
	common.SakuraBaseModel
	IconID     types.String `tfsdk:"icon_id"`
	AccountID  types.String `tfsdk:"account_id"`
	ResourceID types.Int64  `tfsdk:"resource_id"`
	IsSystem   types.Bool   `tfsdk:"is_system"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
	Endpoints  types.Object `tfsdk:"endpoints"`
	Usage      types.Object `tfsdk:"usage"`
}

type metricStorageEndpointsModel struct {
	Address types.String `tfsdk:"address"`
}

func (m metricStorageEndpointsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address": types.StringType,
	}
}

type metricStorageUsageModel struct {
	MetricsRoutings types.Int64 `tfsdk:"metrics_routings"`
	AlertRules      types.Int64 `tfsdk:"alert_rules"`
	LogMeasureRules types.Int64 `tfsdk:"log_measure_rules"`
}

func (m metricStorageUsageModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"metrics_routings":  types.Int64Type,
		"alert_rules":       types.Int64Type,
		"log_measure_rules": types.Int64Type,
	}
}

func updateMetricsStorageState(model *metricStorageBaseModel, storage *monitoringsuiteapi.MetricsStorage) {
	model.UpdateBaseState(strconv.FormatInt(storage.GetID(), 10), storage.GetName().Value, storage.GetDescription().Value, storage.GetTags())
	if icon, ok := storage.GetIcon().Get(); ok {
		model.IconID = stringValueOrNull(icon.GetID())
	} else {
		model.IconID = types.StringNull()
	}
	model.AccountID = types.StringValue(storage.GetAccountID())
	model.ResourceID = optInt64ToType(storage.GetResourceID())
	model.IsSystem = types.BoolValue(storage.GetIsSystem())
	model.CreatedAt = types.StringValue(storage.GetCreatedAt().String())
	model.UpdatedAt = types.StringValue(storage.GetUpdatedAt().String())
	model.Tags = common.StringsToTset(storage.GetTags())

	endpoints := storage.GetEndpoints()
	endpointsModel := &metricStorageEndpointsModel{
		Address: types.StringValue(endpoints.GetAddress()),
	}
	model.Endpoints = types.ObjectNull(metricStorageEndpointsModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), endpointsModel.AttributeTypes(), endpointsModel); !diags.HasError() {
		model.Endpoints = value
	}

	usage := storage.GetUsage()
	usageModel := &metricStorageUsageModel{
		MetricsRoutings: types.Int64Value(int64(usage.GetMetricsRoutings())),
		AlertRules:      types.Int64Value(int64(usage.GetAlertRules())),
		LogMeasureRules: types.Int64Value(int64(usage.GetLogMeasureRules())),
	}
	model.Usage = types.ObjectNull(metricStorageUsageModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), usageModel.AttributeTypes(), usageModel); !diags.HasError() {
		model.Usage = value
	}
}
