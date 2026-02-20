// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMSSO_Basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "data.sakura_iam_sso.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMSSOConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "idp_entity_id", "https://idp.example.com/ile2ephei7saeph6"),
					resource.TestCheckResourceAttr(resourceName, "idp_login_url", "https://idp.example.com/ile2ephei7saeph6/sso/login"),
					resource.TestCheckResourceAttr(resourceName, "idp_logout_url", "https://idp.example.com/ile2ephei7saeph6/sso/logout"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "sp_entity_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sp_acs_url"),
					resource.TestCheckResourceAttrSet(resourceName, "assigned"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceIAMSSOConfig = `
resource "sakura_iam_sso" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  idp_entity_id = "https://idp.example.com/ile2ephei7saeph6"
  idp_login_url = "https://idp.example.com/ile2ephei7saeph6/sso/login"
  idp_logout_url = "https://idp.example.com/ile2ephei7saeph6/sso/logout"
  idp_certificate = file("testdata/rsa.crt")
}

data "sakura_iam_sso" "foobar" {
  name = sakura_iam_sso.foobar.name
}`
