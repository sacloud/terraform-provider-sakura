// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cdrom_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceCDROM_basic(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	resourceName := "data.sakura_cdrom.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraCDROMDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceCDROM_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
				),
			},
		},
	})
}

var testAccSakuraDataSourceCDROM_basic = `
resource "sakura_cdrom" "foobar" {
  name           = "{{ .arg0 }}"
  size           = 5
  iso_image_file = "test/dummy.iso"
  description    = "description"
  tags           = ["tag1", "tag2"]
}

data "sakura_cdrom" "foobar" {
  name = sakura_cdrom.foobar.name
}
`
