// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

func TestAccSakuraDataSourceWorkflowsSubscription_basic(t *testing.T) {
	resourceName := "data.sakura_workflows_subscription.foobar"
	var subscription v1.GetSubscriptionOK

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSakuraDataSourceWorkflowsSubscription_basic,
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsSubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "plan_name", "200Kプラン"),
					resource.TestCheckResourceAttrPair(resourceName, "plan_id", "data.sakura_workflows_plan.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "contract_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "activate_from"),
				),
			},
		},
	})
}

const testAccSakuraDataSourceWorkflowsSubscription_basic = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}

data "sakura_workflows_subscription" "foobar" {
  depends_on = [sakura_workflows_subscription.foobar]
}`
