// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteAlertNotificationRoutingDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_alert_notification_routing.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertNotificationRoutingDataSource_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resend_interval_minutes", "10"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.0.name", "name1"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.0.value", "value1"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "notification_target_id", "sakura_monitoring_suite_alert_notification_target.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "order"),
					resource.TestCheckResourceAttrPair(resourceName, "id", "sakura_monitoring_suite_alert_notification_routing.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraMonitoringSuiteAlertNotificationRoutingDataSource_basic = `
resource "sakura_monitoring_suite_alert_project" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_alert_notification_target" "foobar" {
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
  service_type = "simple_notification"
  url = "https://example.com/notify"
  description = "notification-target"
}

resource "sakura_monitoring_suite_alert_notification_routing" "foobar" {
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
  notification_target_id = sakura_monitoring_suite_alert_notification_target.foobar.id
  resend_interval_minutes = 10
  match_labels = [
    {
      name = "name1"
      value = "value1"
    }
  ]
}

data "sakura_monitoring_suite_alert_notification_routing" "foobar" {
  id = sakura_monitoring_suite_alert_notification_routing.foobar.id
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
}
`
