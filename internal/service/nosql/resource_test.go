// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nosql_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/nosql-api-go"
	v1 "github.com/sacloud/nosql-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraNosql_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envNosqlPassword)

	resourceName := "sakura_nosql.foobar"
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
				Config: test.BuildConfigWithArgs(testAccSakuraNosql_basic, rand, password),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraNosqlExists(resourceName, &nosql),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "plan", "100GB"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.concurrent_writes", "16"),
					resource.TestCheckResourceAttr(resourceName, "parameters.cas_contention_timeout", "1000ms"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNosql_update, rand, password),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraNosqlExists(resourceName, &nosql),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "plan", "100GB"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2-upd"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.reserve_ip_address", "192.168.0.10"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.full.interval", "14"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.full.day_of_week", "fri"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.full.time", "00:00"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.incremental.days_of_week.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.incremental.days_of_week.0", "mon"),
					resource.TestCheckResourceAttr(resourceName, "settings.repair.incremental.time", "00:00"),
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
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.concurrent_writes", "32"),
					resource.TestCheckResourceAttr(resourceName, "parameters.cas_contention_timeout", "1000ms"),
				),
			},
		},
	})
}

func testCheckSakuraNosqlExists(n string, instance *v1.GetNosqlAppliance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Database ID is set")
		}

		client := test.AccClientGetter()
		dbOp := nosql.NewDatabaseOp(client.NosqlClient)
		foundNosql, err := dbOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundNosql.ID.Value != rs.Primary.ID {
			return fmt.Errorf("resource Database[%s] not found", rs.Primary.ID)
		}

		*instance = *foundNosql
		return nil
	}
}

func testCheckSakuraNosqlDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_nosql" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		dbOp := nosql.NewDatabaseOp(client.NosqlClient)
		_, err := dbOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("resource Nosql[%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraNosql_basic = `
resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_nosql" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1", "tag2"]
  zone = "tk1b"
  plan = "100GB"
  description = "description"
  password_wo = "{{ .arg1 }}"
  password_wo_version = 1
  vswitch_id  = sakura_vswitch.foobar.id
  settings = {
    reserve_ip_address = "192.168.0.10"
	/* TODO: Add test with nfs backup
    backup = {
      connect = "nfs://192.168.0.31/export"
      days_of_week = ["sun"]
      time = "00:30"
      rotate = 2
    }
	 */
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
`

const testAccSakuraNosql_update = `
resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}
resource "sakura_nosql" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1-upd", "tag2-upd"]
  zone = "tk1b"
  plan = "100GB"
  description = "description-updated"
  password_wo = "{{ .arg1 }}"
  password_wo_version = 1
  vswitch_id  = sakura_vswitch.foobar.id
  settings = {
    reserve_ip_address = "192.168.0.10"
    repair = {
      full = {
        interval = 14
        day_of_week = "fri"
        time = "00:00"
      }
      incremental = {
        days_of_week = ["mon"]
        time = "00:00"
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
    concurrent_writes = "32"
    cas_contention_timeout = "1000ms"
  }
}`
