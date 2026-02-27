// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAddonSearch_Basic(t *testing.T) {
	resourceName := "data.sakura_addon_search.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceAddonSearchConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "location", "japaneast"),
					resource.TestCheckResourceAttr(resourceName, "partition_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "replica_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "sku", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceAddonSearchConfig = `
resource "sakura_addon_search" "foobar" {
  location = "japaneast"
  partition_count = 1
  replica_count = 1
  sku = 2
}

data "sakura_addon_search" "foobar" {
  id = sakura_addon_search.foobar.id
}`
