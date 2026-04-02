// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteMetricStorage_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_metric_storage.foobar"
	rand := test.RandomName()

	var storage monitoringsuiteapi.MetricsStorage
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteMetricStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricStorage_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "is_system", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.address"),
					resource.TestCheckResourceAttrSet(resourceName, "usage.metric_routings"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricStorage_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "is_system", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.address"),
					resource.TestCheckResourceAttrSet(resourceName, "usage.metric_routings"),
				),
			},
		},
	})
}

func TestAccImportSakuraMonitoringSuiteMetricStorage_basic(t *testing.T) {
	rand := test.RandomName()

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":        rand,
			"description": "description",
			"is_system":   "false",
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0],
			"created_at",
			"project_id",
			"resource_id",
			"endpoints.address",
			"usage.metric_routings",
		)
	}

	resourceName := "sakura_monitoring_suite_metric_storage.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraMonitoringSuiteMetricStorageDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricStorage_basic, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteMetricStorageDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewMetricsStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_metric_storage" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite metric storage: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteMetricStorageExists(n string, storage *monitoringsuiteapi.MetricsStorage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no metric storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewMetricsStorageOp(client.MonitoringSuiteClient)

		found, err := op.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if fmt.Sprintf("%d", found.ID) != rs.Primary.ID {
			return fmt.Errorf("not found metric storage: %s", rs.Primary.ID)
		}

		*storage = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteMetricStorage_basic = `
resource "sakura_monitoring_suite_metric_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}
`

var testAccSakuraMonitoringSuiteMetricStorage_update = `
resource "sakura_monitoring_suite_metric_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description-updated"
  is_system = false
}
`
