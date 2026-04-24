// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
)

type userProvisioningBaseModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	BaseURL   types.String `tfsdk:"base_url"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (model *userProvisioningBaseModel) updateState(scim *v1.ScimConfigurationBase) {
	model.ID = types.StringValue(scim.ID.String())
	model.Name = types.StringValue(scim.Name)
	model.BaseURL = types.StringValue(scim.BaseURL.String())
	model.CreatedAt = types.StringValue(scim.CreatedAt)
	model.UpdatedAt = types.StringValue(scim.UpdatedAt)
}
