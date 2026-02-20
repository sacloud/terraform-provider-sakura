// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
)

type servicePrincipalBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ProjectID   types.String `tfsdk:"project_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (model *servicePrincipalBaseModel) updateState(sp *v1.ServicePrincipal) {
	model.ID = types.StringValue(strconv.Itoa(sp.ID))
	model.Name = types.StringValue(sp.Name)
	model.ProjectID = types.StringValue(strconv.Itoa(sp.ProjectID))
	model.Description = types.StringValue(sp.Description)
	model.CreatedAt = types.StringValue(sp.CreatedAt.Value)
	model.UpdatedAt = types.StringValue(sp.UpdatedAt.Value)
}
