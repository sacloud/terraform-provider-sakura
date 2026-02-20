// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraIAMAuth_basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "sakura_iam_auth.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMAuth_basic),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "password_policy.min_length", "10"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.require_uppercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.require_lowercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.require_symbols", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "conditions.ip_restriction.mode"),
					resource.TestCheckResourceAttrSet(resourceName, "conditions.require_two_factor_auth"),
					resource.TestCheckResourceAttr(resourceName, "conditions.datetime_restriction.%", "2"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMAuth_update),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "password_policy.min_length", "8"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.require_uppercase", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.require_lowercase", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.require_symbols", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "conditions.ip_restriction.mode"),
					resource.TestCheckResourceAttrSet(resourceName, "conditions.require_two_factor_auth"),
					resource.TestCheckResourceAttr(resourceName, "conditions.datetime_restriction.%", "2"),
				),
			},
		},
	})
}

const testAccSakuraIAMAuth_basic = `
resource "sakura_iam_auth" "foobar" {
  password_policy = {
    min_length = 10
    require_uppercase = true
    require_lowercase = true
    require_symbols = false
  }
  /* 現状Conditions系の設定はAPIから完全にdisableにできないためテストではコメントアウト
  conditions = {
    ip_restriction = {
      mode = "allow_list"
      #source_network = ["192.168.10.1"]
    }
    require_two_factor_auth = false
    datetime_restriction = {
      after = "2025-09-01T00:00:00+09:00",
      before = "2026-09-01T00:00:00+09:00",
    }
  }
  */
}
`

const testAccSakuraIAMAuth_update = `
resource "sakura_iam_auth" "foobar" {
  password_policy = {
    min_length = 8
    require_uppercase = false
    require_lowercase = false
    require_symbols = false
  }
  /* 現状Conditions系の設定はAPIから完全にdisableにできないためテストではコメントアウト
  conditions = {
    ip_restriction = {
      mode = "allow_list"
      #source_network = ["192.168.10.1"]
    }
    require_two_factor_auth = false
    datetime_restriction = {
      after = "2025-09-01T00:00:00+09:00",
      before = "2026-09-01T00:00:00+09:00",
    }
  }
  */
}`
