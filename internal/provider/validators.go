// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sakura

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringSakuraIDTypeValidator struct{}

var _ validator.String = stringSakuraIDTypeValidator{}

func (v stringSakuraIDTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string must be a valid Sakura Cloud ID (number only)")
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
		resp.Diagnostics.AddAttributeError(req.Path, v.Description(ctx), value)
		return
	}
}

func sakuraIDValidator() stringSakuraIDTypeValidator {
	return stringSakuraIDTypeValidator{}
}

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

func stringFuncValidator(fun func(value string) error) stringCustomFuncValidator {
	return stringCustomFuncValidator{fun: fun}
}

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

func int64FuncValidator(fun func(value int64) error) int64CustomFuncValidator {
	return int64CustomFuncValidator{fun: fun}
}
