// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func TestVerResourceModelUpdateStatePreservesSecretByKey(t *testing.T) {
	model := verResourceModel{
		EnvVars: []envVarResourceModel{
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

// A value given via value_wo must never land in the state, whether or not the API conceals it.
func TestVerResourceModelUpdateStateKeepsWriteOnlyValueOutOfState(t *testing.T) {
	model := verResourceModel{
		EnvVars: []envVarResourceModel{
			// Value is seeded with garbage on purpose: value_wo_version must win over whatever is there.
			{Key: types.StringValue("SECRET"), Value: types.StringValue("stale"), ValueWOVersion: types.Int32Value(1), Secret: types.BoolValue(true)},
			{Key: types.StringValue("PLAIN"), ValueWOVersion: types.Int32Value(1), Secret: types.BoolValue(false)},
		},
	}

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
	if len(model.EnvVars) != 2 {
		t.Fatalf("EnvVars length = %d, want 2", len(model.EnvVars))
	}
	for _, e := range model.EnvVars {
		key := e.Key.ValueString()

		if !e.Value.IsNull() {
			t.Fatalf("EnvVars[%s].Value = %q, want null", key, e.Value.ValueString())
		}
		if !e.ValueWO.IsNull() {
			t.Fatalf("EnvVars[%s].ValueWO = %q, want null", key, e.ValueWO.ValueString())
		}
		if got := e.ValueWOVersion.ValueInt32(); got != 1 {
			t.Fatalf("EnvVars[%s].ValueWOVersion = %d, want 1", key, got)
		}
	}
}

// terraform import starts from an empty state: every env var comes from the API as is, with nothing write-only.
func TestVerResourceModelUpdateStateFillsEnvVarsOnImport(t *testing.T) {
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
	if len(model.EnvVars) != 2 {
		t.Fatalf("EnvVars length = %d, want 2", len(model.EnvVars))
	}
	for _, e := range model.EnvVars {
		key := e.Key.ValueString()

		if !e.ValueWO.IsNull() || !e.ValueWOVersion.IsNull() {
			t.Fatalf("EnvVars[%s] must not have write-only fields after import, got value_wo=%v value_wo_version=%v", key, e.ValueWO, e.ValueWOVersion)
		}
	}
	if !model.EnvVars[0].Value.IsNull() {
		t.Fatalf("EnvVars[0].Value = %q, want null", model.EnvVars[0].Value.ValueString())
	}
	if got := model.EnvVars[1].Value.ValueString(); got != "plain" {
		t.Fatalf("EnvVars[1].Value = %q, want %q", got, "plain")
	}
}

// Write-only values are absent from the plan; intoCreate has to pick them up from the config.
func TestVerResourceModelIntoCreateTakesWriteOnlyValueFromConfig(t *testing.T) {
	plan := verResourceModel{
		EnvVars: []envVarResourceModel{
			{Key: types.StringValue("SECRET"), ValueWOVersion: types.Int32Value(1), Secret: types.BoolValue(true)},
			{Key: types.StringValue("PLAIN"), Value: types.StringValue("plain"), Secret: types.BoolValue(false)},
		},
	}
	config := verResourceModel{
		EnvVars: []envVarResourceModel{
			{Key: types.StringValue("SECRET"), ValueWO: types.StringValue("s3cr3t"), ValueWOVersion: types.Int32Value(1), Secret: types.BoolValue(true)},
			{Key: types.StringValue("PLAIN"), Value: types.StringValue("plain"), Secret: types.BoolValue(false)},
		},
	}

	params, diagnostics := plan.intoCreate(&config)

	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if len(params.EnvVars) != 2 {
		t.Fatalf("EnvVars length = %d, want 2", len(params.EnvVars))
	}
	if got := params.EnvVars[0]; got.Key != "SECRET" || got.Value == nil || *got.Value != "s3cr3t" || !got.Secret {
		t.Fatalf("EnvVars[0] = {Key: %q, Value: %v, Secret: %t}, want {Key: \"SECRET\", Value: \"s3cr3t\", Secret: true}", got.Key, got.Value, got.Secret)
	}
	if got := params.EnvVars[1]; got.Key != "PLAIN" || got.Value == nil || *got.Value != "plain" || got.Secret {
		t.Fatalf("EnvVars[1] = {Key: %q, Value: %v, Secret: %t}, want {Key: \"PLAIN\", Value: \"plain\", Secret: false}", got.Key, got.Value, got.Secret)
	}
}

// A plan/config mismatch must surface as an error, never as a silently dropped secret.
func TestVerResourceModelIntoCreateRejectsInconsistentEnvVars(t *testing.T) {
	plan := verResourceModel{
		EnvVars: []envVarResourceModel{
			{Key: types.StringValue("SECRET"), ValueWOVersion: types.Int32Value(1), Secret: types.BoolValue(true)},
		},
	}

	for name, config := range map[string]verResourceModel{
		"fewer elements": {},
		"different key": {
			EnvVars: []envVarResourceModel{
				{Key: types.StringValue("OTHER"), ValueWO: types.StringValue("s3cr3t"), ValueWOVersion: types.Int32Value(1), Secret: types.BoolValue(true)},
			},
		},
	} {
		if _, diagnostics := plan.intoCreate(&config); !diagnostics.HasError() {
			t.Fatalf("%s: expected an error diagnostic, got none", name)
		}
	}
}

// The data source has no prior state to preserve secret values from; the API conceals them.
func TestVerDataSourceModelUpdateStateConcealsSecret(t *testing.T) {
	var model verDataSourceModel

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
	if len(model.EnvVars) != 2 {
		t.Fatalf("EnvVars length = %d, want 2", len(model.EnvVars))
	}
	if !model.EnvVars[0].Value.IsNull() {
		t.Fatalf("EnvVars[0].Value = %q, want null", model.EnvVars[0].Value.ValueString())
	}
	if got := model.EnvVars[1].Value.ValueString(); got != "plain" {
		t.Fatalf("EnvVars[1].Value = %q, want %q", got, "plain")
	}
}

// Write-only attributes come with structural rules (no Computed, not under a Set, ...) that the framework enforces.
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
