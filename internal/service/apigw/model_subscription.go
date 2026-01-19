// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
)

type apigwSubscriptionBaseModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	PlanID         types.String `tfsdk:"plan_id"`
	ResourceId     types.Int64  `tfsdk:"resource_id"`
	MonthlyRequest types.Int64  `tfsdk:"monthly_request"`
	//Service        *apigwSubscriptionServiceModel `tfsdk:"service"`
	Service types.Object `tfsdk:"service"`
}

type apigwSubscriptionServiceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (m apigwSubscriptionServiceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	}
}

/* 現状Readが壊れているためListを使うためコメントアウト。修正され次第戻す
func (m *apigwSubscriptionBaseModel) updateState(sub *v1.SubscriptionDetailResponse) {
	m.ID = types.StringValue(sub.ID.Value.String())
	m.Name = types.StringValue(string(sub.Name.Value))
	m.CreatedAt = types.StringValue(sub.CreatedAt.Value.String())
	m.UpdatedAt = types.StringValue(sub.UpdatedAt.Value.String())
	m.ResourceId = types.Int64Value(sub.ResourceId.Value)
	m.MonthlyRequest = types.Int64Value(int64(sub.MonthlyRequest.Value))
	if sub.Service.IsSet() {
		m.Service = &apigwSubscriptionServiceModel{
			ID:   types.StringValue(sub.Service.Value.ID.String()),
			Name: types.StringValue(string(sub.Service.Value.Name)),
		}
	}
	if sub.Plan.IsSet() {
		m.PlanID = types.StringValue(sub.Plan.Value.PlanID.Value.String())
	}
}
*/

func (m *apigwSubscriptionBaseModel) updateState(sub *v1.Subscription) {
	m.ID = types.StringValue(sub.ID.Value.String())
	m.Name = types.StringValue(string(sub.Name.Value))
	m.CreatedAt = types.StringValue(sub.CreatedAt.Value.String())
	m.UpdatedAt = types.StringValue(sub.UpdatedAt.Value.String())
	m.ResourceId = types.Int64Value(sub.ResourceId.Value)
	m.MonthlyRequest = types.Int64Value(int64(sub.MonthlyRequest.Value))
	m.PlanID = types.StringValue(sub.PlanId.Value.String())
	if sub.Service.IsSet() {
		svc := &apigwSubscriptionServiceModel{
			ID:   types.StringValue(sub.Service.Value.ID.String()),
			Name: types.StringValue(string(sub.Service.Value.Name)),
		}
		value, diags := types.ObjectValueFrom(context.Background(), svc.AttributeTypes(), svc)
		if diags.HasError() {
			m.Service = types.ObjectNull(svc.AttributeTypes())
		}
		m.Service = value
	} else {
		m.Service = types.ObjectNull(apigwSubscriptionServiceModel{}.AttributeTypes())
	}
}
