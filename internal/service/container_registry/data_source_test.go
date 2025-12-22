// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package container_registry_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceContainerRegistry_basic(t *testing.T) {
	resourceName := "data.sakura_container_registry.foobar"
	rand := test.RandomName()
	prefix := acctest.RandStringFromCharSet(60, acctest.CharSetAlpha)
	password := test.RandomPassword()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceContainerRegistry_basic, rand, prefix, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "user.0.name", "user1"),
					resource.TestCheckResourceAttr(resourceName, "user.0.permission", "readwrite"),
					resource.TestCheckResourceAttr(resourceName, "user.1.name", "user2"),
					resource.TestCheckResourceAttr(resourceName, "user.1.permission", "readonly"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceContainerRegistry_basic = `
resource "sakura_container_registry" "foobar" {
  name            = "{{ .arg0 }}"
  subdomain_label = "{{ .arg1 }}"
  access_level    = "readwrite"

  description = "description"
  tags        = ["tag1", "tag2"]

  user = [{
    name       = "user1"
    password   = "{{ .arg2 }}"
    permission = "readwrite"
  },
  {
    name     = "user2"
    password = "{{ .arg2 }}"
    permission = "readonly"
  }]
}

data "sakura_container_registry" "foobar" {
  name = sakura_container_registry.foobar.name
}`
