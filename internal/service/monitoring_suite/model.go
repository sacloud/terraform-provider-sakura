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

type logStorageBaseModel struct {
	common.SakuraBaseModel
	IconID         types.String `tfsdk:"icon_id"`
	AccountID      types.String `tfsdk:"account_id"`
	ResourceID     types.Int64  `tfsdk:"resource_id"`
	IsSystem       types.Bool   `tfsdk:"is_system"`
	Classification types.String `tfsdk:"classification"`
	ExpireDay      types.Int64  `tfsdk:"expire_day"`
	CreatedAt      types.String `tfsdk:"created_at"`
	Endpoints      types.Object `tfsdk:"endpoints"`
	Usage          types.Object `tfsdk:"usage"`
}

type logStorageEndpointsModel struct {
	Ingester types.Object `tfsdk:"ingester"`
}

func (m logStorageEndpointsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ingester": types.ObjectType{AttrTypes: logStorageIngesterModel{}.AttributeTypes()},
	}
}

type logStorageIngesterModel struct {
	Address  types.String `tfsdk:"address"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (m logStorageIngesterModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address":  types.StringType,
		"insecure": types.BoolType,
	}
}

type logStorageUsageModel struct {
	LogRoutings     types.Int64 `tfsdk:"log_routings"`
	LogMeasureRules types.Int64 `tfsdk:"log_measure_rules"`
}

func (m logStorageUsageModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"log_routings":      types.Int64Type,
		"log_measure_rules": types.Int64Type,
	}
}

func updateLogStorageState(model *logStorageBaseModel, storage *monitoringsuiteapi.LogStorage) {
	model.UpdateBaseState(strconv.FormatInt(storage.GetID(), 10), storage.GetName().Value, storage.GetDescription().Value, storage.GetTags())
	if icon, ok := storage.GetIcon().Get(); ok {
		model.IconID = stringValueOrNull(icon.GetID())
	} else {
		model.IconID = types.StringNull()
	}
	model.AccountID = types.StringValue(storage.GetAccountID())
	model.ResourceID = optInt64ToType(storage.GetResourceID())
	model.IsSystem = types.BoolValue(storage.GetIsSystem())
	model.ExpireDay = types.Int64Value(int64(storage.GetExpireDay()))
	model.CreatedAt = types.StringValue(storage.GetCreatedAt().String())
	model.Tags = common.StringsToTset(storage.GetTags())

	endpoints := storage.GetEndpoints()
	ingester := endpoints.GetIngester()
	ingesterModel := &logStorageIngesterModel{
		Address:  types.StringValue(ingester.GetAddress()),
		Insecure: optBoolToType(ingester.GetInsecure()),
	}
	ingesterValue := types.ObjectNull(logStorageIngesterModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), ingesterModel.AttributeTypes(), ingesterModel); !diags.HasError() {
		ingesterValue = value
	}
	endpointsModel := &logStorageEndpointsModel{
		Ingester: ingesterValue,
	}
	model.Endpoints = types.ObjectNull(logStorageEndpointsModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), endpointsModel.AttributeTypes(), endpointsModel); !diags.HasError() {
		model.Endpoints = value
	}

	usage := storage.GetUsage()
	usageModel := &logStorageUsageModel{
		LogRoutings:     types.Int64Value(int64(usage.GetLogRoutings())),
		LogMeasureRules: types.Int64Value(int64(usage.GetLogMeasureRules())),
	}
	model.Usage = types.ObjectNull(logStorageUsageModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), usageModel.AttributeTypes(), usageModel); !diags.HasError() {
		model.Usage = value
	}
}

type metricsStorageBaseModel struct {
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

type metricsStorageEndpointsModel struct {
	Address types.String `tfsdk:"address"`
}

func (m metricsStorageEndpointsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address": types.StringType,
	}
}

type metricsStorageUsageModel struct {
	MetricsRoutings types.Int64 `tfsdk:"metrics_routings"`
	AlertRules      types.Int64 `tfsdk:"alert_rules"`
	LogMeasureRules types.Int64 `tfsdk:"log_measure_rules"`
}

func (m metricsStorageUsageModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"metrics_routings":  types.Int64Type,
		"alert_rules":       types.Int64Type,
		"log_measure_rules": types.Int64Type,
	}
}

func updateMetricsStorageState(model *metricsStorageBaseModel, storage *monitoringsuiteapi.MetricsStorage) {
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
	endpointsModel := &metricsStorageEndpointsModel{
		Address: types.StringValue(endpoints.GetAddress()),
	}
	model.Endpoints = types.ObjectNull(metricsStorageEndpointsModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), endpointsModel.AttributeTypes(), endpointsModel); !diags.HasError() {
		model.Endpoints = value
	}

	usage := storage.GetUsage()
	usageModel := &metricsStorageUsageModel{
		MetricsRoutings: types.Int64Value(int64(usage.GetMetricsRoutings())),
		AlertRules:      types.Int64Value(int64(usage.GetAlertRules())),
		LogMeasureRules: types.Int64Value(int64(usage.GetLogMeasureRules())),
	}
	model.Usage = types.ObjectNull(metricsStorageUsageModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), usageModel.AttributeTypes(), usageModel); !diags.HasError() {
		model.Usage = value
	}
}

type traceStorageBaseModel struct {
	common.SakuraBaseModel
	IconID              types.String `tfsdk:"icon_id"`
	AccountID           types.String `tfsdk:"account_id"`
	ResourceID          types.Int64  `tfsdk:"resource_id"`
	RetentionPeriodDays types.Int64  `tfsdk:"retention_period_days"`
	CreatedAt           types.String `tfsdk:"created_at"`
	Endpoints           types.Object `tfsdk:"endpoints"`
}

type traceStorageEndpointsModel struct {
	Ingester types.Object `tfsdk:"ingester"`
}

func (m traceStorageEndpointsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ingester": types.ObjectType{AttrTypes: traceStorageIngesterModel{}.AttributeTypes()},
	}
}

type traceStorageIngesterModel struct {
	Address  types.String `tfsdk:"address"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (m traceStorageIngesterModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address":  types.StringType,
		"insecure": types.BoolType,
	}
}

