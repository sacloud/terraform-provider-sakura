// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
					resource.TestCheckNoResourceAttr(resourceName, "simplenotification_credentials_wo_version"),
					resource.TestCheckNoResourceAttr(resourceName, "simplenotification_access_token_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "simplenotification_access_token_secret_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "simplemq_credentials_wo_version"),
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

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  simplenotification_access_token_wo        = "test"
  simplenotification_access_token_secret_wo = "test"
  simplenotification_credentials_wo_version = 1
}

data "sakura_eventbus_process_configuration" "foobar" {
  name = "{{ .arg0 }}"

  depends_on = [
    sakura_eventbus_process_configuration.foobar
  ]
}`
