// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package packet_filter

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

func schemaPacketFilterExpression() schema.Attribute {
	return schema.ListNestedAttribute{
		Optional:    true,
		Description: "List of packet filter expressions",
		Validators: []validator.List{
			listvalidator.SizeAtMost(30),
		},
		NestedObject: schema.NestedAttributeObject{
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
				"description": common.SchemaResourceDescription("Packet Filter Expression"),
			},
		},
	}
}
