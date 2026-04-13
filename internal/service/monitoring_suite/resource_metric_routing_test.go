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

func TestAccSakuraMonitoringSuiteMetricRouting_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	resourceName := "sakura_monitoring_suite_metric_routing.foobar"
	rand := test.RandomName()
	msId := os.Getenv("SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	var storage monitoringsuiteapi.MetricsRouting
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteMetricRoutingDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricRouting_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricRoutingExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "publisher_code", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "variant", "simplemq_metrics"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", "sakura_simple_mq.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_id", msId),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricRouting_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricRoutingExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "publisher_code", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "variant", "simplemq_metrics"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", "sakura_simple_mq.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_id", "sakura_monitoring_suite_metric_storage.foobar2", "id"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteMetricRouting_withoutResourceID(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	resourceName := "sakura_monitoring_suite_metric_routing.foobar"
	rand := test.RandomName()
	msId := os.Getenv("SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	var storage monitoringsuiteapi.MetricsRouting
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteMetricRoutingDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricRouting_basicWithoutResourceID, rand, msId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricRoutingExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "publisher_code", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "variant", "simplemq_metrics"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckNoResourceAttr(resourceName, "resource_id"),
					resource.TestCheckResourceAttr(resourceName, "storage_id", msId),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricRouting_updateWithoutResourceID, rand, msId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricRoutingExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "publisher_code", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "variant", "simplemq_metrics"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", "sakura_simple_mq.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_id", msId),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteMetricRoutingDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewMetricsRoutingOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_metric_routing" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists monitoring suite metric routing: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteMetricRoutingExists(n string, storage *monitoringsuiteapi.MetricsRouting) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no metric routing ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewMetricsRoutingOp(client.MonitoringSuiteClient)
		found, err := op.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found metric routing: %s", rs.Primary.ID)
		}

		*storage = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteMetricRouting_basic = `
resource "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_monitoring_suite_metric_storage" "foobar1" {
  id = "{{ .arg1 }}"
}

resource "sakura_monitoring_suite_metric_storage" "foobar2" {
  name = "{{ .arg0 }}-2"
  description = "description2"
}

resource "sakura_monitoring_suite_metric_routing" "foobar" {
  resource_id = sakura_simple_mq.foobar.id
  storage_id = data.sakura_monitoring_suite_metric_storage.foobar1.id
  publisher_code = "simplemq"
  variant = "simplemq_metrics"
}`

var testAccSakuraMonitoringSuiteMetricRouting_update = `
resource "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_monitoring_suite_metric_storage" "foobar1" {
  id = "{{ .arg1 }}"
}

resource "sakura_monitoring_suite_metric_storage" "foobar2" {
  name = "{{ .arg0 }}-2"
  description = "description2"
}

resource "sakura_monitoring_suite_metric_routing" "foobar" {
  resource_id = sakura_simple_mq.foobar.id
  storage_id = sakura_monitoring_suite_metric_storage.foobar2.id
  publisher_code = "simplemq"
  variant = "simplemq_metrics"
}`

var testAccSakuraMonitoringSuiteMetricRouting_basicWithoutResourceID = `
resource "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_monitoring_suite_metric_storage" "foobar" {
  id = "{{ .arg1 }}"
}

resource "sakura_monitoring_suite_metric_routing" "foobar" {
  storage_id = data.sakura_monitoring_suite_metric_storage.foobar.id
  publisher_code = "simplemq"
  variant = "simplemq_metrics"
}`

var testAccSakuraMonitoringSuiteMetricRouting_updateWithoutResourceID = `
resource "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_monitoring_suite_metric_storage" "foobar" {
  id = "{{ .arg1 }}"
}

resource "sakura_monitoring_suite_metric_routing" "foobar" {
  resource_id = sakura_simple_mq.foobar.id
  storage_id = data.sakura_monitoring_suite_metric_storage.foobar.id
  publisher_code = "simplemq"
  variant = "simplemq_metrics"
}`
