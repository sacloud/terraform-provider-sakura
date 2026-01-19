// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceCert_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "data.sakura_apigw_cert.foobar"
	rand := test.RandomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceCert_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttrSet(resourceName, "rsa.expired_at"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceCert_basic = testSetupAPIGWSub + `
resource "sakura_apigw_cert" "foobar" {
  name = "{{ .arg0 }}"
  rsa  = {
    cert_wo = file("testdata/rsa.crt")
    key_wo = file("testdata/rsa.key")
    cert_wo_version = 1
  }
}

data "sakura_apigw_cert" "foobar" {
  name = sakura_apigw_cert.foobar.name
}`
