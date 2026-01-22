// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

func schemaDataSourceAPIGWCreatedAt(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The creation timestamp of the %s", name),
	}
}

func schemaDataSourceAPIGWUpdatedAt(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The last update timestamp of the %s", name),
	}
}

func schemaDataSourceAPIGWListFromTo() schema.Attribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"from": schema.StringAttribute{
					Computed:    true,
					Description: "Source name to rename from",
				},
				"to": schema.StringAttribute{
					Computed:    true,
					Description: "Destination name to rename to",
				},
			},
		},
	}
}

func schemaDataSourceAPIGWListKV() schema.Attribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"key": schema.StringAttribute{
					Computed:    true,
					Description: "The target key",
				},
				"value": schema.StringAttribute{
					Computed:    true,
					Description: "The value for the key",
				},
			},
		},
	}
}

func schemaDataSourceAPIGWIfStatusCode() schema.Attribute {
	return schema.SetAttribute{
		ElementType: types.Int32Type,
		Computed:    true,
		Description: "Apply only for these HTTP status codes",
	}
}
