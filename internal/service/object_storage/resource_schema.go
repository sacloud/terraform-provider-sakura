// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

func SchemaResourceSiteID(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: "The ID of the Object Storage Site.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func SchemaResourceEndpoint(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The endpoint for the %s. Currently, only `s3.isk01.sakurastorage.jp` is supported as the endpoint.", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIfConfigured(),
		},
	}
}

func SchemaResourceAccessKey(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Sensitive:   true,
		Description: desc.Sprintf("The access key for the %s.", name),
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_key")),
		},
	}
}

func SchemaResourceSecretKey(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Sensitive:   true,
		Description: desc.Sprintf("The secret key for the %s.", name),
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("access_key")),
		},
	}
}

func SchemaResourceBucket(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: desc.Sprintf("The bucket of the %s.", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}
