// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"math/big"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// a plan that carries the state's active_version (a stale one included) sends nothing to the API
func TestAppResourceUpdateKeepsActiveVersionWhenUnchanged(t *testing.T) {
	ctx := t.Context()
	r := NewAppResource() // no client: any API call panics

	var sch resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sch)

	objectType := sch.Schema.Type().TerraformType(ctx).(tftypes.Object)
	timeoutsType := objectType.AttributeTypes["timeouts"].(tftypes.Object)

	rawWith := func(computed map[string]tftypes.Value, update any) tftypes.Value {
		attrs := map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "12345678-1234-1234-1234-123456789abc"),
			"cluster_id":     tftypes.NewValue(tftypes.String, "87654321-4321-4321-4321-cba987654321"),
			"name":           tftypes.NewValue(tftypes.String, "app"),
			"active_version": tftypes.NewValue(tftypes.Number, big.NewFloat(1)),
			"timeouts": tftypes.NewValue(timeoutsType, map[string]tftypes.Value{
				"create": tftypes.NewValue(tftypes.String, nil),
				"update": tftypes.NewValue(tftypes.String, update),
				"delete": tftypes.NewValue(tftypes.String, nil),
			}),
		}
		for name, v := range computed {
			attrs[name] = v
		}
		return tftypes.NewValue(objectType, attrs)
	}

	stateRaw := rawWith(map[string]tftypes.Value{
		"cluster_name":             tftypes.NewValue(tftypes.String, "cluster"),
		"desired_count":            tftypes.NewValue(tftypes.Number, big.NewFloat(1)),
		"scaling_cooldown_seconds": tftypes.NewValue(tftypes.Number, big.NewFloat(60)),
	}, nil)
	planRaw := rawWith(map[string]tftypes.Value{
		"cluster_name":             tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"desired_count":            tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"scaling_cooldown_seconds": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	}, "10m")

	req := resource.UpdateRequest{
		Config: tfsdk.Config{Schema: sch.Schema, Raw: planRaw},
		Plan:   tfsdk.Plan{Schema: sch.Schema, Raw: planRaw},
		State:  tfsdk.State{Schema: sch.Schema, Raw: stateRaw},
	}
	res := resource.UpdateResponse{State: tfsdk.State{Schema: sch.Schema, Raw: tftypes.NewValue(objectType, nil)}}

	r.Update(ctx, req, &res)

	if res.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", res.Diagnostics)
	}

	var got appResourceModel

	if diagnostics := res.State.Get(ctx, &got); diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if got.ActiveVersion.ValueInt32() != 1 || got.ClusterName.ValueString() != "cluster" || got.DesiredCount.ValueInt32() != 1 || got.ScalingCooldownSeconds.ValueInt32() != 60 {
		t.Fatalf("state = %+v, want the prior active_version and computed values", got.appModel)
	}

	update, diagnostics := got.Timeouts.Update(ctx, 0)

	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if update != 10*time.Minute {
		t.Fatalf("timeouts.update = %s, want 10m", update)
	}
}
