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

package event_bus

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type processConfigurationBaseModel struct {
	common.SakuraBaseModel
	// TODO: iconはsdkで未対応
	// IconID types.String `tfsdk:"icon_id"`

	Destination types.String `tfsdk:"destination"`
	Parameters  types.String `tfsdk:"parameters"`
}

type processConfigurationWithCredentialsBaseModel struct {
	processConfigurationBaseModel

	SimpleNotificationAccessToken       types.String `tfsdk:"simplenotification_access_token"`
	SimpleNotificationAccessTokenSecret types.String `tfsdk:"simplenotification_access_token_secret"`
	SimpleMQAPIKey                      types.String `tfsdk:"simplemq_api_key"`
}

func (model *processConfigurationBaseModel) updateState(data *v1.ProcessConfiguration) {
	id := strconv.FormatInt(data.ID, 10)
	model.ID = types.StringValue(id)
	model.Name = types.StringValue(data.Name)
	model.Description = types.StringValue(data.Description)
	model.Tags = common.StringsToTset(data.Tags)

	model.Destination = types.StringValue(string(data.Settings.Destination))
	model.Parameters = types.StringValue(data.Settings.Parameters)
}
