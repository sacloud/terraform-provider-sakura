// Copyright 2016-2026 The terraform-provider-sakura Authors
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

func TestAccSakuraSimpleNotificationGroup_basic(t *testing.T) {
	resourceName := "sakura_simple_notification_group.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSimpleNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSimpleNotificationGroup_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "destinations.0", "sakura_simple_notification_destination.foobar_1", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "destinations.1", "sakura_simple_notification_destination.foobar_2", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSimpleNotificationGroup_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSimpleNotificationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "destinations.0", "sakura_simple_notification_destination.foobar_1", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "destinations.1", "sakura_simple_notification_destination.foobar_2", "id"),
				),
			},
		},
	})
}

func testCheckSakuraSimpleNotificationDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	destOp := simple_notification.NewGroupOp(client.SimpleNotificationClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_simple_notification_group" {
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

func testCheckSakuraSimpleNotificationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no SimpleNotification ID is set")
		}

		client := test.AccClientGetter()
		destOp := simple_notification.NewGroupOp(client.SimpleNotificationClient)

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

var testAccSakuraSimpleNotificationGroup_basic = `
resource "sakura_simple_notification_destination" "foobar_1" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
  type        = "email"
  value       = "hoge@hogehoge.com"
}

resource "sakura_simple_notification_destination" "foobar_2" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
  type        = "email"
  value       = "hoge@hogehoge.com"
}

resource "sakura_simple_notification_group" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
  destinations = [sakura_simple_notification_destination.foobar_1.id, sakura_simple_notification_destination.foobar_2.id]
}`

var testAccSakuraSimpleNotificationGroup_update = `

resource "sakura_simple_notification_destination" "foobar_1" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
  type        = "email"
  value       = "hoge@hogehoge.com"
}

resource "sakura_simple_notification_destination" "foobar_2" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
  type        = "email"
  value       = "hoge@hogehoge.com"
}

resource "sakura_simple_notification_group" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag1-upd"]
  destinations = [sakura_simple_notification_destination.foobar_1.id, sakura_simple_notification_destination.foobar_2.id]
}`
