// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwGroupBaseModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Tags      types.Set    `tfsdk:"tags"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (m *apigwGroupBaseModel) updateState(group *v1.Group) {
	m.ID = types.StringValue(group.ID.Value.String())
	m.Name = types.StringValue(string(group.Name.Value))
	m.Tags = common.StringsToTset(group.Tags)
	m.CreatedAt = types.StringValue(group.CreatedAt.Value.String())
	m.UpdatedAt = types.StringValue(group.UpdatedAt.Value.String())
}
