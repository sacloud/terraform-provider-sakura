// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/version"
)

// Regression: reading a port whose health_check is not configured panicked because
// the API returns ExposedPort.HealthCheck=nil and updateState dereferenced it.
func TestExposedPortModelUpdateStateNilHealthCheck(t *testing.T) {
	var p exposedPortModel

	p.updateState(version.ExposedPort{TargetPort: v1.Port(80), HealthCheck: nil}) // must not panic

	if p.HealthCheck != nil {
		t.Fatalf("HealthCheck should stay null when the API returns nil, got %+v", p.HealthCheck)
	}
}

// A configured health_check is mapped into every field on read.
func TestExposedPortModelUpdateStateWithHealthCheck(t *testing.T) {
	var p exposedPortModel

	p.updateState(version.ExposedPort{
		TargetPort: v1.Port(80),
		HealthCheck: &v1.HealthCheck{
			Path:            "/healthz",
			IntervalSeconds: 10,
			TimeoutSeconds:  5,
		},
	})

	if p.HealthCheck == nil {
		t.Fatal("HealthCheck should be populated when the API returns it")
	}
	if got := p.HealthCheck.Path.ValueString(); got != "/healthz" {
		t.Fatalf("Path = %q, want %q", got, "/healthz")
	}
	if got := p.HealthCheck.IntervalSeconds.ValueInt32(); got != 10 {
		t.Fatalf("IntervalSeconds = %d, want 10", got)
	}
	if got := p.HealthCheck.TimeoutSeconds.ValueInt32(); got != 5 {
		t.Fatalf("TimeoutSeconds = %d, want 5", got)
	}
}

// health_check is optional, so an omitted block must not panic on create.
func TestExposedPortModelIntoCreateNilHealthCheck(t *testing.T) {
	p := exposedPortModel{TargetPort: types.Int32Value(80)}

	got, diags := p.intoCreate()
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got.HealthCheck != nil {
		t.Fatalf("HealthCheck should be nil when omitted, got %+v", got.HealthCheck)
	}
}

func TestVerModelUpdateStatePreservesSecretByKey(t *testing.T) {
	model := verModel{
		EnvVars: []envVarModel{
			{Key: types.StringValue("ENV_VAR2"), Value: types.StringValue("value2"), Secret: types.BoolValue(true)},
			{Key: types.StringValue("ENV_VAR1"), Value: types.StringValue("value1"), Secret: types.BoolValue(false)},
		},
	}

	detail := version.VersionDetail{
		EnvVars: []version.EnvironmentVariable{
			{Key: "ENV_VAR1", Value: types.StringValue("value1").ValueStringPointer(), Secret: false},
			{Key: "ENV_VAR2", Value: nil, Secret: true},
		},
	}

	var aid v1.ApplicationID

	diagnostics := model.updateState(t.Context(), &detail, aid)

	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if len(model.EnvVars) != 2 {
		t.Fatalf("EnvVars length = %d, want 2", len(model.EnvVars))
	}
	if got := model.EnvVars[0].Key.ValueString(); got != "ENV_VAR2" {
		t.Fatalf("EnvVars[0].Key = %q, want %q", got, "ENV_VAR2")
	}
	if got := model.EnvVars[0].Value.ValueString(); got != "value2" {
		t.Fatalf("EnvVars[0].Value = %q, want %q", got, "value2")
	}
	if got := model.EnvVars[1].Key.ValueString(); got != "ENV_VAR1" {
		t.Fatalf("EnvVars[1].Key = %q, want %q", got, "ENV_VAR1")
	}
	if got := model.EnvVars[1].Value.ValueString(); got != "value1" {
		t.Fatalf("EnvVars[1].Value = %q, want %q", got, "value1")
	}
}

func TestVerModelUpdateStateEscapesUUIDInID(t *testing.T) {
	model := verModel{}

	aid := v1.ApplicationID(uuid.MustParse("12345678-1234-1234-1234-123456789abc"))
	detail := version.VersionDetail{Version: 42}

	diag := model.updateState(t.Context(), &detail, aid)
	if diag.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diag)
	}

	expected := "12345678-1234-1234-1234-123456789abc/42"
	if actual := model.ID.ValueString(); actual != expected {
		t.Fatalf("ID = %q, want %q", actual, expected)
	}
}

