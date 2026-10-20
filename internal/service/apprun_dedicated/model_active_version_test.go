// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"errors"
	"math/big"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	app "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/application"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/version"
)

// write-only attributes have structural rules the framework checks only at runtime
func TestActiveVerResourceSchemaValidateImplementation(t *testing.T) {
	var res resource.SchemaResponse

	NewActiveVersionResource().Schema(t.Context(), resource.SchemaRequest{}, &res)

	if res.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", res.Diagnostics)
	}
	if diagnostics := res.Schema.ValidateImplementation(t.Context()); diagnostics.HasError() {
		t.Fatalf("invalid schema: %v", diagnostics)
	}
}

// a change of timeouts alone keeps the active version; no API is called
func TestActiveVerResourceUpdateKeepsVersionForTimeoutsOnlyChange(t *testing.T) {
	ctx := t.Context()
	r := NewActiveVersionResource() // no client: any API call panics

	var sch resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sch)

	objectType := sch.Schema.Type().TerraformType(ctx).(tftypes.Object)
	timeoutsType := objectType.AttributeTypes["timeouts"].(tftypes.Object)

	timeoutsWith := func(update any) tftypes.Value {
		return tftypes.NewValue(timeoutsType, map[string]tftypes.Value{
			"create": tftypes.NewValue(tftypes.String, nil),
			"update": tftypes.NewValue(tftypes.String, update),
			"delete": tftypes.NewValue(tftypes.String, nil),
		})
	}
	rawWith := func(overrides map[string]tftypes.Value) tftypes.Value {
		attrs := map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, "12345678-1234-1234-1234-123456789abc"),
			"application_id": tftypes.NewValue(tftypes.String, "12345678-1234-1234-1234-123456789abc"),
			"cpu":            tftypes.NewValue(tftypes.Number, big.NewFloat(100)),
			"memory":         tftypes.NewValue(tftypes.Number, big.NewFloat(256)),
			"scaling_mode":   tftypes.NewValue(tftypes.String, "manual"),
			"image":          tftypes.NewValue(tftypes.String, "nginx:latest"),
		}
		for name, typ := range objectType.AttributeTypes {
			if _, ok := attrs[name]; !ok {
				attrs[name] = tftypes.NewValue(typ, nil)
			}
		}
		for name, v := range overrides {
			attrs[name] = v
		}
		return tftypes.NewValue(objectType, attrs)
	}

	stateRaw := rawWith(map[string]tftypes.Value{
		"version":           tftypes.NewValue(tftypes.Number, big.NewFloat(2)),
		"created_at":        tftypes.NewValue(tftypes.String, "19 Oct 26 10:00 JST"),
		"active_node_count": tftypes.NewValue(tftypes.Number, big.NewFloat(1)),
		"timeouts":          timeoutsWith(nil),
	})
	planRaw := rawWith(map[string]tftypes.Value{
		"version":           tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"created_at":        tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"active_node_count": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"timeouts":          timeoutsWith("10m"),
	})

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

	var got activeVerResourceModel

	if diagnostics := res.State.Get(ctx, &got); diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if got.Version.ValueInt32() != 2 || got.CreatedAt.ValueString() != "19 Oct 26 10:00 JST" || got.ActiveNodeCount.ValueInt64() != 1 {
		t.Fatalf("version = %s, created_at = %s, active_node_count = %s, want the state's 2, 19 Oct 26 10:00 JST, 1", got.Version, got.CreatedAt, got.ActiveNodeCount)
	}

	update, diagnostics := got.Timeouts.Update(ctx, 0)

	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if update != 10*time.Minute {
		t.Fatalf("timeouts.update = %s, want 10m", update)
	}
}

