// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceSchedule_basic(t *testing.T) {
	resourceName := "data.sakura_eventbus_schedule.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSchedule_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "recurring_step", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurring_unit", "day"),
					resource.TestCheckResourceAttr(resourceName, "starts_at", "1700000000000"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceSchedule_basic = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  simplenotification_access_token_wo        = "test"
  simplenotification_access_token_secret_wo = "test"
  credentials_wo_version                    = 1
}

resource "sakura_eventbus_schedule" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  recurring_step           = 1
  recurring_unit           = "day"
  starts_at                = 1700000000000
}

data "sakura_eventbus_schedule" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1"]

  depends_on = [
    sakura_eventbus_schedule.foobar
  ]
}`
