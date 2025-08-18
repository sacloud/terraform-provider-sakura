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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

func schemaResourceId(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The ID of the %s.", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func schemaResourceName(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: desc.Sprintf("The name of the %s.", name),
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 64),
		},
	}
}

func schemaResourceDescription(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true, // FrameworkはSDK v2とは違ってComputedをつけないとnullに値をセットしようとしてエラーになる
		Description: desc.Sprintf("The description of the %s. %s", name, desc.Length(1, 512)),
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 512),
		},
	}
}

func schemaResourceIconID(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The icon id to attach to the %s", name),
		Validators: []validator.String{
			sakuraIDValidator(),
		},
	}
}

func schemaResourceTags(name string) schema.Attribute {
	return schema.SetAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The tags of the %s.", name),
	}
}

func schemaResourceZone(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The name of zone that the %s will be created (e.g. `is1a`, `tk1a`)", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIfConfigured(),
		},
	}
}

func schemaPacketFilterExpression() schema.Block {
	return schema.ListNestedBlock{
		Description: "List of packet filter expressions",
		Validators: []validator.List{
			listvalidator.SizeAtMost(30),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"protocol": schema.StringAttribute{
					Required:    true,
					Description: desc.Sprintf("The protocol used for filtering. This must be one of [%s]", iaastypes.PacketFilterProtocolStrings),
					Validators: []validator.String{
						stringvalidator.OneOf(iaastypes.PacketFilterProtocolStrings...),
					},
				},
				"source_network": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
					Description: "A source IP address or CIDR block used for filtering (e.g. `192.0.2.1`, `192.0.2.0/24`)",
				},
				"source_port": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
					Description: "A source port number or port range used for filtering (e.g. `1024`, `1024-2048`)",
				},
				"destination_port": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
					Description: "A destination port number or port range used for filtering (e.g. `1024`, `1024-2048`)",
				},
				"allow": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
					Description: "The flag to allow the packet through the filter",
				},
				"description": schemaResourceDescription("Packet Filter Expression"),
			},
		},
	}
}
