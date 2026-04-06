// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
)

// ダッシュボードは他のモニタリングスイートのリソースと違いIDがUUIDではなくリソースIDだが、他と統一するためにresource_idフィールドも用意。
type dashboardBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ProjectID   types.String `tfsdk:"project_id"` // さくらではプロジェクトという命名で統一するようになっているため、account_idではなくproject_idとする
	ResourceID  types.String `tfsdk:"resource_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (model *dashboardBaseModel) updateState(dashboard *monitoringsuiteapi.DashboardProject) {
	model.ID = types.StringValue(strconv.FormatInt(dashboard.ID, 10))
	model.Name = types.StringValue(dashboard.Name.Value)
	model.Description = types.StringValue(dashboard.Description.Value)
	model.ProjectID = types.StringValue(dashboard.AccountID)
	model.ResourceID = types.StringValue(strconv.FormatInt(dashboard.ResourceID.Value, 10))
	model.CreatedAt = types.StringValue(dashboard.CreatedAt.String())
}
