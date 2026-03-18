// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteLogStorageDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_log_storage.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorageDataSource_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteMetricsStorageDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_metrics_storage.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricsStorageDataSource_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteTraceStorageDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_trace_storage.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorageDataSource_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteLogStorageAccessKeyDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_log_storage_access_key.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorageAccessKeyDataSource_basic, rand),
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

func TestAccSakuraMonitoringSuiteMetricsStorageAccessKeyDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_metrics_storage_access_key.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricsStorageAccessKeyDataSource_basic, rand),
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

func TestAccSakuraMonitoringSuiteTraceStorageAccessKeyDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_monitoring_suite_trace_storage_access_key.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorageAccessKeyDataSource_basic, rand),
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

var testAccSakuraMonitoringSuiteLogStorageDataSource_basic = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  classification = "shared"
  is_system = false
}

data "sakura_monitoring_suite_log_storage" "foobar" {
  id = sakura_monitoring_suite_log_storage.foobar.id

  depends_on = [sakura_monitoring_suite_log_storage.foobar]
}
`

var testAccSakuraMonitoringSuiteMetricsStorageDataSource_basic = `
resource "sakura_monitoring_suite_metrics_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}

data "sakura_monitoring_suite_metrics_storage" "foobar" {
  id = sakura_monitoring_suite_metrics_storage.foobar.id

  depends_on = [sakura_monitoring_suite_metrics_storage.foobar]
}
`

var testAccSakuraMonitoringSuiteTraceStorageDataSource_basic = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

data "sakura_monitoring_suite_trace_storage" "foobar" {
  id = sakura_monitoring_suite_trace_storage.foobar.id

  depends_on = [sakura_monitoring_suite_trace_storage.foobar]
}
`

var testAccSakuraMonitoringSuiteLogStorageAccessKeyDataSource_basic = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  classification = "shared"
  is_system = false
}

resource "sakura_monitoring_suite_log_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_log_storage.foobar.id
  description = "access-key"
}

data "sakura_monitoring_suite_log_storage_access_key" "foobar" {
  id = sakura_monitoring_suite_log_storage_access_key.foobar.id
  storage_id = sakura_monitoring_suite_log_storage.foobar.id

  depends_on = [sakura_monitoring_suite_log_storage_access_key.foobar]
}
`

var testAccSakuraMonitoringSuiteMetricsStorageAccessKeyDataSource_basic = `
resource "sakura_monitoring_suite_metrics_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}

resource "sakura_monitoring_suite_metrics_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_metrics_storage.foobar.id
  description = "access-key"
}

data "sakura_monitoring_suite_metrics_storage_access_key" "foobar" {
  id = sakura_monitoring_suite_metrics_storage_access_key.foobar.id
  storage_id = sakura_monitoring_suite_metrics_storage.foobar.id

  depends_on = [sakura_monitoring_suite_metrics_storage_access_key.foobar]
}
`

var testAccSakuraMonitoringSuiteTraceStorageAccessKeyDataSource_basic = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_trace_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_trace_storage.foobar.id
  description = "access-key"
}

data "sakura_monitoring_suite_trace_storage_access_key" "foobar" {
  id = sakura_monitoring_suite_trace_storage_access_key.foobar.id
  storage_id = sakura_monitoring_suite_trace_storage.foobar.id

  depends_on = [sakura_monitoring_suite_trace_storage_access_key.foobar]
}
`