// import: secrets go to secret_vars, the rest to env_vars
func TestVerResourceModelUpdateStateRoutesSecretsOnImport(t *testing.T) {
	var model verResourceModel

	detail := version.VersionDetail{
		EnvVars: []version.EnvironmentVariable{
			{Key: "SECRET", Value: nil, Secret: true},
			{Key: "PLAIN", Value: types.StringValue("plain").ValueStringPointer(), Secret: false},
		},
	}

	var aid v1.ApplicationID

	if diagnostics := model.updateState(t.Context(), &detail, aid); diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if len(model.SecretVars) != 1 || len(model.EnvVars) != 1 {
		t.Fatalf("SecretVars length = %d, EnvVars length = %d, want 1 and 1", len(model.SecretVars), len(model.EnvVars))
	}
	if got := model.SecretVars[0]; got.Key.ValueString() != "SECRET" || !got.ValueWO.IsNull() || !got.ValueWOVersion.IsNull() {
		t.Fatalf("SecretVars[0] = %+v, want {Key: \"SECRET\", ValueWO: null, ValueWOVersion: null}", got)
	}
	if got := model.EnvVars[0]; got.Key.ValueString() != "PLAIN" || got.Value.ValueString() != "plain" || got.Secret.ValueBool() {
		t.Fatalf("EnvVars[0] = %+v, want {Key: \"PLAIN\", Value: \"plain\", Secret: false}", got)
	}
}

// deprecated secret = true stays in env_vars
func TestVerResourceModelUpdateStateKeepsSecretVarsAndLegacyApart(t *testing.T) {
	model := verResourceModel{
		verModel: verModel{
			EnvVars: []envVarModel{
				{Key: types.StringValue("LEGACY"), Value: types.StringValue("legacy"), Secret: types.BoolValue(true)},
			},
		},
		SecretVars: []secretVarModel{
			{Key: types.StringValue("NEW"), ValueWOVersion: types.Int32Value(1)},
		},
	}

	detail := version.VersionDetail{
		EnvVars: []version.EnvironmentVariable{
			{Key: "NEW", Value: nil, Secret: true},
			{Key: "LEGACY", Value: nil, Secret: true},
		},
	}

	var aid v1.ApplicationID

	if diagnostics := model.updateState(t.Context(), &detail, aid); diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if len(model.SecretVars) != 1 || len(model.EnvVars) != 1 {
		t.Fatalf("SecretVars length = %d, EnvVars length = %d, want 1 and 1", len(model.SecretVars), len(model.EnvVars))
	}
	if got := model.SecretVars[0]; got.Key.ValueString() != "NEW" || !got.ValueWO.IsNull() || got.ValueWOVersion.ValueInt32() != 1 {
		t.Fatalf("SecretVars[0] = %+v, want {Key: \"NEW\", ValueWO: null, ValueWOVersion: 1}", got)
	}
	if got := model.EnvVars[0]; got.Key.ValueString() != "LEGACY" || got.Value.ValueString() != "legacy" || !got.Secret.ValueBool() {
		t.Fatalf("EnvVars[0] = %+v, want {Key: \"LEGACY\", Value: \"legacy\", Secret: true}", got)
	}
}

// write-only values come from the config, not the plan
func TestVerResourceModelIntoCreateMergesSecretVars(t *testing.T) {
	plan := verResourceModel{
		verModel: verModel{
			EnvVars: []envVarModel{
				{Key: types.StringValue("PLAIN"), Value: types.StringValue("plain"), Secret: types.BoolValue(false)},
			},
		},
		SecretVars: []secretVarModel{
			{Key: types.StringValue("SECRET"), ValueWOVersion: types.Int32Value(1)},
			{Key: types.StringValue("RETAINED")},
		},
	}
	config := verResourceModel{
		verModel: plan.verModel,
		SecretVars: []secretVarModel{
			{Key: types.StringValue("SECRET"), ValueWO: types.StringValue("s3cr3t"), ValueWOVersion: types.Int32Value(1)},
			{Key: types.StringValue("RETAINED")},
		},
	}

	params, diagnostics := plan.intoCreate(&config)

	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if len(params.EnvVars) != 3 {
		t.Fatalf("EnvVars length = %d, want 3", len(params.EnvVars))
	}
	if got := params.EnvVars[0]; got.Key != "PLAIN" || got.Value == nil || *got.Value != "plain" || got.Secret {
		t.Fatalf("EnvVars[0] = {Key: %q, Value: %v, Secret: %t}, want {Key: \"PLAIN\", Value: \"plain\", Secret: false}", got.Key, got.Value, got.Secret)
	}
	if got := params.EnvVars[1]; got.Key != "SECRET" || got.Value == nil || *got.Value != "s3cr3t" || !got.Secret {
		t.Fatalf("EnvVars[1] = {Key: %q, Value: %v, Secret: %t}, want {Key: \"SECRET\", Value: \"s3cr3t\", Secret: true}", got.Key, got.Value, got.Secret)
	}
	if got := params.EnvVars[2]; got.Key != "RETAINED" || got.Value != nil || !got.Secret {
		t.Fatalf("EnvVars[2] = {Key: %q, Value: %v, Secret: %t}, want {Key: \"RETAINED\", Value: nil, Secret: true}", got.Key, got.Value, got.Secret)
	}
}

