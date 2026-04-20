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

type traceStorageBaseModel struct {
	msBaseModel
	ProjectID           types.String `tfsdk:"project_id"`
	ResourceID          types.String `tfsdk:"resource_id"`
	RetentionPeriodDays types.Int32  `tfsdk:"retention_period_days"`
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

func (model *traceStorageBaseModel) updateState(storage *monitoringsuiteapi.TraceStorage) {
	model.updateBaseState(strconv.FormatInt(storage.GetID(), 10), storage.GetName().Value, storage.GetDescription().Value)
	model.ProjectID = types.StringValue(storage.GetAccountID())
	model.ResourceID = types.StringValue(strconv.FormatInt(storage.GetResourceID(), 10))
	model.RetentionPeriodDays = types.Int32Value(int32(storage.GetRetentionPeriodDays()))
	model.CreatedAt = types.StringValue(storage.GetCreatedAt().String())

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
