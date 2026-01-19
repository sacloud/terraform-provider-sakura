// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceAPIGWPlan_basic(t *testing.T) {
	resourceName := "data.sakura_apigw_plan.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// description等は変動値の可能性もあるため、テストが環境によって失敗するようであればTestCheckResourceAttrSetに変更
				Config: testAccSakuraDataSourceAPIGWPlan_basic,
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "APIゲートウェイ for エンタープライズ"),
					resource.TestCheckResourceAttr(resourceName, "max_requests_unit", "month"),
					resource.TestCheckResourceAttrSet(resourceName, "max_requests"),
					resource.TestCheckResourceAttrSet(resourceName, "max_services"),
					resource.TestCheckResourceAttrSet(resourceName, "price"),
					resource.TestCheckResourceAttrSet(resourceName, "overage.unit_price"),
					resource.TestCheckResourceAttrSet(resourceName, "overage.unit_requests"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceAPIGWPlan_basic = `
data "sakura_apigw_plan" "foobar" { 
  name = "エンタープライズ"
}`
