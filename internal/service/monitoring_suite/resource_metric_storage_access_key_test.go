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

func TestAccSakuraMonitoringSuiteMetricStorageAccessKey_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_metric_storage_access_key.foobar"
	rand := test.RandomName()

	var key monitoringsuiteapi.MetricsStorageAccessKey
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteMetricStorageAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricStorageAccessKey_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_id", "sakura_monitoring_suite_metric_storage.foobar", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricStorageAccessKey_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_id", "sakura_monitoring_suite_metric_storage.foobar", "id"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteMetricStorageAccessKeyDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewMetricsStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_metric_storage_access_key" {
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
			return fmt.Errorf("still exists monitoring suite metric storage access key: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteMetricStorageAccessKeyExists(n string, key *monitoringsuiteapi.MetricsStorageAccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no metric storage access key ID is set")
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			return errors.New("no metric storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewMetricsStorageOp(client.MonitoringSuiteClient)

		found, err := op.ReadKey(context.Background(), storageID, uid)
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found metric storage access key: %s", rs.Primary.ID)
		}

		*key = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteMetricStorageAccessKey_basic = `
resource "sakura_monitoring_suite_metric_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}

resource "sakura_monitoring_suite_metric_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_metric_storage.foobar.id
  description = "access-key"
}
`

var testAccSakuraMonitoringSuiteMetricStorageAccessKey_update = `
resource "sakura_monitoring_suite_metric_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}

resource "sakura_monitoring_suite_metric_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_metric_storage.foobar.id
  description = "access-key-updated"
}
`
