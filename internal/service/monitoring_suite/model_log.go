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

type logStorageBaseModel struct {
	msBaseModel
	AccountID      types.String `tfsdk:"account_id"`
	ResourceID     types.String `tfsdk:"resource_id"`
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

func (model *logStorageBaseModel) updateState(storage *monitoringsuiteapi.LogStorage) {
	model.updateBaseState(strconv.FormatInt(storage.GetID(), 10), storage.GetName().Value, storage.GetDescription().Value)
	model.AccountID = types.StringValue(storage.GetAccountID())
	model.ResourceID = types.StringValue(strconv.FormatInt(storage.GetResourceID().Value, 10))
	model.IsSystem = types.BoolValue(storage.GetIsSystem())
	model.ExpireDay = types.Int64Value(int64(storage.GetExpireDay()))
	model.CreatedAt = types.StringValue(storage.GetCreatedAt().String())
	model.Classification = types.StringValue(string(storage.GetClassification()))

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

type logRoutingBaseModel struct {
	ID            types.String `tfsdk:"id"`
	ResourceID    types.String `tfsdk:"resource_id"`
	StorageID     types.String `tfsdk:"storage_id"`
	PublisherCode types.String `tfsdk:"publisher_code"`
	Variant       types.String `tfsdk:"variant"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func (model *logRoutingBaseModel) updateState(routing *monitoringsuiteapi.LogRouting) {
	model.ID = types.StringValue(routing.UID.String())
	model.ResourceID = types.StringValue(strconv.Itoa(int(routing.ResourceID.Value)))
	model.StorageID = types.StringValue(strconv.Itoa(int(routing.LogStorage.ID)))
	model.PublisherCode = types.StringValue(routing.Publisher.Code)
	model.Variant = types.StringValue(routing.Variant)
	model.CreatedAt = types.StringValue(routing.CreatedAt.String())
	model.UpdatedAt = types.StringValue(routing.UpdatedAt.String())
}
