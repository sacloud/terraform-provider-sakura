// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringSakuraIDTypeValidator struct{}

var _ validator.String = stringSakuraIDTypeValidator{}

func (v stringSakuraIDTypeValidator) Description(_ context.Context) string {
	return "string must be a valid Sakura Cloud ID (number only)"
}

func (v stringSakuraIDTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringSakuraIDTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()
	_, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, v.Description(ctx), fmt.Sprintf("%q is not a valid SakuraCloud ID", value))
		return
	}
}

func SakuraIDValidator() stringSakuraIDTypeValidator {
	return stringSakuraIDTypeValidator{}
}
