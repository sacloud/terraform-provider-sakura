// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteTraceStorageAccessKey_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_TRACE_STORAGE_ID")

	resourceName := "sakura_monitoring_suite_trace_storage_access_key.foobar"
	id := os.Getenv("SAKURA_MONITORING_SUITE_TRACE_STORAGE_ID")

	var key monitoringsuiteapi.TraceStorageAccessKey
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteTraceStorageAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorageAccessKey_basic, id),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteTraceStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_id", "data.sakura_monitoring_suite_trace_storage.foobar", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorageAccessKey_update, id),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteTraceStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_id", "data.sakura_monitoring_suite_trace_storage.foobar", "id"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteTraceStorageAccessKeyDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewTracesStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_trace_storage_access_key" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			continue
		}

		_, err = op.ReadKey(context.Background(), storageID, uid)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite trace storage access key: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteTraceStorageAccessKeyExists(n string, key *monitoringsuiteapi.TraceStorageAccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no trace storage access key ID is set")
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			return errors.New("no trace storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewTracesStorageOp(client.MonitoringSuiteClient)

		found, err := op.ReadKey(context.Background(), storageID, uid)
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found trace storage access key: %s", rs.Primary.ID)
		}

		*key = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteTraceStorageAccessKey_basic = `
data "sakura_monitoring_suite_trace_storage" "foobar" {
  id = "{{ .arg0 }}"
}

resource "sakura_monitoring_suite_trace_storage_access_key" "foobar" {
  storage_id = data.sakura_monitoring_suite_trace_storage.foobar.id
  description = "access-key"
}
`

var testAccSakuraMonitoringSuiteTraceStorageAccessKey_update = `
data "sakura_monitoring_suite_trace_storage" "foobar" {
  id = "{{ .arg0 }}"
}

resource "sakura_monitoring_suite_trace_storage_access_key" "foobar" {
  storage_id = data.sakura_monitoring_suite_trace_storage.foobar.id
  description = "access-key-updated"
}
`
