// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceSimpleNotificationRouting_basic(t *testing.T) {
	resourceName := "data.sakura_simple_notification_routing.foobar"
	rand := test.RandomName()
	randSourceID := test.RandStringFromCharSet(11, "123456789")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSimpleNotificationRouting, rand, randSourceID),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.0.name", "name1"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.0.value", "value1"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.1.name", "name2"),
					resource.TestCheckResourceAttr(resourceName, "match_labels.1.value", "value2"),
					resource.TestCheckResourceAttr(resourceName, "source_id", randSourceID),
					resource.TestCheckResourceAttrPair(resourceName, "target_group_id", "sakura_simple_notification_group.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceSimpleNotificationRouting = `
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

resource "sakura_simple_notification_routing" "foobar" {
  name            = "{{ .arg0 }}"
  description     = "description"
  tags            = ["tag1", "tag2"]
  match_labels    = [
    { 
      name : "name1",
      value : "value1"
    }, 
    { 
      name : "name2",
      value : "value2"
    } 
  ]
  source_id       = "{{ .arg1 }}" 
  target_group_id = sakura_simple_notification_group.foobar.id
  
  depends_on = [
	sakura_simple_notification_group.foobar,
  ]
}

data "sakura_simple_notification_group" "foobar" {
  name = sakura_simple_notification_group.foobar.name
}

data "sakura_simple_notification_routing" "foobar" {
  name = sakura_simple_notification_routing.foobar.name
}

`
