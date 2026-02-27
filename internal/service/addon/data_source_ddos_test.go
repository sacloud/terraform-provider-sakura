// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAddonDDoS_Basic(t *testing.T) {
	resourceName := "data.sakura_addon_ddos.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceAddonDDoSConfig),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "pricing_level", "1"),
					resource.TestCheckResourceAttr(resourceName, "patterns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "patterns.0", "/*"),
					resource.TestCheckResourceAttr(resourceName, "origin.hostname", "www.usacloud.jp"),
					resource.TestCheckResourceAttr(resourceName, "origin.host_header", "usacloud.jp"),
					resource.TestCheckNoResourceAttr(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceAddonDDoSConfig = `
resource "sakura_addon_ddos" "foobar" {
  location = "japaneast"
  pricing_level = 1
  patterns = ["/*"]
  origin = {
    hostname = "www.usacloud.jp"
    host_header = "usacloud.jp"
  }
}

data "sakura_addon_ddos" "foobar" {
  id = sakura_addon_ddos.foobar.id
}`
