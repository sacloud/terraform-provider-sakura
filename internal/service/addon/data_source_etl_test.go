// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAddonETL_Basic(t *testing.T) {
	resourceName := "data.sakura_addon_etl.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceAddonETLConfig),
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

var testAccCheckSakuraDataSourceAddonETLConfig = `
resource "sakura_addon_etl" "foobar" {
  location = "japaneast"
}

data "sakura_addon_etl" "foobar" {
  id = sakura_addon_etl.foobar.id
}`
