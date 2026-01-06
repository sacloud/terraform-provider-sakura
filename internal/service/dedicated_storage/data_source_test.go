// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dedicated_storage_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceDedicatedStorage_basic(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	resourceName := "data.sakura_dedicated_storage.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceDedicatedStorage_basic, rand),
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

var testAccSakuraDataSourceDedicatedStorage_basic = `
resource "sakura_dedicated_storage" "foobar"{
  name        = "{{ .arg0 }}"
  tags        = ["tag1", "tag2", "tag3"]
  description = "description"
}

data "sakura_dedicated_storage" "foobar" {
  name = sakura_dedicated_storage.foobar.name
}`
