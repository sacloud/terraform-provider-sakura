// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAPIGWDomain_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "data.sakura_apigw_domain.foobar"
	rand := test.RandomName()
	domain := rand + ".com"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceDomain_basic, domain, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", domain),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_name"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceDomain_basic = testSetupAPIGWSub + `
resource "sakura_apigw_cert" "foobar" {
  name = "{{ .arg1 }}"
  rsa  = {
    cert_wo = file("testdata/rsa.crt")
    key_wo = file("testdata/rsa.key")
    cert_wo_version = 1
  }
}

resource "sakura_apigw_domain" "foobar" {
  name = "{{ .arg0 }}"
  certificate_id = sakura_apigw_cert.foobar.id
}

data "sakura_apigw_domain" "foobar" {
  name = sakura_apigw_domain.foobar.name
}`
