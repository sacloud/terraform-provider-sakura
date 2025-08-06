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

var _ validator.String = stringSakuraIDTypeValidator{}

type stringSakuraIDTypeValidator struct{}

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