// the ID is the application ID, and an empty cmd / exposed_ports / host of the plan or
// the state survives updateState, which reads them back as null
func TestActiveVerResourceModelUpdateState(t *testing.T) {
	ctx := t.Context()
	aid := appID(uuid.MustParse("12345678-1234-1234-1234-123456789abc"))
	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	emptySet := types.SetValueMust(types.StringType, []attr.Value{})

	empty := activeVerResourceModel{verResourceModel: verResourceModel{verModel: verModel{Cmd: emptyList, ExposedPorts: []exposedPortModel{}}}}

	if diagnostics := empty.updateState(ctx, &version.VersionDetail{Version: 3}, aid); diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if got := empty.ID.ValueString(); got != "12345678-1234-1234-1234-123456789abc" {
		t.Fatalf("ID = %q, want the application ID", got)
	}
	if empty.Cmd.IsNull() || len(empty.Cmd.Elements()) != 0 || empty.ExposedPorts == nil || len(empty.ExposedPorts) != 0 {
		t.Fatalf("Cmd = %s, ExposedPorts = %+v, want both empty, not null", empty.Cmd, empty.ExposedPorts)
	}

	withHost := activeVerResourceModel{verResourceModel: verResourceModel{verModel: verModel{
		ExposedPorts: []exposedPortModel{{TargetPort: types.Int32Value(80), Host: emptySet}},
	}}}
	detail := version.VersionDetail{ExposedPorts: []version.ExposedPort{{TargetPort: 80}}}

	if diagnostics := withHost.updateState(ctx, &detail, aid); diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if len(withHost.ExposedPorts) != 1 || withHost.ExposedPorts[0].Host.IsNull() || len(withHost.ExposedPorts[0].Host.Elements()) != 0 {
		t.Fatalf("ExposedPorts = %+v, want one port with an empty host set", withHost.ExposedPorts)
	}

	// ports are matched by target_port, not by position
	reordered := activeVerResourceModel{verResourceModel: verResourceModel{verModel: verModel{
		ExposedPorts: []exposedPortModel{
			{TargetPort: types.Int32Value(80), Host: emptySet},
			{TargetPort: types.Int32Value(8080)},
		},
	}}}
	detail = version.VersionDetail{ExposedPorts: []version.ExposedPort{{TargetPort: 8080}, {TargetPort: 80}}}

	if diagnostics := reordered.updateState(ctx, &detail, aid); diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if len(reordered.ExposedPorts) != 2 || !reordered.ExposedPorts[0].Host.IsNull() || reordered.ExposedPorts[1].Host.IsNull() || len(reordered.ExposedPorts[1].Host.Elements()) != 0 {
		t.Fatalf("ExposedPorts = %+v, want host null for 8080 and an empty set for 80", reordered.ExposedPorts)
	}

	var null activeVerResourceModel

	if diagnostics := null.updateState(ctx, &version.VersionDetail{}, aid); diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if !null.Cmd.IsNull() || null.ExposedPorts != nil {
		t.Fatalf("Cmd = %s, ExposedPorts = %+v, want both null", null.Cmd, null.ExposedPorts)
	}
}

// variables the active version no longer has, or has with the other secrecy, are dropped;
// nil stays nil and an emptied list stays an empty list
func TestDropMissingVars(t *testing.T) {
	env := func(key string, secret bool) envVarModel {
		return envVarModel{Key: types.StringValue(key), Value: types.StringValue("v"), Secret: types.BoolValue(secret)}
	}
	sec := func(key string) secretVarModel {
		return secretVarModel{Key: types.StringValue(key), ValueWOVersion: types.Int32Value(1)}
	}
	live := func(key string, secret bool) version.EnvironmentVariable {
		return version.EnvironmentVariable{Key: key, Secret: secret}
	}
	envKeys := func(vars []envVarModel) []string {
		if vars == nil {
			return nil
		}
		keys := []string{}
		for _, e := range vars {
			keys = append(keys, e.Key.ValueString())
		}
		return keys
	}
	secretKeys := func(vars []secretVarModel) []string {
		if vars == nil {
			return nil
		}
		keys := []string{}
		for _, s := range vars {
			keys = append(keys, s.Key.ValueString())
		}
		return keys
	}
	same := func(got, want []string) bool {
		return (got == nil) == (want == nil) && slices.Equal(got, want)
	}

	for name, c := range map[string]struct {
		envVars    []envVarModel
		secretVars []secretVarModel
		live       []version.EnvironmentVariable
		wantEnv    []string
		wantSecret []string
	}{
		"missing keys dropped":   {[]envVarModel{env("A", false), env("GONE", false)}, []secretVarModel{sec("S"), sec("SGONE")}, []version.EnvironmentVariable{live("A", false), live("S", true)}, []string{"A"}, []string{"S"}},
		"nil stays nil":          {nil, nil, []version.EnvironmentVariable{live("A", false)}, nil, nil},
		"emptied stays empty":    {[]envVarModel{env("GONE", false)}, []secretVarModel{sec("SGONE")}, nil, []string{}, []string{}},
		"secret turned plain":    {nil, []secretVarModel{sec("X")}, []version.EnvironmentVariable{live("X", false)}, nil, []string{}},
		"plain turned secret":    {[]envVarModel{env("X", false)}, nil, []version.EnvironmentVariable{live("X", true)}, []string{}, nil},
		"deprecated secret kept": {[]envVarModel{env("X", true)}, nil, []version.EnvironmentVariable{live("X", true)}, []string{"X"}, nil},
	} {
		state := activeVerResourceModel{verResourceModel: verResourceModel{verModel: verModel{EnvVars: c.envVars}, SecretVars: c.secretVars}}

		state.dropMissingVars(&version.VersionDetail{EnvVars: c.live})

		if got := envKeys(state.EnvVars); !same(got, c.wantEnv) {
			t.Fatalf("%s: env_vars keys = %#v, want %#v", name, got, c.wantEnv)
		}
		if got := secretKeys(state.SecretVars); !same(got, c.wantSecret) {
			t.Fatalf("%s: secret_vars keys = %#v, want %#v", name, got, c.wantSecret)
		}
	}
}

