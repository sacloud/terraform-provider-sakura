// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
)

type folderBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ParentID    types.String `tfsdk:"parent_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (model *folderBaseModel) updateState(folder *v1.Folder) {
	model.ID = types.StringValue(strconv.Itoa(folder.ID))
	model.Name = types.StringValue(folder.Name)
	model.Description = types.StringValue(folder.Description)
	if folder.ParentID.IsNull() {
		model.ParentID = types.StringNull()
	} else {
		model.ParentID = types.StringValue(strconv.Itoa(folder.ParentID.Value))
	}
	model.CreatedAt = types.StringValue(folder.CreatedAt)
	model.UpdatedAt = types.StringValue(folder.UpdatedAt)
}
