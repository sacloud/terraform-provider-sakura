// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package nlb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceNLB_basic(t *testing.T) {
	if !test.IsFakeModeEnabled() {
		test.SkipIfEnvIsNotSet(t, "SAKURA_ENABLE_NLB_TEST")
	}

	resourceName := "data.sakura_nlb.foobar"
	rand := test.RandomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceNLB_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceNLB_basic = `
resource sakura_vswitch "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_nlb" "foobar" {
  network_interface = {
    vswitch_id   = sakura_vswitch.foobar.id
    vrid         = 1
    ip_addresses = ["192.168.11.101"]
    netmask      = 24
    gateway      = "192.168.11.1"
  }

  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2", "tag3"]
}

data "sakura_nlb" "foobar" {
  name = sakura_nlb.foobar.name
}`
