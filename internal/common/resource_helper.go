// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/sacloud/packages-go/mutexkv"
)

var SakuraMutexKV = mutexkv.NewMutexKV()

// SDK v2のHasChangeの代替。複雑なGoの値の比較に使う。Terraformの値の比較にはEqualを使う
func HasChange(x, y any) bool {
	return !cmp.Equal(x, y)
}

func UpdateResourceByRead(ctx context.Context, r resource.Resource, state *tfsdk.State, diags *diag.Diagnostics, id string) {
	UpdateResourceByReadWithZone(ctx, r, state, diags, id, "")
}

func UpdateResourceByReadWithZone(ctx context.Context, r resource.Resource, state *tfsdk.State, diags *diag.Diagnostics, id string, zone string) {
	state.SetAttribute(ctx, path.Root("id"), id)
	if len(zone) > 0 {
		state.SetAttribute(ctx, path.Root("zone"), zone)
	}

	newReadResp := &resource.ReadResponse{State: *state, Diagnostics: *diags}
	r.Read(ctx, resource.ReadRequest{State: *state}, newReadResp)
	*diags = newReadResp.Diagnostics
	*state = newReadResp.State
}
