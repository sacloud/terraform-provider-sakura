// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceWorkflowPlan_basic(t *testing.T) {
	resourceName := "data.sakura_workflows_plan.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSakuraDataSourceWorkflowPlan_basic,
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "200Kプラン"),
					resource.TestCheckResourceAttrSet(resourceName, "grade"),
					resource.TestCheckResourceAttrSet(resourceName, "base_price"),
					resource.TestCheckResourceAttrSet(resourceName, "included_steps"),
					resource.TestCheckResourceAttrSet(resourceName, "overage_step_unit"),
					resource.TestCheckResourceAttrSet(resourceName, "overage_price_per_unit"),
				),
			},
		},
	})
}

func TestAccSakuraDataSourceWorkflowPlan_multipleMatches(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSakuraDataSourceWorkflowPlan_multipleMatches,
				ExpectError: regexp.MustCompile(`multiple Workflow plans found with name containing 'プラン'.`),
			},
		},
	})
}

var testAccSakuraDataSourceWorkflowPlan_basic = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}`

var testAccSakuraDataSourceWorkflowPlan_multipleMatches = `
data "sakura_workflows_plan" "foobar" {
  name = "プラン"
}`
