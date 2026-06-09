// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package sw1tch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccImportSakuraSwitch_basic(t *testing.T) {
	resourceName := "sakura_switch.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSwitchDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSwitch_import, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"zone",
				},
			},
		},
	})
}

func testCheckSakuraSwitchDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	swOp := iaas.NewSwitchOp(client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_switch" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := swOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists Switch: %s", rs.Primary.ID)
		}
	}

	return nil
}

var testAccSakuraSwitch_import = `
resource "sakura_switch" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
}
`
