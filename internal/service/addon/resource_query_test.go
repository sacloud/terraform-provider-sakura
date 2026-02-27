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

func TestAccSakuraAddonQuery_basic(t *testing.T) {
	resourceName := "sakura_addon_query.foobar"

	var query v1.GetResourceResponse
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraAddonQueryDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAddonQuery_basic),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAddonQueryExists(resourceName, &query),
					resource.TestCheckResourceAttr(resourceName, "location", "japaneast"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

func testCheckSakuraAddonQueryExists(n string, res *v1.GetResourceResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no Addon Query ID is set")
		}

		client := test.AccClientGetter()
		found, err := addon.NewQueryOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*res = *found
		return nil
	}
}

func testCheckSakuraAddonQueryDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_addon_query" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := addon.NewQueryOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("Addon Query still exists")
		}
	}

	return nil
}

var testAccSakuraAddonQuery_basic = `
resource "sakura_addon_query" "foobar" {
  location = "japaneast"
}
`
