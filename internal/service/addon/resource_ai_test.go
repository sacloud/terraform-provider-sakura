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

func TestAccSakuraAddonAI_basic(t *testing.T) {
	resourceName := "sakura_addon_ai.foobar"

	var ai v1.GetResourceResponse
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraAddonAIDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAddonAI_basic),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAddonAIExists(resourceName, &ai),
					resource.TestCheckResourceAttr(resourceName, "location", "japaneast"),
					resource.TestCheckResourceAttr(resourceName, "sku", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

func testCheckSakuraAddonAIExists(n string, res *v1.GetResourceResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no Addon AI ID is set")
		}

		client := test.AccClientGetter()
		found, err := addon.NewAIOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*res = *found
		return nil
	}
}

func testCheckSakuraAddonAIDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_addon_ai" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := addon.NewAIOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("Addon AI still exists")
		}
	}

	return nil
}

var testAccSakuraAddonAI_basic = `
resource "sakura_addon_ai" "foobar" {
  location = "japaneast"
  sku = 1
}
`
