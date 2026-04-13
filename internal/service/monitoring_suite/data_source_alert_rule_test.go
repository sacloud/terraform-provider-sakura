// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteAlertRuleDataSource_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	resourceName := "data.sakura_monitoring_suite_alert_rule.foobar"
	rand := test.RandomName()
	msId := os.Getenv("SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertRuleDataSource_basic, rand, msId),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "metric_storage_id", msId),
					resource.TestCheckResourceAttr(resourceName, "query", "count_values"),
					resource.TestCheckResourceAttr(resourceName, "enabled_warning", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled_critical", "true"),
					resource.TestCheckResourceAttr(resourceName, "threshold_warning", ">=10"),
					resource.TestCheckResourceAttr(resourceName, "threshold_critical", ">=20"),
					resource.TestCheckResourceAttr(resourceName, "threshold_duration_warning", "600"),
					resource.TestCheckResourceAttr(resourceName, "threshold_duration_critical", "600"),
					resource.TestCheckResourceAttrSet(resourceName, "open"),
					resource.TestCheckResourceAttrPair(resourceName, "id", "sakura_monitoring_suite_alert_rule.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraMonitoringSuiteAlertRuleDataSource_basic = `
resource "sakura_monitoring_suite_alert_project" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_alert_rule" "foobar" {
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
  metric_storage_id = {{ .arg1 }}
  name = "{{ .arg0 }}"
  query = "count_values"
  enabled_warning = true
  enabled_critical = true
  threshold_warning = ">=10"
  threshold_critical = ">=20"
  threshold_duration_warning = 600
  threshold_duration_critical = 600
}

data "sakura_monitoring_suite_alert_rule" "foobar" {
  id = sakura_monitoring_suite_alert_rule.foobar.id
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
}`
