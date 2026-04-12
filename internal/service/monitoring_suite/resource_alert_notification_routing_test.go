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

func TestAccSakuraMonitoringSuiteAlertNotificationRouting_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_alert_notification_routing.foobar"
	rand := test.RandomName()

	var routing monitoringsuiteapi.NotificationRouting
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteAlertNotificationRoutingDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertNotificationRouting_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertNotificationRoutingExists(resourceName, &routing),
					resource.TestCheckResourceAttr(resourceName, "resend_interval_minutes", "10"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.0.name", "name1"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.0.value", "value1"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "notification_target_id", "sakura_monitoring_suite_alert_notification_target.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "order"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertNotificationRouting_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertNotificationRoutingExists(resourceName, &routing),
					resource.TestCheckResourceAttr(resourceName, "resend_interval_minutes", "20"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.0.name", "name1"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.0.value", "value1"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.1.name", "name2"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.1.value", "value2"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "notification_target_id", "sakura_monitoring_suite_alert_notification_target.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "order"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteAlertNotificationRoutingDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewNotificationRoutingOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_alert_notification_routing" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.Attributes["alert_project_id"], uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists monitoring suite alert notification routing: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteAlertNotificationRoutingExists(n string, routing *monitoringsuiteapi.NotificationRouting) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no alert notification routing ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewNotificationRoutingOp(client.MonitoringSuiteClient)
		found, err := op.Read(context.Background(), rs.Primary.Attributes["alert_project_id"], uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found alert notification routing: %s", rs.Primary.ID)
		}

		*routing = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteAlertNotificationRouting_basic = `
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
`

var testAccSakuraMonitoringSuiteAlertNotificationRouting_update = `
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
  resend_interval_minutes = 20
  match_labels = [
    {
      name = "name1"
      value = "value1"
    },
    {
      name = "name2"
      value = "value2"
    }
  ]
}
`
