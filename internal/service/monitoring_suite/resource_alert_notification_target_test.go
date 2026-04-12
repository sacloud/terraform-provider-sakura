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

func TestAccSakuraMonitoringSuiteAlertNotificationTarget_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_alert_notification_target.foobar"
	rand := test.RandomName()

	var target monitoringsuiteapi.NotificationTarget
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteAlertNotificationTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertNotificationTarget_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertNotificationTargetExists(resourceName, &target),
					resource.TestCheckResourceAttr(resourceName, "service_type", "simple_notification"),
					resource.TestCheckResourceAttr(resourceName, "url", "https://example.com/notify"),
					resource.TestCheckResourceAttr(resourceName, "description", "notification-target"),
					resource.TestCheckResourceAttrSet(resourceName, "config"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertNotificationTarget_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertNotificationTargetExists(resourceName, &target),
					resource.TestCheckResourceAttr(resourceName, "service_type", "simple_notification"),
					resource.TestCheckResourceAttr(resourceName, "url", "https://example.com/notify-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "notification-target-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "config"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteAlertNotificationTargetDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewNotificationTargetOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_alert_notification_target" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.Attributes["alert_project_id"], uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists monitoring suite alert notification target: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteAlertNotificationTargetExists(n string, target *monitoringsuiteapi.NotificationTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no alert notification target ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewNotificationTargetOp(client.MonitoringSuiteClient)
		found, err := op.Read(context.Background(), rs.Primary.Attributes["alert_project_id"], uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found alert notification target: %s", rs.Primary.ID)
		}

		*target = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteAlertNotificationTarget_basic = `
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
`

var testAccSakuraMonitoringSuiteAlertNotificationTarget_update = `
resource "sakura_monitoring_suite_alert_project" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_alert_notification_target" "foobar" {
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
  service_type = "simple_notification"
  url = "https://example.com/notify-updated"
  description = "notification-target-updated"
}
`
