// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteMetricStorageDataSource_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	resourceName := "data.sakura_monitoring_suite_metric_storage.foobar"
	id := os.Getenv("SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricStorageDataSource_basic, id),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test"),
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
data "sakura_monitoring_suite_metric_storage" "foobar" {
  id = "{{ .arg0 }}"
}
`
