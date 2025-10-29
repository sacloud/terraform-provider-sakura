// Copyright 2016-2025 terraform-provider-sakuracloud authors
// SPDX-License-Identifier: Apache-2.0

package zone_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceZone_basic(t *testing.T) {
	resourceName := "data.sakura_zone.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSakuraDataSourceZone_basic,
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "is1b"),
					resource.TestCheckResourceAttr(resourceName, "zone_id", "31002"),
					resource.TestCheckResourceAttr(resourceName, "description", "石狩第2ゾーン"),
					resource.TestCheckResourceAttr(resourceName, "region_id", "310"),
					resource.TestCheckResourceAttr(resourceName, "region_name", "石狩"),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.0", "133.242.0.3"),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.1", "133.242.0.4"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceZone_basic = `
data "sakura_zone" "foobar" { 
  name = "is1b"
}`
