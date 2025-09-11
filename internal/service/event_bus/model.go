// Copyright 2016-2025 terraform-provider-sakuracloud authors
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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

type eventBusProcessConfigurationBaseModel struct {
	common.SakuraBaseModel
	IconID types.String `tfsdk:"icon_id"`

	Destination types.String `tfsdk:"destination"`
	Parameters  types.String `tfsdk:"parameters"`

	// GroupName         types.String `tfsdk:"group_name"`
	// Message           types.String `tfsdk:"message"`
	// AccessToken       types.String `tfsdk:"access_token"`
	// AccessTokenSecret types.String `tfsdk:"access_token_secret"`
}
