// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
)

type metricStorageBaseModel struct {
	msBaseModel
	AccountID  types.String `tfsdk:"account_id"`
	ResourceID types.String `tfsdk:"resource_id"`
	IsSystem   types.Bool   `tfsdk:"is_system"`
	CreatedAt  types.String `tfsdk:"created_at"`
	Endpoints  types.Object `tfsdk:"endpoints"`
	Usage      types.Object `tfsdk:"usage"`
	// metrics APIではupdated_atは返却されるが、他のストレージ系リソースでは返却されない。統一するため一旦モデルからは削除する。
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
	MetricRoutings  types.Int64 `tfsdk:"metric_routings"`
	AlertRules      types.Int64 `tfsdk:"alert_rules"`
	LogMeasureRules types.Int64 `tfsdk:"log_measure_rules"`
}

func (m metricStorageUsageModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"metric_routings":   types.Int64Type,
		"alert_rules":       types.Int64Type,
		"log_measure_rules": types.Int64Type,
	}
}

func (model *metricStorageBaseModel) updateState(storage *monitoringsuiteapi.MetricsStorage) {
	model.updateBaseState(strconv.FormatInt(storage.GetID(), 10), storage.GetName().Value, storage.GetDescription().Value)
	model.AccountID = types.StringValue(storage.GetAccountID())
	model.ResourceID = types.StringValue(strconv.FormatInt(storage.GetResourceID().Value, 10))
	model.IsSystem = types.BoolValue(storage.GetIsSystem())
	model.CreatedAt = types.StringValue(storage.GetCreatedAt().String())

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
		MetricRoutings:  types.Int64Value(int64(usage.GetMetricsRoutings())),
		AlertRules:      types.Int64Value(int64(usage.GetAlertRules())),
		LogMeasureRules: types.Int64Value(int64(usage.GetLogMeasureRules())),
	}
	model.Usage = types.ObjectNull(metricStorageUsageModel{}.AttributeTypes())
	if value, diags := types.ObjectValueFrom(context.Background(), usageModel.AttributeTypes(), usageModel); !diags.HasError() {
		model.Usage = value
	}
}

type metricRoutingBaseModel struct {
	ID            types.String `tfsdk:"id"`
	ResourceID    types.String `tfsdk:"resource_id"`
	StorageID     types.String `tfsdk:"storage_id"`
	PublisherCode types.String `tfsdk:"publisher_code"`
	Variant       types.String `tfsdk:"variant"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func (model *metricRoutingBaseModel) updateState(routing *monitoringsuiteapi.MetricsRouting) {
	model.ID = types.StringValue(routing.UID.String())
	model.ResourceID = types.StringValue(strconv.Itoa(int(routing.ResourceID.Value)))
	model.StorageID = types.StringValue(strconv.Itoa(int(routing.MetricsStorage.ID)))
	model.PublisherCode = types.StringValue(routing.Publisher.Code)
	model.Variant = types.StringValue(routing.Variant)
	model.CreatedAt = types.StringValue(routing.CreatedAt.String())
	model.UpdatedAt = types.StringValue(routing.UpdatedAt.String())
}
