// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

const (
	destinationSimpleMQ           = string(v1.ProcessConfigurationSettingsDestinationSimplemq)
	destinationSimpleNotification = string(v1.ProcessConfigurationSettingsDestinationSimplenotification)
	destinationAutoScale          = string(v1.ProcessConfigurationSettingsDestinationAutoscale)
)

type processConfigurationBaseModel struct {
	common.SakuraBaseModel
	IconID types.String `tfsdk:"icon_id"`

	Destination types.String `tfsdk:"destination"`
	Parameters  types.String `tfsdk:"parameters"`
}

func (model *processConfigurationBaseModel) updateState(data *v1.CommonServiceItem) error {
	model.UpdateBaseState(data.ID, data.Name, data.Description.Value, data.Tags)
	if iconID, ok := data.Icon.Value.ID.Get(); ok {
		model.IconID = types.StringValue(iconID)
	} else {
		model.IconID = types.StringNull()
	}

	pc, ok := data.Settings.GetProcessConfigurationSettings()
	if !ok {
		return errors.New("invalid settings for ProcessConfiguration")
	}
	model.Destination = types.StringValue(string(pc.Destination))
	model.Parameters = types.StringValue(pc.Parameters)

	return nil
}
