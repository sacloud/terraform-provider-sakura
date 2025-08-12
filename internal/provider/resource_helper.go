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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func updateResourceByRead(ctx context.Context, r resource.Resource, state *tfsdk.State, diags *diag.Diagnostics, id string) {
	updateResourceByReadWithZone(ctx, r, state, diags, id, "")
}

func updateResourceByReadWithZone(ctx context.Context, r resource.Resource, state *tfsdk.State, diags *diag.Diagnostics, id string, zone string) {
	state.SetAttribute(ctx, path.Root("id"), id)
	if len(zone) > 0 {
		state.SetAttribute(ctx, path.Root("zone"), zone)
	}

	newReadResp := &resource.ReadResponse{State: *state, Diagnostics: *diags}
	r.Read(ctx, resource.ReadRequest{State: *state}, newReadResp)
	*diags = newReadResp.Diagnostics
	*state = newReadResp.State
}
