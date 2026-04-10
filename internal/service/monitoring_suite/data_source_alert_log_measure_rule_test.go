// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteAlertLogMeasureRuleDataSource_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID", "SAKURA_MONITORING_SUITE_LOG_STORAGE_ID")

	resourceName := "data.sakura_monitoring_suite_alert_log_measure_rule.foobar"
	name := strings.ToLower(test.RandomName())
	lsId := os.Getenv("SAKURA_MONITORING_SUITE_LOG_STORAGE_ID")
	msId := os.Getenv("SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogMeasureRuleDataSource_basic, name, lsId, msId),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_id", "sakura_monitoring_suite_alert.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "log_storage_id", lsId),
					resource.TestCheckResourceAttr(resourceName, "metric_storage_id", msId),
					resource.TestCheckResourceAttr(resourceName, "rule.version", "v1"),
					resource.TestCheckResourceAttrSet(resourceName, "rule.query.matchers"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "id", "sakura_monitoring_suite_alert_log_measure_rule.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraMonitoringSuiteLogMeasureRuleDataSource_basic = `
resource "sakura_monitoring_suite_alert" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_alert_log_measure_rule" "foobar" {
  alert_id = sakura_monitoring_suite_alert.foobar.id
  log_storage_id = {{ .arg1 }}
  metric_storage_id = {{ .arg2 }}
  name = "{{ .arg0 }}"
  description = "description"
  rule = {
    version = "v1"
    query = {
      matchers = jsonencode([
        {
          type = "string"
          operator = "eq"
          field = "service_name"
          value = "example"
          value_list = ["example"]
        }
      ])
    }
  }
}

data "sakura_monitoring_suite_alert_log_measure_rule" "foobar" {
  id = sakura_monitoring_suite_alert_log_measure_rule.foobar.id
  alert_id = sakura_monitoring_suite_alert.foobar.id
}`
