// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package auto_scale

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type autoScaleBaseModel struct {
	common.SakuraBaseModel
	Zones                  types.Set                             `tfsdk:"zones"`
	Config                 types.String                          `tfsdk:"config"`
	APIKeyID               types.String                          `tfsdk:"api_key_id"`
	Enabled                types.Bool                            `tfsdk:"enabled"`
	TriggerType            types.String                          `tfsdk:"trigger_type"`
	RouterThresholdScaling *autoScaleRouterThresholdScalingModel `tfsdk:"router_threshold_scaling"`
	CPUThresholdScaling    *autoScaleCPUThresholdScalingModel    `tfsdk:"cpu_threshold_scaling"`
	ScheduleScaling        []autoScaleScheduleScalingModel       `tfsdk:"schedule_scaling"`
	IconID                 types.String                          `tfsdk:"icon_id"`
}

type autoScaleRouterThresholdScalingModel struct {
	RouterPrefix types.String `tfsdk:"router_prefix"`
	Direction    types.String `tfsdk:"direction"`
	Mbps         types.Int32  `tfsdk:"mbps"`
}

type autoScaleCPUThresholdScalingModel struct {
	ServerPrefix types.String `tfsdk:"server_prefix"`
	Up           types.Int32  `tfsdk:"up"`
	Down         types.Int32  `tfsdk:"down"`
}

type autoScaleScheduleScalingModel struct {
	Hour       types.Int32  `tfsdk:"hour"`
	Minute     types.Int32  `tfsdk:"minute"`
	Action     types.String `tfsdk:"action"`
	DaysOfWeek types.Set    `tfsdk:"days_of_week"`
}

func (m *autoScaleBaseModel) updateState(data *iaas.AutoScale) {
	m.UpdateBaseState(data.ID.String(), data.Name, data.Description, data.Tags)
	m.Zones = common.StringsToTset(data.Zones)
	m.Config = types.StringValue(data.Config)
	m.APIKeyID = types.StringValue(data.APIKeyID)
	m.TriggerType = types.StringValue(data.TriggerType.String())
	m.CPUThresholdScaling = flattenAutoScaleCPUThresholdScaling(data.CPUThresholdScaling)
	m.RouterThresholdScaling = flattenAutoScaleRouterThresholdScaling(data.RouterThresholdScaling)
	m.ScheduleScaling = flattenAutoScaleScalingSchedules(data.ScheduleScaling)
	m.Enabled = types.BoolValue(!data.Disabled)
	if data.IconID.IsEmpty() {
		m.IconID = types.StringNull()
	} else {
		m.IconID = types.StringValue(data.IconID.String())
	}
}

func flattenAutoScaleCPUThresholdScaling(scaling *iaas.AutoScaleCPUThresholdScaling) *autoScaleCPUThresholdScalingModel {
	if scaling != nil {
		return &autoScaleCPUThresholdScalingModel{
			ServerPrefix: types.StringValue(scaling.ServerPrefix),
			Up:           types.Int32Value(int32(scaling.Up)),
			Down:         types.Int32Value(int32(scaling.Down)),
		}
	}
	return nil
}

func flattenAutoScaleRouterThresholdScaling(scaling *iaas.AutoScaleRouterThresholdScaling) *autoScaleRouterThresholdScalingModel {
	if scaling != nil {
		return &autoScaleRouterThresholdScalingModel{
			RouterPrefix: types.StringValue(scaling.RouterPrefix),
			Direction:    types.StringValue(scaling.Direction),
			Mbps:         types.Int32Value(int32(scaling.Mbps)),
		}
	}
	return nil
}

func flattenAutoScaleScalingSchedules(schedules []*iaas.AutoScaleScheduleScaling) []autoScaleScheduleScalingModel {
	if len(schedules) == 0 {
		return nil
	}

	var result []autoScaleScheduleScalingModel
	for _, s := range schedules {
		model := autoScaleScheduleScalingModel{
			Hour:   types.Int32Value(int32(s.Hour)),
			Minute: types.Int32Value(int32(s.Minute)),
			Action: types.StringValue(s.Action.String()),
		}
		if len(s.DayOfWeek) > 0 {
			model.DaysOfWeek = common.StringsToTset(common.MapTo(s.DayOfWeek, common.ToString))
		} else {
			model.DaysOfWeek = types.SetNull(types.StringType)
		}
		result = append(result, model)
	}
	return result
}
