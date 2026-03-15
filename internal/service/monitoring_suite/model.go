// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type logStorageBaseModel struct {
	common.SakuraBaseModel
	IconID         types.String              `tfsdk:"icon_id"`
	AccountID      types.String              `tfsdk:"account_id"`
	ResourceID     types.Int64               `tfsdk:"resource_id"`
	IsSystem       types.Bool                `tfsdk:"is_system"`
	Classification types.String              `tfsdk:"classification"`
	ExpireDay      types.Int64               `tfsdk:"expire_day"`
	CreatedAt      types.String              `tfsdk:"created_at"`
	Endpoints      *logStorageEndpointsModel `tfsdk:"endpoints"`
	Usage          *logStorageUsageModel     `tfsdk:"usage"`
}

type logStorageEndpointsModel struct {
	Ingester *logStorageIngesterModel `tfsdk:"ingester"`
}

type logStorageIngesterModel struct {
	Address  types.String `tfsdk:"address"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

type logStorageUsageModel struct {
	LogRoutings     types.Int64 `tfsdk:"log_routings"`
	LogMeasureRules types.Int64 `tfsdk:"log_measure_rules"`
}

type metricsStorageBaseModel struct {
	common.SakuraBaseModel
	IconID     types.String                  `tfsdk:"icon_id"`
	AccountID  types.String                  `tfsdk:"account_id"`
	ResourceID types.Int64                   `tfsdk:"resource_id"`
	IsSystem   types.Bool                    `tfsdk:"is_system"`
	CreatedAt  types.String                  `tfsdk:"created_at"`
	UpdatedAt  types.String                  `tfsdk:"updated_at"`
	Endpoints  *metricsStorageEndpointsModel `tfsdk:"endpoints"`
	Usage      *metricsStorageUsageModel     `tfsdk:"usage"`
}

type metricsStorageEndpointsModel struct {
	Address types.String `tfsdk:"address"`
}

type metricsStorageUsageModel struct {
	MetricsRoutings types.Int64 `tfsdk:"metrics_routings"`
	AlertRules      types.Int64 `tfsdk:"alert_rules"`
	LogMeasureRules types.Int64 `tfsdk:"log_measure_rules"`
}

type traceStorageBaseModel struct {
	common.SakuraBaseModel
	IconID              types.String                `tfsdk:"icon_id"`
	AccountID           types.String                `tfsdk:"account_id"`
	ResourceID          types.Int64                 `tfsdk:"resource_id"`
	RetentionPeriodDays types.Int64                 `tfsdk:"retention_period_days"`
	CreatedAt           types.String                `tfsdk:"created_at"`
	Endpoints           *traceStorageEndpointsModel `tfsdk:"endpoints"`
}

type traceStorageEndpointsModel struct {
	Ingester *traceStorageIngesterModel `tfsdk:"ingester"`
}

type traceStorageIngesterModel struct {
	Address  types.String `tfsdk:"address"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

type accessKeyBaseModel struct {
	ID          types.String `tfsdk:"id"`
	StorageID   types.String `tfsdk:"storage_id"`
	Description types.String `tfsdk:"description"`
	Token       types.String `tfsdk:"token"`
	Secret      types.String `tfsdk:"secret"`
}

type optInt64 interface {
	Get() (int64, bool)
}

func optInt64ToType(value optInt64) types.Int64 {
	if v, ok := value.Get(); ok {
		return types.Int64Value(v)
	}
	return types.Int64Null()
}

type optString interface {
	Get() (string, bool)
	Or(string) string
}

type optBool interface {
	Get() (bool, bool)
}

func optBoolToType(value optBool) types.Bool {
	if v, ok := value.Get(); ok {
		return types.BoolValue(v)
	}
	return types.BoolNull()
}

func stringValueOrNull(value optString) types.String {
	if v, ok := value.Get(); ok {
		return types.StringValue(v)
	}
	return types.StringNull()
}

func updateLogStorageState(model *logStorageBaseModel, storage *monitoringsuiteapi.LogStorage) {
	model.UpdateBaseState(int64ToString(storage.GetID()), storage.GetName().Or(""), storage.GetDescription().Or(""), storage.GetTags())
	if icon, ok := storage.GetIcon().Get(); ok {
		model.IconID = stringValueOrNull(icon.GetID())
	} else {
		model.IconID = types.StringNull()
	}
	model.AccountID = types.StringValue(storage.GetAccountID())
	model.ResourceID = optInt64ToType(storage.GetResourceID())
	model.IsSystem = types.BoolValue(storage.GetIsSystem())
	model.ExpireDay = optInt64ToType(storage.GetExpireDay())
	model.CreatedAt = types.StringValue(storage.GetCreatedAt().String())
	model.Tags = common.StringsToTset(storage.GetTags())

	endpoints := storage.GetEndpoints()
	ingester := endpoints.GetIngester()
	model.Endpoints = &logStorageEndpointsModel{
		Ingester: &logStorageIngesterModel{
			Address:  types.StringValue(ingester.GetAddress()),
			Insecure: optBoolToType(ingester.GetInsecure()),
		},
	}

	usage := storage.GetUsage()
	model.Usage = &logStorageUsageModel{
		LogRoutings:     types.Int64Value(int64(usage.GetLogRoutings())),
		LogMeasureRules: types.Int64Value(int64(usage.GetLogMeasureRules())),
	}
}

func updateMetricsStorageState(model *metricsStorageBaseModel, storage *monitoringsuiteapi.MetricsStorage) {
	model.UpdateBaseState(int64ToString(storage.GetID()), storage.GetName().Or(""), storage.GetDescription().Or(""), storage.GetTags())
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
	model.Endpoints = &metricsStorageEndpointsModel{
		Address: types.StringValue(endpoints.GetAddress()),
	}

	usage := storage.GetUsage()
	model.Usage = &metricsStorageUsageModel{
		MetricsRoutings: types.Int64Value(int64(usage.GetMetricsRoutings())),
		AlertRules:      types.Int64Value(int64(usage.GetAlertRules())),
		LogMeasureRules: types.Int64Value(int64(usage.GetLogMeasureRules())),
	}
}

func updateTraceStorageState(model *traceStorageBaseModel, storage *monitoringsuiteapi.TraceStorage) {
	model.UpdateBaseState(int64ToString(storage.GetID()), storage.GetName().Or(""), storage.GetDescription().Or(""), storage.GetTags())
	if icon, ok := storage.GetIcon().Get(); ok {
		model.IconID = stringValueOrNull(icon.GetID())
	} else {
		model.IconID = types.StringNull()
	}
	model.AccountID = types.StringValue(storage.GetAccountID())
	model.ResourceID = types.Int64Value(storage.GetResourceID())
	model.RetentionPeriodDays = types.Int64Value(int64(storage.GetRetentionPeriodDays()))
	model.CreatedAt = types.StringValue(storage.GetCreatedAt().String())
	model.Tags = common.StringsToTset(storage.GetTags())

	endpoints := storage.GetEndpoints()
	ingester := endpoints.GetIngester()
	model.Endpoints = &traceStorageEndpointsModel{
		Ingester: &traceStorageIngesterModel{
			Address:  types.StringValue(ingester.GetAddress()),
			Insecure: optBoolToType(ingester.GetInsecure()),
		},
	}
}

func updateAccessKeyState(model *accessKeyBaseModel, storageID string, uid string, description optString, token string, secret string) {
	model.ID = types.StringValue(uid)
	model.StorageID = types.StringValue(storageID)
	model.Description = stringValueOrNull(description)
	model.Token = types.StringValue(token)
	model.Secret = types.StringValue(secret)
}

func int64ToString(id int64) string {
	return fmt.Sprintf("%d", id)
}
