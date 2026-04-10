// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

func schemaResourceAlertId() schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: "The resource ID of the Alert Project.",
		Validators: []validator.String{
			sacloudvalidator.SakuraIDValidator(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}
