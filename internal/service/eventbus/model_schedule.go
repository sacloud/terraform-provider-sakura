// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type scheduleBaseModel struct {
	common.SakuraBaseModel
	IconID types.String `tfsdk:"icon_id"`

	ProcessConfigurationID types.String `tfsdk:"process_configuration_id"`
	RecurringStep          types.Int64  `tfsdk:"recurring_step"`
	RecurringUnit          types.String `tfsdk:"recurring_unit"`
	Crontab                types.String `tfsdk:"crontab"`
	StartsAt               types.Int64  `tfsdk:"starts_at"`
}

func (model *scheduleBaseModel) updateState(data *v1.CommonServiceItem) error {
	model.UpdateBaseState(data.ID, data.Name, data.Description.Value, data.Tags)
	if iconID, ok := data.Icon.Value.ID.Get(); ok {
		model.IconID = types.StringValue(iconID)
	} else {
		model.IconID = types.StringNull()
	}

	schedule, ok := data.Settings.GetScheduleSettings()
	if !ok {
		return errors.New("invalid settings for Schedule")
	}
	model.ProcessConfigurationID = types.StringValue(schedule.ProcessConfigurationID)
	if v := schedule.StartsAt; v.IsString() {
		vInt64, err := strconv.ParseInt(v.String, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid StartsAt value as int64: %w", err)
		}
		model.StartsAt = types.Int64Value(vInt64)
	} else if v.IsInt64() {
		model.StartsAt = types.Int64Value(v.Int64)
	}

	if crontab, ok := schedule.Crontab.Get(); ok {
		model.Crontab = types.StringValue(crontab)
	} else {
		model.Crontab = types.StringNull()
	}

	if schedule.RecurringStep.IsSet() && schedule.RecurringUnit.IsSet() {
		model.RecurringStep = types.Int64Value(int64(schedule.RecurringStep.Value))
		model.RecurringUnit = types.StringValue(string(schedule.RecurringUnit.Value))
	} else {
		model.RecurringStep = types.Int64Null()
		model.RecurringUnit = types.StringNull()
	}
	return nil
}
