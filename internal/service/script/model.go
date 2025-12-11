// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package script

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

// APIのパスはnoteになっているが、マニュアルやコントロールパネルではスクリプトと呼ばれているため、scriptを使用する
type scriptBaseModel struct {
	common.SakuraBaseModel
	IconID  types.String `tfsdk:"icon_id"`
	Class   types.String `tfsdk:"class"`
	Content types.String `tfsdk:"content"`
}

func (model *scriptBaseModel) updateState(script *iaas.Note) {
	model.UpdateBaseState(script.ID.String(), script.Name, script.Description, script.Tags)
	model.Class = types.StringValue(script.Class)
	model.Content = types.StringValue(script.Content)
	if script.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(script.IconID.String())
	}
}
