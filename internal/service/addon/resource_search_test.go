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

func TestAccSakuraAddonSearch_basic(t *testing.T) {
	resourceName := "sakura_addon_search.foobar"
	rand := test.RandomName()

	var addonSearch v1.GetResourceResponse
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraAddonSearchDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAddonSearch_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAddonSearchExists(resourceName, &addonSearch),
					resource.TestCheckResourceAttr(resourceName, "location", "japaneast"),
					resource.TestCheckResourceAttr(resourceName, "partition_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "replica_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "sku", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

func testCheckSakuraAddonSearchExists(n string, res *v1.GetResourceResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no Addon Search ID is set")
		}

		client := test.AccClientGetter()
		found, err := addon.NewSearchOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*res = *found
		return nil
	}
}

func testCheckSakuraAddonSearchDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_addon_search" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := addon.NewSearchOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("Addon Search still exists")
		}
	}

	return nil
}

var testAccSakuraAddonSearch_basic = `
resource "sakura_addon_search" "foobar" {
  location = "japaneast"
  partition_count = 1
  replica_count = 1
  sku = 2
}`
