// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type int64CustomFuncValidator struct {
	fun func(value int64) error
}

var _ validator.Int64 = int64CustomFuncValidator{}

func (v int64CustomFuncValidator) Description(_ context.Context) string {
	return "Validates an int64 attribute using a custom function"
}

func (v int64CustomFuncValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v int64CustomFuncValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueInt64()
	if err := v.fun(value); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, fmt.Sprintf("invalid value: %d", value), err.Error())
		return
	}
}

func Int64FuncValidator(fun func(value int64) error) int64CustomFuncValidator {
	return int64CustomFuncValidator{fun: fun}
}
