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

func TestAccSakuraAddonStreaming_basic(t *testing.T) {
	resourceName := "sakura_addon_streaming.foobar"

	var streaming v1.GetResourceResponse
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraAddonStreamingDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAddonStreaming_basic),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAddonStreamingExists(resourceName, &streaming),
					resource.TestCheckResourceAttr(resourceName, "location", "japaneast"),
					resource.TestCheckResourceAttr(resourceName, "unit_count", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_name"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
		},
	})
}

func testCheckSakuraAddonStreamingExists(n string, res *v1.GetResourceResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no Addon Streaming ID is set")
		}

		client := test.AccClientGetter()
		found, err := addon.NewStreamingOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*res = *found
		return nil
	}
}

func testCheckSakuraAddonStreamingDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_addon_streaming" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := addon.NewStreamingOp(client.AddonClient).Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("Addon Streaming still exists")
		}
	}

	return nil
}

var testAccSakuraAddonStreaming_basic = `
resource "sakura_addon_streaming" "foobar" {
  location = "japaneast"
  unit_count = "3"
}
`
