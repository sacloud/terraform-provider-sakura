// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/simple-notification-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type routingBaseModel struct {
	common.SakuraBaseModel
	IconID        types.String      `tfsdk:"icon_id"`
	MatchLabels   []matchLabelModel `tfsdk:"match_labels"`
	SourceID      types.String      `tfsdk:"source_id"`
	TargetGroupID types.String      `tfsdk:"target_group_id"`
}

type matchLabelModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (model *routingBaseModel) updateState(data *v1.CommonServiceItem) error {
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

	routingSettings, _ := data.Settings.GetRoutingSettings()
	model.MatchLabels = flattenMatchLabels(routingSettings.MatchLabels)
	model.SourceID = types.StringValue(routingSettings.SourceID)
	model.TargetGroupID = types.StringValue(routingSettings.TargetGroupID)

	return nil
}

func flattenMatchLabels(matchLabels []v1.RoutingSettingsMatchLabelsItem) []matchLabelModel {
	result := make([]matchLabelModel, len(matchLabels))
	for i, ml := range matchLabels {
		result[i] = matchLabelModel{
			Name:  types.StringValue(ml.Name),
			Value: types.StringValue(ml.Value),
		}
	}
	return result
}

func expandMatchLabels(models []matchLabelModel) []v1.RoutingSettingsMatchLabelsItem {
	result := make([]v1.RoutingSettingsMatchLabelsItem, len(models))
	for i, ml := range models {
		result[i] = v1.RoutingSettingsMatchLabelsItem{
			Name:  ml.Name.ValueString(),
			Value: ml.Value.ValueString(),
		}
	}
	return result
}
