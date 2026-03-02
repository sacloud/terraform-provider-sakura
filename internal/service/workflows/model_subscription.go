// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowsSubscriptionBaseModel struct {
	ID           types.String `tfsdk:"id"`
	AccountID    types.String `tfsdk:"account_id"`
	ContractID   types.String `tfsdk:"contract_id"`
	PlanID       types.String `tfsdk:"plan_id"`
	PlanName     types.String `tfsdk:"plan_name"`
	ActivateFrom types.String `tfsdk:"activate_from"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (model *workflowsSubscriptionBaseModel) updateState(data *v1.GetSubscriptionOK) {
	if v, ok := data.CurrentPlan.Get(); ok {
		model.ID = types.StringValue(v.ID)
		model.AccountID = types.StringValue(v.AccountId)
		model.ContractID = types.StringValue(v.ContractId)
		model.PlanName = types.StringValue(v.PlanName)
		model.PlanID = types.StringValue(fmt.Sprintf("%.0f", v.PlanId))
		model.ActivateFrom = types.StringValue(v.ActivateFrom.String())
		model.CreatedAt = types.StringValue(v.CreatedAt.String())
		model.UpdatedAt = types.StringValue(v.UpdatedAt.String())
	}
}
