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

func TestAccSakuraMonitoringSuiteLogRouting_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_LOG_STORAGE_ID")

	resourceName := "sakura_monitoring_suite_log_routing.foobar"
	rand := test.RandomName()
	lsId := os.Getenv("SAKURA_MONITORING_SUITE_LOG_STORAGE_ID")

	var storage monitoringsuiteapi.LogRouting
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteLogRoutingDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogRouting_basic, rand, lsId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogRoutingExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "publisher_code", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "variant", "simplemq_log"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", "sakura_simple_mq.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_id", lsId),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogRouting_update, rand, lsId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogRoutingExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "publisher_code", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "variant", "simplemq_log"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", "sakura_simple_mq.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_id", "sakura_monitoring_suite_log_storage.foobar2", "id"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteLogRouting_withoutResourceID(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_LOG_STORAGE_ID")

	resourceName := "sakura_monitoring_suite_log_routing.foobar"
	rand := test.RandomName()
	lsId := os.Getenv("SAKURA_MONITORING_SUITE_LOG_STORAGE_ID")

	var storage monitoringsuiteapi.LogRouting
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteLogRoutingDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogRouting_basicWithoutResourceID, rand, lsId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogRoutingExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "publisher_code", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "variant", "simplemq_log"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckNoResourceAttr(resourceName, "resource_id"),
					resource.TestCheckResourceAttr(resourceName, "storage_id", lsId),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogRouting_updateWithoutResourceID, rand, lsId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogRoutingExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "publisher_code", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "variant", "simplemq_log"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", "sakura_simple_mq.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_id", lsId),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteLogRoutingDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewLogRoutingOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_log_routing" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists monitoring suite log routing: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteLogRoutingExists(n string, storage *monitoringsuiteapi.LogRouting) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no log routing ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewLogRoutingOp(client.MonitoringSuiteClient)
		found, err := op.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found log routing: %s", rs.Primary.ID)
		}

		*storage = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteLogRouting_basic = `
resource "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_monitoring_suite_log_storage" "foobar1" {
  id = "{{ .arg1 }}"
}

resource "sakura_monitoring_suite_log_storage" "foobar2" {
  name = "{{ .arg0 }}-2"
  description = "description2"
}

resource "sakura_monitoring_suite_log_routing" "foobar" {
  resource_id = sakura_simple_mq.foobar.id
  storage_id = data.sakura_monitoring_suite_log_storage.foobar1.id
  publisher_code = "simplemq"
  variant = "simplemq_log"
}
`

var testAccSakuraMonitoringSuiteLogRouting_update = `
resource "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_monitoring_suite_log_storage" "foobar1" {
  id = "{{ .arg1 }}"
}

resource "sakura_monitoring_suite_log_storage" "foobar2" {
  name = "{{ .arg0 }}-2"
  description = "description2"
}

resource "sakura_monitoring_suite_log_routing" "foobar" {
  resource_id = sakura_simple_mq.foobar.id
  storage_id = sakura_monitoring_suite_log_storage.foobar2.id
  publisher_code = "simplemq"
  variant = "simplemq_log"
}
`

var testAccSakuraMonitoringSuiteLogRouting_basicWithoutResourceID = `
resource "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_monitoring_suite_log_storage" "foobar" {
  id = "{{ .arg1 }}"
}

resource "sakura_monitoring_suite_log_routing" "foobar" {
  storage_id = data.sakura_monitoring_suite_log_storage.foobar.id
  publisher_code = "simplemq"
  variant = "simplemq_log"
}`

var testAccSakuraMonitoringSuiteLogRouting_updateWithoutResourceID = `
resource "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_monitoring_suite_log_storage" "foobar" {
  id = "{{ .arg1 }}"
}

resource "sakura_monitoring_suite_log_routing" "foobar" {
  resource_id = sakura_simple_mq.foobar.id
  storage_id = data.sakura_monitoring_suite_log_storage.foobar.id
  publisher_code = "simplemq"
  variant = "simplemq_log"
}`
