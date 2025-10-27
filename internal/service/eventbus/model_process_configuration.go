// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

const (
	destinationSimpleMQ           = "simplemq"
	destinationSimpleNotification = "simplenotification"
)

type processConfigurationBaseModel struct {
	common.SakuraBaseModel
	// TODO: iconはsdkで未対応
	// IconID types.String `tfsdk:"icon_id"`

	Destination types.String `tfsdk:"destination"`
	Parameters  types.String `tfsdk:"parameters"`
}

func (model *processConfigurationBaseModel) updateState(data *v1.ProcessConfiguration) {
	id := strconv.FormatInt(data.ID, 10)
	model.UpdateBaseState(id, data.Name, data.Description, data.Tags)

	model.Destination = types.StringValue(string(data.Settings.Destination))
	model.Parameters = types.StringValue(data.Settings.Parameters)
}
