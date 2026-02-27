// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/addon-api-go"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraAddonDDoS_basic(t *testing.T) {
	resourceName := "sakura_addon_ddos.foobar"

	var ddos v1.GetResourceResponse
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraAddonDDoSDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAddonDDoS_basic),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAddonDDoSExists(resourceName, &ddos),
					resource.TestCheckResourceAttr(resourceName, "location", "japaneast"),
					resource.TestCheckResourceAttr(resourceName, "pricing_level", "1"),
					resource.TestCheckResourceAttr(resourceName, "patterns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "patterns.0", "/*"),
					resource.TestCheckResourceAttr(resourceName, "origin.hostname", "www.usacloud.jp"),
					resource.TestCheckResourceAttr(resourceName, "origin.host_header", "usacloud.jp"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

func testCheckSakuraAddonDDoSExists(n string, res *v1.GetResourceResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no Addon DDoS ID is set")
		}

		client := test.AccClientGetter()
		found, err := addon.NewDDoSOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*res = *found
		return nil
	}
}

func testCheckSakuraAddonDDoSDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_addon_ddos" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := addon.NewDDoSOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("Addon DDoS still exists")
		}
	}

	return nil
}

var testAccSakuraAddonDDoS_basic = `
resource "sakura_addon_ddos" "foobar" {
  location = "japaneast"
  pricing_level = 1
  patterns = ["/*"]
  origin = {
    hostname = "www.usacloud.jp"
    host_header = "usacloud.jp"
  }
}
`
