// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/simple-notification-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type destinationBaseModel struct {
	common.SakuraBaseModel
	IconID types.String `tfsdk:"icon_id"`
	Type   types.String `tfsdk:"type"`
	Value  types.String `tfsdk:"value"`
}

func (model *destinationBaseModel) updateState(data *v1.CommonServiceItem) error {
	model.UpdateBaseState(data.ID, data.Name, data.Description, data.Tags)

	icon, ok := data.Icon.Get()
	if !ok {
		return errors.New("invalid icon for Destination")
	}

	if icon.GetID() == "" {
		// icon_id is empty string when not set, so set null in that case
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(icon.GetID())
	}

	ds, ok := data.Settings.GetCommonServiceItemDestinationSettings()
	if !ok {
		return errors.New("invalid settings for Destination")
	}
	model.Type = types.StringValue(string(ds.Type))
	model.Value = types.StringValue(ds.Value)

	return nil
}
