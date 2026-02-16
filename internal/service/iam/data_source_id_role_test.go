// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMIDRole_Basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "data.sakura_iam_id_role.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMIDRoleConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "id", "identity-admin"),
					resource.TestCheckResourceAttr(resourceName, "name", "ID管理者"),
					resource.TestCheckResourceAttr(resourceName, "description", "ID階層に関する全操作が行える"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceIAMIDRoleConfig = `
data "sakura_iam_id_role" "foobar" {
  id = "identity-admin"
}`
