// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceAPIGWSubscription_basic(t *testing.T) {
	// サブスクリプションが存在してないことを現状動的にチェックできないため、環境変数で指定する
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "sakura_apigw_subscription.foobar"
	rand := test.RandomName()
	var sub v1.Subscription
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraAPIGWSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWSubscription_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWSubscriptionExists(resourceName, &sub),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "monthly_request"),
					resource.TestCheckNoResourceAttr(resourceName, "service"),
					resource.TestCheckResourceAttrPair(resourceName, "plan_id", "data.sakura_apigw_plan.foobar", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWSubscription_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWSubscriptionExists(resourceName, &sub),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
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

func testCheckSakuraAPIGWSubscriptionDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apigw_subscription" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := getSubscriptionFromList(client.ApigwClient, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists APIGW Subscription: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraAPIGWSubscriptionExists(n string, sub *v1.Subscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no APIGW Group ID is set")
		}

		foundSub, err := getSubscriptionFromList(test.AccClientGetter().ApigwClient, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundSub.ID.Value.String() != rs.Primary.ID {
			return fmt.Errorf("not found APIGW Group: %s", rs.Primary.ID)
		}

		*sub = *foundSub
		return nil
	}
}

func getSubscriptionFromList(client *v1.Client, id string) (*v1.Subscription, error) {
	subOp := apigw.NewSubscriptionOp(client)
	subs, err := subOp.List(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list API Gateway subscriptions: %w", err)
	}

	for _, s := range subs {
		if s.ID.Value.String() == id {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("failed to find API Gateway subscription by ID: %s", id)
}

// used by other tests
var testSetupAPIGWSub = `
data "sakura_apigw_plan" "foobar" { 
  name = "エンタープライズ"
}

resource "sakura_apigw_subscription" "foobar" {
  name    = "{{ .arg0 }}"
  plan_id = data.sakura_apigw_plan.foobar.id
}
`

var testAccSakuraAPIGWSubscription_basic = `
data "sakura_apigw_plan" "foobar" { 
  name = "エンタープライズ"
}

resource "sakura_apigw_subscription" "foobar" {
  name = "{{ .arg0 }}"
  plan_id = data.sakura_apigw_plan.foobar.id
}`

var testAccSakuraAPIGWSubscription_update = `
data "sakura_apigw_plan" "foobar" { 
  name = "エンタープライズ"
}

resource "sakura_apigw_subscription" "foobar" {
  name = "{{ .arg0 }}-updated"
  plan_id = data.sakura_apigw_plan.foobar.id
}`
