// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAPIGWSubscription_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "data.sakura_apigw_subscription.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSubscription_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "monthly_request"),
					resource.TestCheckNoResourceAttr(resourceName, "service"),
					resource.TestCheckResourceAttrPair(resourceName, "plan_id", "data.sakura_apigw_plan.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceSubscription_basic = `
data "sakura_apigw_plan" "foobar" { 
  name = "エンタープライズ"
}

resource "sakura_apigw_subscription" "foobar" {
  name = "{{ .arg0 }}"
  plan_id = data.sakura_apigw_plan.foobar.id
}
  
data "sakura_apigw_subscription" "foobar" {
  name = sakura_apigw_subscription.foobar.name
}`
