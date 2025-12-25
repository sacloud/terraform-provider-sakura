// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package auto_scale_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAutoScale_basic(t *testing.T) {
	resourceName := "data.sakura_auto_scale.foobar"
	rand := test.RandomName()
	if !test.IsFakeModeEnabled() {
		test.SkipIfEnvIsNotSet(t, "SAKURA_API_KEY_ID")
	}
	apiKeyId := os.Getenv("SAKURA_API_KEY_ID")
	if apiKeyId == "" {
		apiKeyId = "111111111111" // dummy
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceAutoScale_basic, rand, apiKeyId),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),

					resource.TestCheckResourceAttr(resourceName, "zones.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "zones.*", "is1c"),
					resource.TestCheckResourceAttr(resourceName, "cpu_threshold_scaling.server_prefix", rand),
					resource.TestCheckResourceAttr(resourceName, "cpu_threshold_scaling.up", "80"),
					resource.TestCheckResourceAttr(resourceName, "cpu_threshold_scaling.down", "20"),
					resource.TestCheckResourceAttr(resourceName, "config", test.BuildConfigWithArgs(testAccSakuraAutoScale_encodedConfig, rand)),
					resource.TestCheckResourceAttrSet(resourceName, "api_key_id"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceAutoScale_basic = `
resource "sakura_server" "foobar" {
  name = "{{ .arg0 }}"
  force_shutdown = true
  zone = "is1c"
}

resource "sakura_auto_scale" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]

  zones  = ["is1c"]
  config = yamlencode({
    resources: [{
      type: "Server",
      selector: {
        names: [sakura_server.foobar.name],
        zones: ["is1c"],
      },
      shutdown_force: true,
    }],
  })

  api_key_id = "{{ .arg1 }}"

  trigger_type = "cpu"
  cpu_threshold_scaling = {
    server_prefix = sakura_server.foobar.name
    up   = 80
    down = 20
  }
}

data "sakura_auto_scale" "foobar" {
  name = sakura_auto_scale.foobar.name
}`
