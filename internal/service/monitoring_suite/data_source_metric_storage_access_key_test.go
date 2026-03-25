// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteMetricStorageAccessKeyDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_metric_storage_access_key.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricStorageAccessKeyDataSource_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
				),
			},
		},
	})
}

var testAccSakuraMonitoringSuiteMetricStorageAccessKeyDataSource_basic = `
resource "sakura_monitoring_suite_metric_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}

resource "sakura_monitoring_suite_metric_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_metric_storage.foobar.id
  description = "access-key"
}

data "sakura_monitoring_suite_metric_storage_access_key" "foobar" {
  id = sakura_monitoring_suite_metric_storage_access_key.foobar.id
  storage_id = sakura_monitoring_suite_metric_storage.foobar.id
}
`
