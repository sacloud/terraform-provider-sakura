// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

func SchemaResourceId(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The ID of the %s.", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func SchemaResourceName(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: desc.Sprintf("The name of the %s.", name),
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 64),
		},
	}
}

func SchemaResourceDescription(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true, // FrameworkはSDK v2とは違ってComputedをつけないとnullに値をセットしようとしてエラーになる
		Description: desc.Sprintf("The description of the %s. %s", name, desc.Length(1, 512)),
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 512),
		},
	}
}

func SchemaResourceIconID(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: desc.Sprintf("The icon id to attach to the %s", name),
		Validators: []validator.String{
			sacloudvalidator.SakuraIDValidator(),
		},
	}
}

func SchemaResourceServerID(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The id of the server connected to the %s", name),
		Validators: []validator.String{
			sacloudvalidator.SakuraIDValidator(),
		},
	}
}

func SchemaResourceSwitchID(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: desc.Sprintf("The id of the vSwitch to which the %s connects", name),
		Validators: []validator.String{
			sacloudvalidator.SakuraIDValidator(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func SchemaResourceTags(name string) schema.Attribute {
	return schema.SetAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The tags of the %s.", name),
	}
}

func SchemaResourceZone(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The name of zone that the %s will be created (e.g. `is1a`, `tk1a`)", name),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIfConfigured(),
		},
	}
}

func SchemaResourceSize(name string, defaultValue int64, validSizes ...int64) schema.Attribute {
	s := schema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The size of %s in GiB", name),
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplaceIfConfigured(),
		},
	}
	if defaultValue > 0 {
		s.Default = int64default.StaticInt64(defaultValue)
	}
	if len(validSizes) > 0 {
		s.Validators = []validator.Int64{
			int64validator.OneOf(validSizes...),
		}
		s.Description = desc.Sprintf("%s. This must be one of [%s]", s.Description, validSizes)
	}

	return s
}

func SchemaResourcePlan(name string, defaultValue string, plans []string) schema.Attribute {
	s := schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.ResourcePlan(name, plans),
		Validators: []validator.String{
			stringvalidator.OneOf(plans...),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIfConfigured(),
		},
	}
	if defaultValue != "" {
		s.Default = stringdefault.StaticString(defaultValue)
	}

	return s
}

func SchemaResourceMonitoringSuite(name string) schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The monitoring suite settings of the %s.", name),
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable sending signals to Monitoring Suite",
			},
		},
	}
}
