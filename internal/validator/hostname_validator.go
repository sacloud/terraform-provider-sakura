// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringHostnameValidator struct {
	lengthValidator validator.String
}

var _ validator.String = stringHostnameValidator{}

func (v stringHostnameValidator) Description(_ context.Context) string {
	return "string must be a valid hostname"
}

func (v stringHostnameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringHostnameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	v.lengthValidator.ValidateString(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !regexp.MustCompile(`^(?i)([a-z0-9]+(-[a-z0-9]+)*)(\.[a-z0-9]+(-[a-z0-9]+)*)*$`).MatchString(value) {
		resp.Diagnostics.AddAttributeError(req.Path, v.Description(ctx), fmt.Sprintf("%q is not a valid hostname", value))
		return
	}
}

func HostnameValidator() stringHostnameValidator {
	return stringHostnameValidator{
		lengthValidator: stringvalidator.LengthBetween(1, 64),
	}
}
