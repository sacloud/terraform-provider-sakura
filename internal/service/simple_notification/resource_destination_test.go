// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	simple_notification "github.com/sacloud/simple-notification-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraSimpleNotificationDestination_basic(t *testing.T) {
	resourceName := "sakura_simple_notification_destination.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSimpleNotificationDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSimpleNotificationDestination_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "icon_id", "112901627732"),
					resource.TestCheckResourceAttr(resourceName, "type", "email"),
					resource.TestCheckResourceAttr(resourceName, "value", "hoge@hogehoge.com"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSimpleNotificationDestination_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSimpleNotificationDestinationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					resource.TestCheckResourceAttr(resourceName, "icon_id", "112901627732"),
					resource.TestCheckResourceAttr(resourceName, "type", "email"),
					resource.TestCheckResourceAttr(resourceName, "value", "hoge2@hogehoge.com"),
				),
			},
		},
	})
}

func testCheckSakuraSimpleNotificationDestinationDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	destOp := simple_notification.NewDestinationOp(client.SimpleNotificationClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_simple_notification_destination" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := destOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists SimpleNotification: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraSimpleNotificationDestinationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no SimpleNotification ID is set")
		}

		client := test.AccClientGetter()
		destOp := simple_notification.NewDestinationOp(client.SimpleNotificationClient)

		res, err := destOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}
		if res.CommonServiceItem.ID != rs.Primary.ID {
			return fmt.Errorf("not found SimpleNotification: %s", rs.Primary.ID)
		}
		return nil
	}
}

var testAccSakuraSimpleNotificationDestination_basic = `
resource "sakura_simple_notification_destination" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
  icon_id     = "112901627732"
  type        = "email"
  value       = "hoge@hogehoge.com"
}`

var testAccSakuraSimpleNotificationDestination_update = `
resource "sakura_simple_notification_destination" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag1-upd"]
  icon_id     = "112901627732"
  type        = "email"
  value       = "hoge2@hogehoge.com"
}`
