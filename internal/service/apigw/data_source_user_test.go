// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAPIGWUser_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "data.sakura_apigw_user.foobar"
	rand := test.RandomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceUser_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "custom_id", "custom-test-id-3500"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.protocols", "http"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.restricted_by", "allowIps"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.0", "192.168.1.10"),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", rand),
					resource.TestCheckResourceAttr(resourceName, "authentication.basic_auth.username", rand+"-ds-user"),
					resource.TestCheckResourceAttr(resourceName, "authentication.jwt.key", rand+"-ds-key"),
					resource.TestCheckNoResourceAttr(resourceName, "authentication.hmac_auth"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceUser_basic = testSetupAPIGWSub + `
resource "sakura_apigw_group" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_apigw_user" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1"]
  custom_id = "custom-test-id-3500"
  ip_restriction = {
    protocols = "http"
    restricted_by = "allowIps"
    ips = ["192.168.1.10"]
  }
  groups = [{name = "{{ .arg0 }}"}]
  authentication = {
    basic_auth = {
       username = "{{ .arg0 }}-ds-user",
       password_wo = "password"
	   password_wo_version = 1
    },
    jwt = {
      key = "{{ .arg0 }}-ds-key",
      secret_wo = "secret",
	  secret_wo_version = 1,
      algorithm = "HS256"
    },
  }
}

data "sakura_apigw_user" "foobar" {
  name = sakura_apigw_user.foobar.name
}`
