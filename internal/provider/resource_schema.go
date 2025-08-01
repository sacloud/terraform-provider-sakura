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

package sakura

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func schemaResourceId(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: fmt.Sprintf("The ID of the %s.", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func schemaResourceName(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: fmt.Sprintf("The name of the %s.", name),
	}
}

func schemaResourceDescription(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: fmt.Sprintf("The description of the %s.", name),
	}
}

func schemaResourceTags(name string) schema.Attribute {
	return schema.SetAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true, // FrameworkはSDK v2とは違ってComputedをつけないとnullに値をセットしようとしてエラーになる
		Description: fmt.Sprintf("The tags of the %s.", name),
	}
}
