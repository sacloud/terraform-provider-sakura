// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package icon

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type iconBaseModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
	Tags types.Set    `tfsdk:"tags"`
}

func (model *iconBaseModel) updateState(icon *iaas.Icon) {
	model.ID = types.StringValue(icon.ID.String())
	model.Name = types.StringValue(icon.Name)
	model.Tags = common.StringsToTset(icon.Tags)
	model.URL = types.StringValue(icon.URL)
}
