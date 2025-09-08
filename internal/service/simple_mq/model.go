// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	model.VisibilityTimeoutSeconds = types.Int64Value(int64(data.Settings.VisibilityTimeoutSeconds))
	model.ExpireSeconds = types.Int64Value(int64(data.Settings.ExpireSeconds))
	if v, ok := data.Description.Value.GetString(); ok {
		model.Description = types.StringValue(v)
	}
	if iconID, ok := data.Icon.Value.Icon1.ID.Get(); ok {
		id, ok := iconID.GetString()
		if !ok {
			id = strconv.Itoa(iconID.Int)
		}
		model.IconID = types.StringValue(id)
	} else {
		model.IconID = types.StringValue("")
	}
	model.Tags = common.StringsToTset(data.Tags)
}