// plan/config mismatch is an error, not a silently dropped secret
func TestVerResourceModelIntoCreateRejectsInconsistentSecretVars(t *testing.T) {
	plan := verResourceModel{
		SecretVars: []secretVarModel{
			{Key: types.StringValue("SECRET"), ValueWOVersion: types.Int32Value(1)},
		},
	}

	for name, config := range map[string]verResourceModel{
		"fewer elements": {},
		"different key": {
			SecretVars: []secretVarModel{
				{Key: types.StringValue("OTHER"), ValueWO: types.StringValue("s3cr3t"), ValueWOVersion: types.Int32Value(1)},
			},
		},
	} {
		if _, diagnostics := plan.intoCreate(&config); !diagnostics.HasError() {
			t.Fatalf("%s: expected an error diagnostic, got none", name)
		}
	}
}

// write-only attributes have structural rules the framework checks only at runtime
func TestVerResourceSchemaValidateImplementation(t *testing.T) {
	var res resource.SchemaResponse

	NewVersionResource().Schema(t.Context(), resource.SchemaRequest{}, &res)

	if res.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", res.Diagnostics)
	}
	if diagnostics := res.Schema.ValidateImplementation(t.Context()); diagnostics.HasError() {
		t.Fatalf("invalid schema: %v", diagnostics)
	}
}

// no key in both lists, at most 50 in total; unknown values must not error
func TestVerResourceValidateConfig(t *testing.T) {
	ctx := t.Context()

	var sch resource.SchemaResponse
	NewVersionResource().Schema(ctx, resource.SchemaRequest{}, &sch)

	envVarsType := sch.Schema.Attributes["env_vars"].GetType().TerraformType(ctx).(tftypes.List)
	secretVarsType := sch.Schema.Attributes["secret_vars"].GetType().TerraformType(ctx).(tftypes.List)

	envVar := func(key any) tftypes.Value {
		return tftypes.NewValue(envVarsType.ElementType, map[string]tftypes.Value{
			"key":    tftypes.NewValue(tftypes.String, key),
			"value":  tftypes.NewValue(tftypes.String, "plain"),
			"secret": tftypes.NewValue(tftypes.Bool, nil),
		})
	}
	secretVar := func(key any) tftypes.Value {
		return tftypes.NewValue(secretVarsType.ElementType, map[string]tftypes.Value{
			"key":              tftypes.NewValue(tftypes.String, key),
			"value_wo":         tftypes.NewValue(tftypes.String, "s3cr3t"),
			"value_wo_version": tftypes.NewValue(tftypes.Number, big.NewFloat(1)),
		})
	}
	many := func(n int, prefix string, elem func(key any) tftypes.Value) []tftypes.Value {
		ret := make([]tftypes.Value, 0, n)
		for i := range n {
			ret = append(ret, elem(fmt.Sprintf("%s%d", prefix, i)))
		}
		return ret
	}
	configWith := func(envVars, secretVars any) tfsdk.Config {
		attrs := make(map[string]tftypes.Value, len(sch.Schema.Attributes))
		for name, a := range sch.Schema.Attributes {
			attrs[name] = tftypes.NewValue(a.GetType().TerraformType(ctx), nil)
		}
		attrs["env_vars"] = tftypes.NewValue(envVarsType, envVars)
		attrs["secret_vars"] = tftypes.NewValue(secretVarsType, secretVars)

		return tfsdk.Config{Schema: sch.Schema, Raw: tftypes.NewValue(sch.Schema.Type().TerraformType(ctx), attrs)}
	}

	for name, c := range map[string]struct {
		envVars, secretVars any
		wantError           bool
	}{
		"distinct keys":    {[]tftypes.Value{envVar("PLAIN")}, []tftypes.Value{secretVar("SECRET")}, false},
		"duplicate key":    {[]tftypes.Value{envVar("DUP")}, []tftypes.Value{secretVar("DUP")}, true},
		"50 in total":      {many(25, "E", envVar), many(25, "S", secretVar), false},
		"51 in total":      {many(26, "E", envVar), many(25, "S", secretVar), true},
		"env_vars unknown": {tftypes.UnknownValue, []tftypes.Value{secretVar("DUP")}, false},
		"secret_vars null": {[]tftypes.Value{envVar("DUP")}, nil, false},
		"unknown element":  {[]tftypes.Value{envVar("DUP")}, []tftypes.Value{tftypes.NewValue(secretVarsType.ElementType, tftypes.UnknownValue)}, false},
		"unknown key":      {[]tftypes.Value{envVar("DUP")}, []tftypes.Value{secretVar(tftypes.UnknownValue)}, false},
	} {
		var res resource.ValidateConfigResponse

		NewVersionResource().(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: configWith(c.envVars, c.secretVars)}, &res)

		if res.Diagnostics.HasError() != c.wantError {
			t.Fatalf("%s: HasError() = %t, want %t: %v", name, !c.wantError, c.wantError, res.Diagnostics)
		}
	}
}
