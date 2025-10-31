// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type objectCustomFuncValidator struct {
	fun func(value types.Object) error
}

var _ validator.Object = objectCustomFuncValidator{}

func (v objectCustomFuncValidator) Description(_ context.Context) string {
	return "Validates a object attribute using a custom function"
}

func (v objectCustomFuncValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v objectCustomFuncValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue
	if err := v.fun(value); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, fmt.Sprintf("invalid value: %s", value), err.Error())
		return
	}
}

func ObjectFuncValidator(fun func(value types.Object) error) objectCustomFuncValidator {
	return objectCustomFuncValidator{fun: fun}
}
