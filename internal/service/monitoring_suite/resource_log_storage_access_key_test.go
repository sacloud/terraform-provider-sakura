// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteLogStorageAccessKey_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_log_storage_access_key.foobar"
	rand := test.RandomName()

	var key monitoringsuiteapi.LogStorageAccessKey
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteLogStorageAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorageAccessKey_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_id", "sakura_monitoring_suite_log_storage.foobar", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorageAccessKey_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_id", "sakura_monitoring_suite_log_storage.foobar", "id"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteLogStorageAccessKeyDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewLogsStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_log_storage_access_key" {
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
			return fmt.Errorf("still exists monitoring suite log storage access key: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteLogStorageAccessKeyExists(n string, key *monitoringsuiteapi.LogStorageAccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no log storage access key ID is set")
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			return errors.New("no log storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewLogsStorageOp(client.MonitoringSuiteClient)

		found, err := op.ReadKey(context.Background(), storageID, uid)
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found log storage access key: %s", rs.Primary.ID)
		}

		*key = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteLogStorageAccessKey_basic = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  classification = "shared"
  is_system = false
}

resource "sakura_monitoring_suite_log_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_log_storage.foobar.id
  description = "access-key"
}
`

var testAccSakuraMonitoringSuiteLogStorageAccessKey_update = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  classification = "shared"
  is_system = false
}

resource "sakura_monitoring_suite_log_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_log_storage.foobar.id
  description = "access-key-updated"
}
`
