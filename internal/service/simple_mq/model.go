// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_mq

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/simplemq-api-go"
	"github.com/sacloud/simplemq-api-go/apis/v1/queue"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type simpleMqBaseModel struct {
	common.SakuraBaseModel
	IconID                   types.String `tfsdk:"icon_id"`
	VisibilityTimeoutSeconds types.Int64  `tfsdk:"visibility_timeout_seconds"`
	ExpireSeconds            types.Int64  `tfsdk:"expire_seconds"`
}

func (model *simpleMqBaseModel) updateState(data *queue.CommonServiceItem) {
	model.ID = types.StringValue(simplemq.GetQueueID(data))
	model.Name = types.StringValue(simplemq.GetQueueName(data))
	model.Description = types.StringValue(data.Description.Value)
	model.VisibilityTimeoutSeconds = types.Int64Value(int64(data.Settings.VisibilityTimeoutSeconds))
	model.ExpireSeconds = types.Int64Value(int64(data.Settings.ExpireSeconds))
	if iconID, ok := data.Icon.Value.ID.Get(); ok {
		id, ok := iconID.GetString()
		if !ok {
			id = strconv.Itoa(iconID.Int)
		}
		model.IconID = types.StringValue(id)
	} else {
		model.IconID = types.StringNull()
	}
	model.Tags = common.StringsToTset(data.Tags)
}
