// Copyright 2016-2025 terraform-provider-sakuracloud authors
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

package event_bus_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/test"
)

func TestAccSakuraDataSourceProcessConfiguration_basic(t *testing.T) {
	resourceName := "data.sakura_event_bus_process_configuration.foobar"
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
					resource.TestCheckResourceAttr(resourceName, "destination", "unknown"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceProcessConfiguration_basic = `
resource "sakura_event_bus_process_configuration" "foobar" {
  name          = "{{ .arg0 }}"

	destination = "unknown"
	undefined = "hello"
}

data "sakura_event_bus_process_configuration" "foobar" {
	id = sakura_event_bus.foobar.id
}`
