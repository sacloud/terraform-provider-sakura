// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

func schemaResourceAddonLocation(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: desc.Sprintf("The location of the %s", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func schemaResourceAddonSKU(name string, values []int32) schema.Attribute {
	return schema.Int32Attribute{
		Required:    true,
		Description: desc.Sprintf("The SKU of the %s", name),
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.RequiresReplace(),
		},
		Validators: []validator.Int32{
			int32validator.OneOf(values...),
		},
	}
}

func schemaResourceAddonPricingLevel(name string, values []int32) schema.Attribute {
	return schema.Int32Attribute{
		Required:    true,
		Description: desc.Sprintf("The pricing level of the %s", name),
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.RequiresReplace(),
		},
		Validators: []validator.Int32{
			int32validator.OneOf(values...),
		},
	}
}

func schemaResourceAddonDeploymentName(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The deployment name of the %s", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func schemaResourceAddonURL(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The URL of the %s", name),
	}
}
