// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ssh_key_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceSSHKey_basic(t *testing.T) {
	resourceName := "data.sakura_ssh_key.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSSHKey_basic, rand, testAccPublicKey),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "public_key", testAccPublicKey),
					resource.TestCheckResourceAttr(resourceName, "fingerprint", testAccFingerprint),
				),
			},
		},
	})
}

var testAccSakuraDataSourceSSHKey_basic = `
resource "sakura_ssh_key" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  public_key  = "{{ .arg1 }}"
}

data "sakura_ssh_key" "foobar" {
  name = sakura_ssh_key.foobar.name
}`
