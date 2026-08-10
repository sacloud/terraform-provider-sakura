// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/sacloud-sdk-go/srn"
)

type srnValidator struct{}

var _ validator.String = srnValidator{}

func (v srnValidator) Description(_ context.Context) string {
	return "string must be a valid SRN"
}

func (v srnValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v srnValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !srn.IsSRN(value) {
		resp.Diagnostics.AddAttributeError(req.Path, v.Description(ctx), fmt.Sprintf("%q is not a valid SRN", value))
		return
	}
}

func SRNValidator() srnValidator {
	return srnValidator{}
}