func updateTraceStorageState(model *traceStorageBaseModel, storage *monitoringsuiteapi.TraceStorage) {
	model.UpdateBaseState(strconv.FormatInt(storage.GetID(), 10), storage.GetName().Value, storage.GetDescription().Value, storage.GetTags())
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
	ingesterModel := &traceStorageIngesterModel{
		Address:  types.StringValue(ingester.GetAddress()),
		Insecure: optBoolToType(ingester.GetInsecure()),
	}
	ingesterValue := types.ObjectNull(traceStorageIngesterModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), ingesterModel.AttributeTypes(), ingesterModel); !diags.HasError() {
		ingesterValue = value
	}
	endpointsModel := &traceStorageEndpointsModel{
		Ingester: ingesterValue,
	}
	model.Endpoints = types.ObjectNull(traceStorageEndpointsModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), endpointsModel.AttributeTypes(), endpointsModel); !diags.HasError() {
		model.Endpoints = value
	}
}

type accessKeyBaseModel struct {
	ID          types.String `tfsdk:"id"`
	StorageID   types.String `tfsdk:"storage_id"`
	Description types.String `tfsdk:"description"`
	Token       types.String `tfsdk:"token"`
	Secret      types.String `tfsdk:"secret"`
}

func updateAccessKeyState(model *accessKeyBaseModel, storageID string, uid string, description string, token string, secret string) {
	model.ID = types.StringValue(uid)
	model.StorageID = types.StringValue(storageID)
	model.Description = types.StringValue(description)
	model.Token = types.StringValue(token)
	model.Secret = types.StringValue(secret)
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
