// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type projectApiKeyBaseModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ProjectID        types.String `tfsdk:"project_id"`
	AccessToken      types.String `tfsdk:"access_token"`
	IAMRoles         types.List   `tfsdk:"iam_roles"`
	ServerResourceID types.String `tfsdk:"server_resource_id"`
	Zone             types.String `tfsdk:"zone"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (model *projectApiKeyBaseModel) updateState(paKey *v1.ProjectApiKey) {
	model.ID = types.StringValue(strconv.Itoa(paKey.ID))
	model.Name = types.StringValue(paKey.Name)
	model.Description = types.StringValue(paKey.Description)
	model.ProjectID = types.StringValue(strconv.Itoa(paKey.ProjectID))
	model.AccessToken = types.StringValue(paKey.AccessToken)
	model.IAMRoles = common.StringsToTlist(paKey.IamRoles)
	if paKey.ServerResourceID.IsSet() && paKey.ServerResourceID.Value != "" {
		model.ServerResourceID = types.StringValue(paKey.ServerResourceID.Value)
	} else {
		model.ServerResourceID = types.StringNull()
	}
	if paKey.ZoneID.IsSet() && paKey.ZoneID.Value != "" {
		model.Zone = types.StringValue(paKey.ZoneID.Value)
	} else {
		model.Zone = types.StringNull()
	}
	model.CreatedAt = types.StringValue(paKey.CreatedAt.Value)
	model.UpdatedAt = types.StringValue(paKey.UpdatedAt.Value)
}
