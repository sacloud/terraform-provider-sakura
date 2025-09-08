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

package icon

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type iconBaseModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
	Tags types.Set    `tfsdk:"tags"`
}

func (model *iconBaseModel) updateState(icon *iaas.Icon) {
	model.ID = types.StringValue(icon.ID.String())
	model.Name = types.StringValue(icon.Name)
	model.Tags = common.StringsToTset(icon.Tags)
	model.URL = types.StringValue(icon.URL)
}
