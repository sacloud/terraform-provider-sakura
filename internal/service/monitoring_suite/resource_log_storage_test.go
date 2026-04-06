// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteLogStorage_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_log_storage.foobar"
	rand := test.RandomName()

	var storage monitoringsuiteapi.LogStorage
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteLogStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorage_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "classification", "shared"),
					resource.TestCheckResourceAttr(resourceName, "is_system", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "expire_day"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.ingester.address"),
					resource.TestCheckResourceAttrSet(resourceName, "usage.log_routings"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorage_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "classification", "shared"),
					resource.TestCheckResourceAttr(resourceName, "is_system", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "expire_day"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.ingester.address"),
					resource.TestCheckResourceAttrSet(resourceName, "usage.log_routings"),
				),
			},
		},
	})
}

func TestAccImportSakuraMonitoringSuiteLogStorage_basic(t *testing.T) {
	rand := test.RandomName()

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":           rand,
			"description":    "description",
			"classification": "shared",
			"is_system":      "false",
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0],
			"created_at",
			"project_id",
			"resource_id",
			"expire_day",
			"endpoints.ingester.address",
			"usage.log_routings",
		)
	}

	resourceName := "sakura_monitoring_suite_log_storage.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraMonitoringSuiteLogStorageDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorage_basic, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteLogStorageDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewLogsStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_log_storage" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite log storage: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteLogStorageExists(n string, storage *monitoringsuiteapi.LogStorage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no log storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewLogsStorageOp(client.MonitoringSuiteClient)

		found, err := op.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if fmt.Sprintf("%d", found.ID) != rs.Primary.ID {
			return fmt.Errorf("not found log storage: %s", rs.Primary.ID)
		}

		*storage = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteLogStorage_basic = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  classification = "shared"
  is_system = false
}
`

var testAccSakuraMonitoringSuiteLogStorage_update = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description-updated"
  classification = "shared"
  is_system = false
}
`
