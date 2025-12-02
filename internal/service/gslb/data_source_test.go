// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package gslb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceGSLB_basic(t *testing.T) {
	resourceName := "data.sakura_gslb.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceGSLB_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "health_check.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "health_check.delay_loop", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.host_header", "usacloud.jp"),
					resource.TestCheckResourceAttr(resourceName, "health_check.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check.status", "200"),
					resource.TestCheckResourceAttr(resourceName, "sorry_server", "8.8.8.8"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceGSLB_basic = `
resource "sakura_gslb" "foobar" {
  name = "{{ .arg0 }}"
  health_check = {
    protocol    = "http"
    delay_loop  = 10
    host_header = "usacloud.jp"
    port        = "80"
    path        = "/"
    status      = "200"
  }
  sorry_server = "8.8.8.8"
  description  = "description"
  tags         = ["tag1", "tag2", "tag3"]
}

data "sakura_gslb" "foobar" {
  name = sakura_gslb.foobar.name
}`
