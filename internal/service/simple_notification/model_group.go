// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/simple-notification-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type groupBaseModel struct {
	common.SakuraBaseModel
	Destinations types.List `tfsdk:"destinations"`
}

func (model *groupBaseModel) updateState(data *v1.CommonServiceItem) error {
	model.UpdateBaseState(data.ID, data.Name, data.Description, data.Tags)

	gr, ok := data.Settings.GetCommonServiceItemGroupSettings()
	if !ok {
		return errors.New("invalid settings for Group")
	}
	model.Destinations = common.StringsToTlist(gr.Destinations)
	return nil
}
