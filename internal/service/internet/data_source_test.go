// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package internet_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceInternet_basic(t *testing.T) {
	resourceName := "data.sakura_internet.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceInternet_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "netmask", "28"),
					resource.TestCheckResourceAttr(resourceName, "band_width", "100"),
					resource.TestCheckResourceAttr(resourceName, "server_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ip_addresses.#", "11"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceInternet_basic = `
resource "sakura_internet" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2", "tag3"]
}

data "sakura_internet" "foobar" {
  name = sakura_internet.foobar.name
}`
