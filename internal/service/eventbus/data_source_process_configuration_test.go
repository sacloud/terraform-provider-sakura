// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceProcessConfiguration_basic(t *testing.T) {
	resourceName := "data.sakura_eventbus_process_configuration.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceProcessConfiguration_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "destination", "simplenotification"),
					resource.TestCheckResourceAttr(resourceName, "parameters", "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"),
					resource.TestCheckNoResourceAttr(resourceName, "sakura_access_token_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "sakura_access_token_secret_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "credentials_wo_version"),
					resource.TestCheckNoResourceAttr(resourceName, "simplemq_api_key_wo"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceProcessConfiguration_basic = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1"]

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  sakura_access_token_wo        = "test"
  sakura_access_token_secret_wo = "test"
  credentials_wo_version        = 1
}

data "sakura_eventbus_process_configuration" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1"]

  depends_on = [
    sakura_eventbus_process_configuration.foobar
  ]
}`
