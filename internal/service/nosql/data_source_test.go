// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nosql_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v1 "github.com/sacloud/nosql-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

const envNosqlPassword = "SAKURA_NOSQL_PASSWORD"

func TestAccSakuraDataSourceNosql_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envNosqlPassword)

	resourceName := "data.sakura_nosql.foobar"
	rand := test.RandomName()
	password := os.Getenv(envNosqlPassword) // test.RandomPassword() often causes bad request

	var nosql v1.GetNosqlAppliance
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraNosqlDestroy,
			test.CheckSakuravSwitchDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceNosql_basic, rand, password),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraNosqlExists(resourceName, &nosql),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "plan", "100GB"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckNoResourceAttr(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "settings.reserve_ip_address", "192.168.0.10"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.full.interval", "14"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.full.day_of_week", "fri"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.full.time", "00:15"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.incremental.days_of_week.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.incremental.days_of_week.0", "mon"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.incremental.time", "00:15"),
					resource.TestCheckResourceAttr(resourceName, "settings.reserve_ip_address", "192.168.0.10"),
					resource.TestCheckResourceAttr(resourceName, "remark.nosql.default_user", "tftest"),
					resource.TestCheckResourceAttr(resourceName, "remark.nosql.port", "9042"),
					resource.TestCheckResourceAttr(resourceName, "remark.nosql.engine", "Cassandra"),
					resource.TestCheckResourceAttr(resourceName, "remark.nosql.nodes", "3"),
					resource.TestCheckResourceAttr(resourceName, "remark.servers.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "remark.servers.0", "192.168.0.7"),
					resource.TestCheckResourceAttr(resourceName, "remark.servers.1", "192.168.0.8"),
					resource.TestCheckResourceAttr(resourceName, "remark.servers.2", "192.168.0.9"),
					resource.TestCheckResourceAttr(resourceName, "remark.network.gateway", "192.168.0.1"),
					resource.TestCheckResourceAttr(resourceName, "remark.network.netmask", "24"),
				),
			},
		},
	})
}

const testAccSakuraDataSourceNosql_basic = `
resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_nosql" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1", "tag2"]
  zone = "tk1b"
  plan = "100GB"
  description = "description"
  password    = "{{ .arg1 }}"
  vswitch_id  = sakura_vswitch.foobar.id
  settings = {
    reserve_ip_address = "192.168.0.10"
    repair = {
      full = {
        interval = 14
        day_of_week = "fri"
        time = "00:15"
      }
      incremental = {
        days_of_week = ["mon"]
        time = "00:15"
      }
    }
  }
  remark = {
    nosql = {
      default_user = "tftest"
      port = 9042
    }
    servers = [
      "192.168.0.7",
      "192.168.0.8",
      "192.168.0.9",
    ]
    network = {
      gateway = "192.168.0.1"
      netmask = 24
    }
  }
  parameters = {
    concurrent_writes = "16"
    cas_contention_timeout = "1000ms"
  }
}

data "sakura_nosql" "foobar" {
  name = sakura_nosql.foobar.name
}
`
