// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package sw1tch_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccImportSakuraSwitch_basic(t *testing.T) {
	resourceName := "sakura_switch.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuravSwitchDestroy,
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

var testAccSakuraSwitch_import = `
resource "sakura_switch" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
}
`
