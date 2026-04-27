// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMAuthContext_Basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "data.sakura_iam_auth_context.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMAuthContextConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "auth_type"),
					resource.TestCheckResourceAttrSet(resourceName, "limited_to_project_id"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceIAMAuthContextConfig = `
data "sakura_iam_auth_context" "foobar" {}
`
