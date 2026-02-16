// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMUser_Basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "data.sakura_iam_user.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMUserConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "code", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "email", "tf-test-ds-email@sakura.ad.jp"),
					resource.TestCheckResourceAttrSet(resourceName, "passwordless"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "otp.status"),
					resource.TestCheckResourceAttrSet(resourceName, "otp.has_recovery_code"),
					resource.TestCheckResourceAttrSet(resourceName, "member.id"),
					resource.TestCheckResourceAttrSet(resourceName, "member.code"),
					resource.TestCheckResourceAttrSet(resourceName, "security_key_registered"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceIAMUserConfig = `
resource "sakura_iam_user" "foobar" {
  name        = "{{ .arg0 }}"
  code        = "{{ .arg0 }}"
  description = "description"
  email       = "tf-test-ds-email@sakura.ad.jp"
  password_wo = "PassWord!12345678"
  password_wo_version = 1
}

data "sakura_iam_user" "foobar" {
  name = sakura_iam_user.foobar.name
}`
