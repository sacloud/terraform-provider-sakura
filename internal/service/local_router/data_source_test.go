// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package local_router_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceLocalRouter_basic(t *testing.T) {
	if !test.IsResourceRequiredTest() {
		t.Skip("This test only run if SAKURA_RESOURCE_REQUIRED_TEST environment variable is set")
	}

	resourceName := "data.sakura_local_router.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceLocalRouter_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
					resource.TestCheckResourceAttrPair(
						resourceName, "switch.code",
						"sakura_vswitch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "switch.category", "cloud"),
					resource.TestCheckResourceAttrPair(
						resourceName, "switch.zone",
						"data.sakura_zone.current", "name"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.vip", "192.168.11.1"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.0", "192.168.11.11"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.1", "192.168.11.12"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.netmask", "24"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.vrid", "1"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceLocalRouter_basic = `
resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_zone" "current" {}

resource "sakura_local_router" "foobar" {
  switch = {
    code     = sakura_vswitch.foobar.id
    category = "cloud"
    zone     = data.sakura_zone.current.name
  }
  network_interface = {
    vip          = "192.168.11.1"
    ip_addresses = ["192.168.11.11", "192.168.11.12"]
    netmask      = 24
    vrid         = 1
  }

  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2", "tag3"]
}

data "sakura_local_router" "foobar" {
  name = sakura_local_router.foobar.name
}`
