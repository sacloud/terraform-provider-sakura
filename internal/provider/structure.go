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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaastypes "github.com/sacloud/iaas-api-go/types"
)

func sakuraCloudID(id string) iaastypes.ID {
	return iaastypes.StringID(id)
}

func getApiClientFromProvider(providerData any, diags *diag.Diagnostics) *APIClient {
	if providerData == nil {
		return nil
	}

	apiclient, ok := providerData.(*APIClient)
	if !ok {
		diags.AddError("Unexpected ProviderData type", "Expected *APIClient.")
		return nil
	}

	return apiclient
}

func tlistToStrings(d types.List) []string {
	var tags []string
	for _, v := range d.Elements() {
		if vStr, ok := v.(types.String); ok && !vStr.IsNull() && !vStr.IsUnknown() {
			tags = append(tags, vStr.ValueString())
		}
	}
	return tags
}

func tsetToStrings(d types.Set) []string {
	var tags []string
	for _, v := range d.Elements() {
		if vStr, ok := v.(types.String); ok && !vStr.IsNull() && !vStr.IsUnknown() {
			tags = append(tags, vStr.ValueString())
		}
	}
	return tags
}

func stringsToTset(ctx context.Context, tags []string) types.Set {
	setValue, _ := types.SetValueFrom(ctx, types.StringType, tags)
	return setValue
}
