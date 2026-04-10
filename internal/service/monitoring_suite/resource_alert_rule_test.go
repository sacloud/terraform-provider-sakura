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

func TestAccSakuraMonitoringSuiteAlertRule_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	resourceName := "sakura_monitoring_suite_alert_rule.foobar"
	rand := test.RandomName()
	sId := os.Getenv("SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	var alertRule monitoringsuiteapi.AlertRule
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteAlertRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertRule_basic, rand, sId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertRuleExists(resourceName, &alertRule),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "open"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_id", "sakura_monitoring_suite_alert.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "metric_storage_id", sId),
					resource.TestCheckResourceAttr(resourceName, "query", "count_values"),
					resource.TestCheckResourceAttr(resourceName, "enabled_warning", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled_critical", "true"),
					resource.TestCheckResourceAttr(resourceName, "threshold_warning", ">=10"),
					resource.TestCheckResourceAttr(resourceName, "threshold_critical", ">=20"),
					resource.TestCheckResourceAttr(resourceName, "threshold_duration_warning", "600"),
					resource.TestCheckResourceAttr(resourceName, "threshold_duration_critical", "600"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertRule_update, rand, sId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertRuleExists(resourceName, &alertRule),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "open"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_id", "sakura_monitoring_suite_alert.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "metric_storage_id", sId),
					resource.TestCheckResourceAttr(resourceName, "query", "group"),
					resource.TestCheckResourceAttr(resourceName, "enabled_warning", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled_critical", "false"),
					resource.TestCheckResourceAttr(resourceName, "threshold_warning", ">=10"),
					resource.TestCheckResourceAttr(resourceName, "threshold_critical", ">=40"),
					resource.TestCheckResourceAttr(resourceName, "threshold_duration_warning", "600"),
					resource.TestCheckResourceAttr(resourceName, "threshold_duration_critical", "1200"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteAlertRuleDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewAlertRuleOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_alert_rule" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.Attributes["alert_id"], uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists monitoring suite alert rule: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteAlertRuleExists(n string, alertRule *monitoringsuiteapi.AlertRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no alert rule ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewAlertRuleOp(client.MonitoringSuiteClient)
		found, err := op.Read(context.Background(), rs.Primary.Attributes["alert_id"], uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found alert rule: %s", rs.Primary.ID)
		}

		*alertRule = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteAlertRule_basic = `
resource "sakura_monitoring_suite_alert" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_alert_rule" "foobar" {
  alert_id = sakura_monitoring_suite_alert.foobar.id
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
`

var testAccSakuraMonitoringSuiteAlertRule_update = `
resource "sakura_monitoring_suite_alert" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_alert_rule" "foobar" {
  alert_id = sakura_monitoring_suite_alert.foobar.id
  metric_storage_id = {{ .arg1 }}
  name = "{{ .arg0 }}-updated"
  query = "group"
  enabled_warning = true
  enabled_critical = false
  threshold_warning = ">=10"
  threshold_critical = ">=40"
  threshold_duration_warning = 600
  threshold_duration_critical = 1200
}
`
