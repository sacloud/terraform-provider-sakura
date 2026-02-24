// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

func TestAccSakuraResourceWorkflowsSubscription_basic(t *testing.T) {
	resourceName := "sakura_workflows_subscription.foobar"
	var subscription v1.GetSubscriptionOK

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSakuraWorkflowsSubscription_basic,
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsSubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "plan_name", "200Kプラン"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				Config: testAccSakuraWorkflowsSubscription_update,
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsSubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "plan_name", "600Kプラン"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func testCheckSakuraWorkflowsSubscriptionDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	subscriptionOp := workflows.NewSubscriptionOp(client.WorkflowsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_workflows_subscription" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := subscriptionOp.Read(context.Background())
		if err == nil {
			return fmt.Errorf("subscription still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraWorkflowsSubscriptionExists(n string, subscription *v1.GetSubscriptionOK) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no subscription ID is set")
		}

		client := test.AccClientGetter()
		subscriptionOp := workflows.NewSubscriptionOp(client.WorkflowsClient)

		foundSubscription, err := subscriptionOp.Read(context.Background())
		if err != nil {
			return err
		}

		*subscription = *foundSubscription
		return nil
	}
}

const testAccSakuraWorkflowsSubscription_basic = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}`

const testAccSakuraWorkflowsSubscription_update = `
data "sakura_workflows_plan" "foobar" {
  name = "600K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}`
