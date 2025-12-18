// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package database_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceDatabase_basic(t *testing.T) {
	resourceName := "data.sakura_database.foobar"
	rand := test.RandomName()
	password := test.RandomPassword()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceDatabase_basic, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "plan", "10g"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "backup.days_of_week.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "backup.days_of_week.0", "mon"),
					resource.TestCheckResourceAttr(resourceName, "backup.days_of_week.1", "tue"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceDatabase_basic = `
resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_database" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2", "tag3"]

  username = "defuser"
  password = "{{ .arg1 }}"

  network_interface = {
    vswitch_id    = sakura_vswitch.foobar.id
    ip_address    = "192.168.101.101"
    netmask       = 24
    gateway       = "192.168.101.1"
    port          = 54321
    source_ranges = ["192.168.101.0/24", "192.168.102.0/24"]
  }
  backup = {
    days_of_week = ["mon", "tue"]
    time         = "00:00"
  }
}

data "sakura_database" "foobar" {
  name = sakura_database.foobar.name
}`
