// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"testing"

	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
)

func TestFilterLogStorageByNameAndTags(t *testing.T) {
	storages := []monitoringsuiteapi.LogStorage{
		{
			ID:   1,
			Name: monitoringsuiteapi.NewOptString("alpha"),
			Tags: []string{"tag1", "tag2"},
		},
		{
			ID:   2,
			Name: monitoringsuiteapi.NewOptString("beta"),
			Tags: []string{"tag3"},
		},
	}

	cases := []struct {
		name    string
		query   string
		tags    []string
		wantID  int64
		wantErr bool
	}{
		{
			name:   "match by name",
			query:  "alpha",
			wantID: 1,
		},
		{
			name:   "match by tags",
			query:  "",
			tags:   []string{"tag2"},
			wantID: 1,
		},
		{
			name:   "match by name and tags",
			query:  "alpha",
			tags:   []string{"tag1"},
			wantID: 1,
		},
		{
			name:    "not found",
			query:   "gamma",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := filterLogStorageByName(storages, tc.query)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected storage but got nil")
			}
			if got.ID != tc.wantID {
				t.Fatalf("got ID %d, want %d", got.ID, tc.wantID)
			}
		})
	}

	dup := append([]monitoringsuiteapi.LogStorage{}, storages...)
	dup = append(dup, monitoringsuiteapi.LogStorage{ID: 3, Name: monitoringsuiteapi.NewOptString("alpha"), Tags: []string{"tag1"}})
	if _, err := filterLogStorageByName(dup, "alpha"); err == nil {
		t.Fatalf("expected error for duplicate matches")
	}
}

func TestFilterMetricsStorageByNameAndTags(t *testing.T) {
	storages := []monitoringsuiteapi.MetricsStorage{
		{
			ID:   1,
			Name: monitoringsuiteapi.NewOptString("alpha"),
			Tags: []string{"tag1", "tag2"},
		},
		{
			ID:   2,
			Name: monitoringsuiteapi.NewOptString("beta"),
			Tags: []string{"tag3"},
		},
	}

	cases := []struct {
		name    string
		query   string
		tags    []string
		wantID  int64
		wantErr bool
	}{
		{
			name:   "match by name",
			query:  "alpha",
			wantID: 1,
		},
		{
			name:   "match by tags",
			query:  "",
			tags:   []string{"tag2"},
			wantID: 1,
		},
		{
			name:   "match by name and tags",
			query:  "alpha",
			tags:   []string{"tag1"},
			wantID: 1,
		},
		{
			name:    "not found",
			query:   "gamma",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := filterMetricsStorageByName(storages, tc.query)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected storage but got nil")
			}
			if got.ID != tc.wantID {
				t.Fatalf("got ID %d, want %d", got.ID, tc.wantID)
			}
		})
	}

	dup := append([]monitoringsuiteapi.MetricsStorage{}, storages...)
	dup = append(dup, monitoringsuiteapi.MetricsStorage{ID: 3, Name: monitoringsuiteapi.NewOptString("alpha"), Tags: []string{"tag1"}})
	if _, err := filterMetricsStorageByName(dup, "alpha"); err == nil {
		t.Fatalf("expected error for duplicate matches")
	}
}

func TestFilterTraceStorageByNameAndTags(t *testing.T) {
	storages := []monitoringsuiteapi.TraceStorage{
		{
			ID:   1,
			Name: monitoringsuiteapi.NewOptString("alpha"),
			Tags: []string{"tag1", "tag2"},
		},
		{
			ID:   2,
			Name: monitoringsuiteapi.NewOptString("beta"),
			Tags: []string{"tag3"},
		},
	}

	cases := []struct {
		name    string
		query   string
		tags    []string
		wantID  int64
		wantErr bool
	}{
		{
			name:   "match by name",
			query:  "alpha",
			wantID: 1,
		},
		{
			name:   "match by tags",
			query:  "",
			tags:   []string{"tag2"},
			wantID: 1,
		},
		{
			name:   "match by name and tags",
			query:  "alpha",
			tags:   []string{"tag1"},
			wantID: 1,
		},
		{
			name:    "not found",
			query:   "gamma",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := filterTraceStorageByName(storages, tc.query)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected storage but got nil")
			}
			if got.ID != tc.wantID {
				t.Fatalf("got ID %d, want %d", got.ID, tc.wantID)
			}
		})
	}

	dup := append([]monitoringsuiteapi.TraceStorage{}, storages...)
	dup = append(dup, monitoringsuiteapi.TraceStorage{ID: 3, Name: monitoringsuiteapi.NewOptString("alpha"), Tags: []string{"tag1"}})
	if _, err := filterTraceStorageByName(dup, "alpha"); err == nil {
		t.Fatalf("expected error for duplicate matches")
	}
}
