// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package enhanced_lb_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraCloudDataSourceProxyLB_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envEnhancedLBRealServerIP0, envEnhancedLBRealServerIP1)

	resourceName := "data.sakura_enhanced_lb.foobar"
	rand := test.RandomName()
	ip0 := os.Getenv(envEnhancedLBRealServerIP0)
	ip1 := os.Getenv(envEnhancedLBRealServerIP1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceEnhancedLB_basic, rand, ip0, ip1),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "plan", "100"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "health_check.delay_loop", "20"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.proxy_mode", "http"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "server.0.ip_address", ip0),
					resource.TestCheckResourceAttr(resourceName, "server.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "server.0.group", "group1"),
					resource.TestCheckResourceAttr(resourceName, "server.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server.1.ip_address", ip1),
					resource.TestCheckResourceAttr(resourceName, "server.1.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "server.1.group", "group2"),
					resource.TestCheckResourceAttr(resourceName, "server.1.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.host", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.0.path", "/path1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.group", "group1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.host", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.1.path", "/path2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.group", "group2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.request_header_name", "foo"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.request_header_value", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.request_header_value_ignore_case", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.request_header_value_not_match", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.request_header_name", "bar"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.request_header_value", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.request_header_value_ignore_case", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.request_header_value_not_match", "false"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_suite.enabled", "true"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceEnhancedLB_basic = `
resource "sakura_enhanced_lb" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2", "tag3"]

  health_check = {
    protocol   = "tcp"
    delay_loop = 20
  }

  bind_port = [{
    proxy_mode = "http"
    port       = 80
  }]

  server = [{
    ip_address = "{{ .arg1 }}"
    port       = 80
    group      = "group1"
  },
  {
    ip_address = "{{ .arg2 }}"
    port       = 80
    group      = "group2"
  }]

  rule = [{
    path  = "/path1"
    group = "group1"
    request_header_name = "foo"
    request_header_value = "1"
    request_header_value_ignore_case = true
    request_header_value_not_match = true
  },
  {
    path  = "/path2"
    group = "group2"
    request_header_name = "bar"
    request_header_value = "2"
    request_header_value_ignore_case = false
    request_header_value_not_match = false
  }]

  monitoring_suite = {
    enabled = true
  }
}

data "sakura_enhanced_lb" "foobar" {
  name = sakura_enhanced_lb.foobar.name
}`
