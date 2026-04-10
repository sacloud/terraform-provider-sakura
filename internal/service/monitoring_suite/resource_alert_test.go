// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteAlert_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_alert.foobar"
	rand := test.RandomName()

	var alertProject monitoringsuiteapi.AlertProject
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteAlertDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlert_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertExists(resourceName, &alertProject),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlert_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertExists(resourceName, &alertProject),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteAlertDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewAlertProjectOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_alert" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite alert: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteAlertExists(n string, alertProject *monitoringsuiteapi.AlertProject) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no alert project ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewAlertProjectOp(client.MonitoringSuiteClient)
		found, err := op.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if strconv.FormatInt(found.ID, 10) != rs.Primary.ID {
			return fmt.Errorf("not found alert project: %s", rs.Primary.ID)
		}

		*alertProject = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteAlert_basic = `
resource "sakura_monitoring_suite_alert" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}
`

var testAccSakuraMonitoringSuiteAlert_update = `
resource "sakura_monitoring_suite_alert" "foobar" {
  name = "{{ .arg0 }}-updated"
  description = "description-updated"
}
`
