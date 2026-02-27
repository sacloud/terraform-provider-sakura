// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAddonQuery_Basic(t *testing.T) {
	resourceName := "data.sakura_addon_query.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceAddonQueryConfig),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "location", "japaneast"),
					resource.TestCheckNoResourceAttr(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceAddonQueryConfig = `
resource "sakura_addon_query" "foobar" {
  location = "japaneast"
}

data "sakura_addon_query" "foobar" {
  id = sakura_addon_query.foobar.id
}`
