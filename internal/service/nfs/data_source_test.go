// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nfs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceNFS_basic(t *testing.T) {
	resourceName := "data.sakura_nfs.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceNFS_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
					resource.TestCheckResourceAttrPair(
						resourceName, "network_interface.vswitch_id",
						"sakura_vswitch.foobar", "id",
					),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_address", "192.168.11.101"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.netmask", "24"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.gateway", "192.168.11.1"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceNFS_basic = `
resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_nfs" "foobar" {
  name = "{{ .arg0 }}"
  plan = "ssd"
  size = "500"
  description = "description"
  tags        = ["tag1", "tag2", "tag3"]

  network_interface = {
    vswitch_id = sakura_vswitch.foobar.id
    ip_address = "192.168.11.101"
    netmask    = 24
    gateway    = "192.168.11.1"
  }
}

data "sakura_nfs" "foobar" {
  name = sakura_nfs.foobar.name
}`
