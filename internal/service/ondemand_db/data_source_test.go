// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ondemand_db_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceOnDemandDB_basic(t *testing.T) {
	resourceName := "data.sakura_ondemand_db.foobar"
	rand := test.RandomName()
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	password := test.RandomPassword()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceOnDemandDB_basic, rand, databaseName, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "database_name", databaseName),
					resource.TestCheckResourceAttr(resourceName, "database_type", "tidb"),
					resource.TestCheckResourceAttr(resourceName, "region", "tk1"),
					resource.TestCheckResourceAttr(resourceName, "max_connections", "50"),
					resource.TestCheckResourceAttr(resourceName, "hostname", databaseName+".tidb-tk1.db.sakurausercontent.com"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceOnDemandDB_basic = `
resource "sakura_ondemand_db" "foobar" {
  name            = "{{ .arg0 }}"
  database_name   = "{{ .arg1 }}"
  database_type   = "tidb"
  region          = "tk1"
  password_wo     = "{{ .arg2 }}"
  password_wo_version = 1

  description = "description"
  tags        = ["tag1", "tag2"]
}

data "sakura_ondemand_db" "foobar" {
  name = sakura_ondemand_db.foobar.name
}`
