// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
)

type projectBaseModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Code           types.String `tfsdk:"code"`
	Description    types.String `tfsdk:"description"`
	ParentFolderID types.String `tfsdk:"parent_folder_id"`
	Status         types.String `tfsdk:"status"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func (model *projectBaseModel) updateState(project *v1.Project) {
	model.ID = types.StringValue(strconv.Itoa(project.ID))
	model.Name = types.StringValue(project.Name)
	model.Code = types.StringValue(project.Code)
	model.Description = types.StringValue(project.Description)
	if project.ParentFolderID.IsNull() {
		model.ParentFolderID = types.StringNull()
	} else {
		model.ParentFolderID = types.StringValue(strconv.Itoa(project.ParentFolderID.Value))
	}
	model.Status = types.StringValue(string(project.Status))
	model.CreatedAt = types.StringValue(project.CreatedAt)
	model.UpdatedAt = types.StringValue(project.UpdatedAt)
}
