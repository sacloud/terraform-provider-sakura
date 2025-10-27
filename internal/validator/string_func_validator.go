// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringCustomFuncValidator struct {
	fun func(value string) error
}

var _ validator.String = stringCustomFuncValidator{}

func (v stringCustomFuncValidator) Description(_ context.Context) string {
	return "Validates a string attribute using a custom function"
}

func (v stringCustomFuncValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringCustomFuncValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()
	if err := v.fun(value); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, fmt.Sprintf("invalid value: %s", value), err.Error())
		return
	}
}

func StringFuncValidator(fun func(value string) error) stringCustomFuncValidator {
	return stringCustomFuncValidator{fun: fun}
}
