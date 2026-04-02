// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteMetricStorageDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_metric_storage.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricStorageDataSource_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
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
		},
	})
}

var testAccSakuraMonitoringSuiteMetricStorageDataSource_basic = `
resource "sakura_monitoring_suite_metric_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}

data "sakura_monitoring_suite_metric_storage" "foobar" {
  id = sakura_monitoring_suite_metric_storage.foobar.id
}
`
