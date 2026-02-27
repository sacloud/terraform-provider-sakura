// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceSimpleNotificationGroup_basic(t *testing.T) {
	resourceName := "data.sakura_simple_notification_group.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSimpleNotificationGroup, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destinations.0", "sakura_simple_notification_destination.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceSimpleNotificationGroup = `
resource "sakura_simple_notification_destination" "foobar" {
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
  destinations = [sakura_simple_notification_destination.foobar.id]
  depends_on = [sakura_simple_notification_destination.foobar]
}

data "sakura_simple_notification_group" "foobar" {
  name = sakura_simple_notification_group.foobar.name
}
`
