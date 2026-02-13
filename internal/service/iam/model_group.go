// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
)

type groupBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (model *groupBaseModel) updateState(group *v1.Group) {
	model.ID = types.StringValue(strconv.Itoa(group.ID))
	model.Name = types.StringValue(group.Name)
	model.Description = types.StringValue(group.Description)
	model.CreatedAt = types.StringValue(group.CreatedAt)
	model.UpdatedAt = types.StringValue(group.UpdatedAt)
}
