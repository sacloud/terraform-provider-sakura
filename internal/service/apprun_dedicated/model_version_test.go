// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/apprun-dedicated-api-go/apis/version"
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
