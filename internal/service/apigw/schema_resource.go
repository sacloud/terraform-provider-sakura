// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

func schemaResourceAPIGWName(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: desc.Sprintf("The name of the %s", name),
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 255),
			stringvalidator.RegexMatches(regexp.MustCompile(`^[\p{L}\p{N}._\-]+$`), "can only contain unicode letters, numbers, dots, underscores, and hyphens."),
		},
	}
}

func schemaResourceAPIGWCreatedAt(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The creation timestamp of the %s", name),
	}
}

func schemaResourceAPIGWUpdatedAt(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The last update timestamp of the %s", name),
	}
}

func schemaResourceAPIGWCert(description string, required bool) schema.Attribute {
	s := schema.SingleNestedAttribute{
		Description: description,
		Attributes: map[string]schema.Attribute{
			"cert_wo": schema.StringAttribute{
				Required:    true,
				WriteOnly:   true,
				Description: "PEM encoded certificate data",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("cert_wo_version")),
				},
			},
			"key_wo": schema.StringAttribute{
				Required:    true,
				WriteOnly:   true,
				Description: "PEM encoded private key data",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("cert_wo_version")),
				},
			},
			"cert_wo_version": schema.Int32Attribute{
				Optional:    true,
				Description: "The version of the certificate. This value must be greater than 0 when set. Increment this when changing certificate",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
					int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("cert_wo")),
					int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("key_wo")),
				},
			},
			"expired_at": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration timestamp",
			},
		},
	}
	if required {
		s.Required = true
	} else {
		s.Optional = true
	}
	return s
}

func schemaResourceAPIGWListFromTo() schema.Attribute {
	return schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"from": schema.StringAttribute{
					Required:    true,
					Description: "Source name to rename from",
				},
				"to": schema.StringAttribute{
					Required:    true,
					Description: "Destination name to rename to",
				},
			},
		},
	}
}

func schemaResourceAPIGWListKV() schema.Attribute {
	return schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"key": schema.StringAttribute{
					Required:    true,
					Description: "The target key",
				},
				"value": schema.StringAttribute{
					Required:    true,
					Description: "The value for the key",
				},
			},
		},
	}
}

func schemaResourceAPIGWIfStatusCode() schema.Attribute {
	return schema.SetAttribute{
		ElementType: types.Int32Type,
		Optional:    true,
		Description: "Apply only for these HTTP status codes",
		Validators: []validator.Set{
			setvalidator.ValueInt32sAre(int32validator.Between(100, 900)),
		},
	}
}
