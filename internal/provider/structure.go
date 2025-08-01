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

	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func TagsToStringList(d tftypes.Set) []string {
	var tags []string
	for _, v := range d.Elements() {
		if vStr, ok := v.(tftypes.String); ok && !vStr.IsNull() && !vStr.IsUnknown() {
			tags = append(tags, vStr.ValueString())
		}
	}
	return tags
}

func TagsToTFSet(ctx context.Context, tags []string) tftypes.Set {
	setValue, _ := tftypes.SetValueFrom(ctx, tftypes.StringType, tags)
	return setValue
}
