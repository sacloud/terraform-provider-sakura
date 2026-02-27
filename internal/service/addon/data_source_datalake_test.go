// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAddonDataLake_Basic(t *testing.T) {
	resourceName := "data.sakura_addon_datalake.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceAddonDataLakeConfig),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "location", "japaneast"),
					resource.TestCheckResourceAttr(resourceName, "performance", "1"),
					resource.TestCheckResourceAttr(resourceName, "redundancy", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceAddonDataLakeConfig = `
resource "sakura_addon_datalake" "foobar" {
  location = "japaneast"
  performance = 1
  redundancy = 1
}

data "sakura_addon_datalake" "foobar" {
  id = sakura_addon_datalake.foobar.id
}`
