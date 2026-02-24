// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMRole_Basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "data.sakura_iam_role.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMRoleConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "id", "securitycontrol-agent"),
					resource.TestCheckResourceAttr(resourceName, "name", "セキュリティコントロールエージェント"),
					resource.TestCheckResourceAttr(resourceName, "description", "セキュリティコントロールの全評価ルールの実行ができる"),
					resource.TestCheckResourceAttr(resourceName, "category", "securitycontrol"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceIAMRoleConfig = `
data "sakura_iam_role" "foobar" {
  id = "securitycontrol-agent"
}`
