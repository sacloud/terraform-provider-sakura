// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAPIGWService_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION", "SAKURA_APIGW_SERVICE_HOST")

	resourceName := "data.sakura_apigw_service.foobar"
	rand := test.RandomName()
	host := os.Getenv("SAKURA_APIGW_SERVICE_HOST")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceService_basic, rand, host),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "read_timeout", "60000"),
					resource.TestCheckResourceAttr(resourceName, "read_timeout", "60000"),
					resource.TestCheckResourceAttr(resourceName, "write_timeout", "60000"),
					resource.TestCheckResourceAttr(resourceName, "connect_timeout", "60000"),
					resource.TestCheckResourceAttr(resourceName, "authentication", "none"),
					resource.TestCheckResourceAttrSet(resourceName, "route_host"),
					resource.TestCheckNoResourceAttr(resourceName, "oidc"),
					resource.TestCheckNoResourceAttr(resourceName, "cors_config"),
					resource.TestCheckNoResourceAttr(resourceName, "object_storage_config"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_id", "sakura_apigw_subscription.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceService_basic = testSetupAPIGWSub + `
resource "sakura_apigw_service" "foobar" {
  name     = "{{ .arg0 }}"
  tags     =  ["tag1"]
  protocol = "http"
  host     = "{{ .arg1 }}"
  port     = 8080
  retries  = 3
  subscription_id = sakura_apigw_subscription.foobar.id
}

data "sakura_apigw_service" "foobar" {
  name = sakura_apigw_service.foobar.name
}`
