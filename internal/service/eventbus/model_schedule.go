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

package eventbus

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type scheduleBaseModel struct {
	common.SakuraBaseModel
	// TODO: iconはsdkで未対応
	// IconID types.String `tfsdk:"icon_id"`

	ProcessConfigurationID types.String `tfsdk:"process_configuration_id"`
	RecurringStep          types.Int64  `tfsdk:"recurring_step"`
	RecurringUnit          types.String `tfsdk:"recurring_unit"`
	StartsAt               types.Int64  `tfsdk:"starts_at"`
}

func (model *scheduleBaseModel) updateState(data *v1.Schedule) {
	id := strconv.FormatInt(data.ID, 10)
	model.UpdateBaseState(id, data.Name, data.Description, data.Tags)

	model.ProcessConfigurationID = types.StringValue(data.Settings.ProcessConfigurationID)
	model.StartsAt = types.Int64Value(data.Settings.StartsAt)

	// TODO: 現状はcrontabの指定が未対応なので実質Requiredとなっているが本来Optionalなので、
	// SDKの更新に伴いcrontab対応をする際に指定のあるなしで適切に分岐をする or 代入をdata sourceのみに移す。
	// https://github.com/sacloud/terraform-provider-sakura/pull/8#discussion_r2404952468
	model.RecurringStep = types.Int64Value(int64(data.Settings.RecurringStep))
	model.RecurringUnit = types.StringValue(string(data.Settings.RecurringUnit))
}