func testPlacement(desired, placed []int64) (p app.Placement) {
	for _, v := range desired {
		p.Desired.Containers = append(p.Desired.Containers, v1.ApplicationDesiredContainer{ApplicationVersion: v})
	}
	for _, v := range placed {
		p.ContainersStats.Containers = append(p.ContainersStats.Containers, v1.ApplicationContainerSummary{ApplicationVersion: v})
	}
	return
}

// no node may have a container of another version; whether the new one is up does not matter
func TestIsPlacementConverged(t *testing.T) {
	two := int32(2)

	for name, c := range map[string]struct {
		placements []app.Placement
		version    *int32
		want       bool
	}{
		"old version still desired":    {[]app.Placement{testPlacement([]int64{1, 2}, []int64{2})}, &two, false},
		"old version still placed":     {[]app.Placement{testPlacement([]int64{2}, []int64{1})}, &two, false},
		"old version on another node":  {[]app.Placement{testPlacement([]int64{2}, []int64{2}), testPlacement(nil, []int64{1})}, &two, false},
		"empty node":                   {[]app.Placement{testPlacement(nil, nil)}, &two, true},
		"no nodes":                     {nil, &two, true},
		"new version not yet placed":   {[]app.Placement{testPlacement([]int64{2}, nil)}, &two, true},
		"converged":                    {[]app.Placement{testPlacement([]int64{2}, []int64{2}), testPlacement([]int64{2}, []int64{2, 2})}, &two, true},
		"deactivated but still placed": {[]app.Placement{testPlacement(nil, []int64{2})}, nil, false},
		"deactivated and drained":      {[]app.Placement{testPlacement(nil, nil)}, nil, true},
	} {
		if got := isPlacementConverged(c.placements, c.version); got != c.want {
			t.Fatalf("%s: isPlacementConverged() = %t, want %t", name, got, c.want)
		}
	}
}

// embedded nil interface: any call but Containers panics
type testApplicationAPI struct {
	app.ApplicationAPI
	containers func() ([]app.Placement, error)
}

func (a testApplicationAPI) Containers(context.Context, v1.ApplicationID) ([]app.Placement, error) {
	return a.containers()
}

// the wait ends on convergence, an API error or a done context, reported as the timeout
func TestWaitForActiveVersion(t *testing.T) {
	two := int32(2)
	stale := []app.Placement{testPlacement([]int64{1}, []int64{1})}
	apiError := errors.New("api error")

	var aid appID

	for name, c := range map[string]struct {
		run     func(t *testing.T, version *int32) error
		wantErr error
		wantMsg string
	}{
		"converged": {
			func(t *testing.T, version *int32) error {
				return waitForActiveVersion(t.Context(), testApplicationAPI{containers: func() ([]app.Placement, error) { return nil, nil }}, aid, version)
			}, nil, "",
		},
		"api error": {
			func(t *testing.T, version *int32) error {
				return waitForActiveVersion(t.Context(), testApplicationAPI{containers: func() ([]app.Placement, error) { return nil, apiError }}, aid, version)
			}, apiError, "",
		},
		"timed out between polls": {
			func(t *testing.T, version *int32) error {
				ctx, cancel := context.WithCancel(t.Context())
				api := testApplicationAPI{containers: func() ([]app.Placement, error) { cancel(); return stale, nil }}
				return waitForActiveVersion(ctx, api, aid, version)
			}, context.Canceled, "timed out waiting for version 2 to be placed",
		},
		"timed out before the first poll": {
			func(t *testing.T, version *int32) error {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				api := testApplicationAPI{containers: func() ([]app.Placement, error) {
					t.Fatal("Containers must not be called with a context that is done")
					return nil, nil
				}}
				return waitForActiveVersion(ctx, api, aid, version)
			}, context.Canceled, "timed out waiting for version 2 to be placed",
		},
	} {
		err := c.run(t, &two)

		if !errors.Is(err, c.wantErr) {
			t.Fatalf("%s: err = %v, want %v", name, err, c.wantErr)
		}
		if c.wantMsg != "" && !strings.Contains(err.Error(), c.wantMsg) {
			t.Fatalf("%s: err = %q, want it to contain %q", name, err, c.wantMsg)
		}
	}

	// nil waits for the containers to stop, with its own message
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := waitForActiveVersion(ctx, testApplicationAPI{}, aid, nil)

	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "timed out waiting for the containers to stop") {
		t.Fatalf("err = %v, want the drain timeout wrapping context.Canceled", err)
	}
}
