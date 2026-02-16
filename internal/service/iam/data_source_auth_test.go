// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMAuth_Basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "data.sakura_iam_auth.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMAuthConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "password_policy.min_length"),
					resource.TestCheckResourceAttrSet(resourceName, "password_policy.require_uppercase"),
					resource.TestCheckResourceAttrSet(resourceName, "password_policy.require_lowercase"),
					resource.TestCheckResourceAttrSet(resourceName, "password_policy.require_symbols"),
					resource.TestCheckResourceAttrSet(resourceName, "conditions.ip_restriction.mode"),
					resource.TestCheckResourceAttrSet(resourceName, "conditions.require_two_factor_auth"),
					resource.TestCheckResourceAttr(resourceName, "conditions.datetime_restriction.%", "2"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceIAMAuthConfig = `
data "sakura_iam_auth" "foobar" {}`
