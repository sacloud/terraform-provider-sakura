// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type triggerBaseModel struct {
	common.SakuraBaseModel
	IconID types.String `tfsdk:"icon_id"`

	ProcessConfigurationID types.String           `tfsdk:"process_configuration_id"`
	Source                 types.String           `tfsdk:"source"`
	Types                  types.Set              `tfsdk:"types"`
	Conditions             []conditionObjectModel `tfsdk:"conditions"`
}

type conditionObjectModel struct {
	Key      types.String `tfsdk:"key"`
	Operator types.String `tfsdk:"op"`
	Values   types.Set    `tfsdk:"values"`
}

func (model *triggerBaseModel) updateState(data *v1.CommonServiceItem) error {
	model.UpdateBaseState(data.ID, data.Name, data.Description.Value, data.Tags)
	if iconID, ok := data.Icon.Value.ID.Get(); ok {
		model.IconID = types.StringValue(iconID)
	} else {
		model.IconID = types.StringNull()
	}

	trigger, ok := data.Settings.GetTriggerSettings()
	if !ok {
		return errors.New("invalid settings for Trigger")
	}

	model.Source = types.StringValue(trigger.Source)
	model.ProcessConfigurationID = types.StringValue(trigger.ProcessConfigurationID)

	if typesValue, ok := trigger.Types.Get(); ok {
		model.Types = common.StringsToTset(typesValue)
	} else {
		model.Types = types.SetNull(types.StringType)
	}

	if conditions, ok := trigger.Conditions.Get(); ok {
		model.Conditions = make([]conditionObjectModel, 0, len(conditions))
		for _, c := range conditions {
			if eqItem, ok := c.GetTriggerConditionEq(); ok {
				model.Conditions = append(model.Conditions, conditionObjectModel{
					Key:      types.StringValue(eqItem.Key),
					Operator: types.StringValue(string(eqItem.Op)),
					Values:   common.StringsToTset(eqItem.Values),
				})
				continue
			}

			if inItem, ok := c.GetTriggerConditionIn(); ok {
				model.Conditions = append(model.Conditions, conditionObjectModel{
					Key:      types.StringValue(inItem.Key),
					Operator: types.StringValue(string(inItem.Op)),
					Values:   common.StringsToTset(inItem.Values),
				})
				continue
			}

			return errors.New("unknown condition kind in Trigger conditions")
		}
	} else {
		model.Conditions = nil
	}

	return nil
}
