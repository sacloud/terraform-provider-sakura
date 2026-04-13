// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteAlertNotificationTargetDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_alert_notification_target.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertNotificationTargetDataSource_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_type", "simple_notification"),
					resource.TestCheckResourceAttr(resourceName, "url", "https://example.com/notify"),
					resource.TestCheckResourceAttr(resourceName, "description", "notification-target"),
					resource.TestCheckResourceAttrSet(resourceName, "config"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", "sakura_monitoring_suite_alert_notification_target.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraMonitoringSuiteAlertNotificationTargetDataSource_basic = `
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

data "sakura_monitoring_suite_alert_notification_target" "foobar" {
  id = sakura_monitoring_suite_alert_notification_target.foobar.id
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
}
`
