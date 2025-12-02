// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_monitor_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceSimpleMonitor_basic(t *testing.T) {
	resourceName := "data.sakura_simple_monitor.foobar"
	target := fmt.Sprintf("%s.com", test.RandomName())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSimpleMonitor_basic, target, testAccSlackWebhook),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target", target),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "delay_loop", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "health_check.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check.status", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.host_header", "usacloud.jp"),
					resource.TestCheckResourceAttr(resourceName, "notify_email_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notify_slack_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notify_slack_webhook", testAccSlackWebhook),
				),
			},
		},
	})
}

var testAccSakuraDataSourceSimpleMonitor_basic = `
resource "sakura_simple_monitor" "foobar" {
  target      = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2", "tag3"]
  delay_loop  = 60
  health_check = {
    protocol    = "http"
    path        = "/"
    status      = 200
    host_header = "usacloud.jp"
  }
  notify_email_enabled = true
  notify_slack_enabled = true
  notify_slack_webhook = "{{ .arg1 }}"
}

data "sakura_simple_monitor" "foobar" {
  target = sakura_simple_monitor.foobar.target
}`
