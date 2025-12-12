// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package packet_filter_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourcePacketFilter_basic(t *testing.T) {
	resourceName := "data.sakura_packet_filter.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourcePacketFilter_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "expression.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.source_network", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.source_port", "0-65535"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.destination_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.allow", "true"),
				),
			},
		},
	})
}

var testAccSakuraDataSourcePacketFilter_basic = `
resource "sakura_packet_filter" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_packet_filter_rules" "foobar" {
  packet_filter_id = sakura_packet_filter.foobar.id
  expression = [{
    protocol         = "tcp"
    source_network   = "0.0.0.0/0"
    source_port      = "0-65535"
    destination_port = "80"
    allow            = true
  },
  {
    protocol         = "udp"
    source_network   = "0.0.0.0/0"
    source_port      = "0-65535"
    destination_port = "80"
    allow            = true
  }]
}

data "sakura_packet_filter" "foobar" {
  name = sakura_packet_filter.foobar.name

  depends_on = [sakura_packet_filter_rules.foobar]
}`
