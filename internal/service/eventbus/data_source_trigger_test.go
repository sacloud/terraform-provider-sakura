// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceTrigger_basic(t *testing.T) {
	resourceName := "data.sakura_eventbus_trigger.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceTrigger_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttrPair(resourceName, "process_configuration_id",
						"sakura_eventbus_process_configuration.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "source", "test-source"),
					resource.TestCheckResourceAttr(resourceName, "types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "types.0", "type1"),
					resource.TestCheckResourceAttr(resourceName, "types.1", "type2"),
					resource.TestCheckResourceAttr(resourceName, "conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.op", "eq"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.values.0", "value1"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceTrigger_basic = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  sakura_access_token_wo        = "test"
  sakura_access_token_secret_wo = "test"
  credentials_wo_version        = 1
}

resource "sakura_eventbus_trigger" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  source                   = "test-source"
  types                    = ["type1", "type2"]
  conditions               = [
    {
      key    = "key1"
      op     = "eq"
      values = ["value1"]
    },
  ]
}

data "sakura_eventbus_trigger" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1"]

  depends_on = [
    sakura_eventbus_trigger.foobar
  ]
}`
