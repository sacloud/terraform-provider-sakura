// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package packet_filter_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraPacketFilterRules_basic(t *testing.T) {
	resourceName := "sakura_packet_filter_rules.rules"
	rand := test.RandomName()

	var filter iaas.PacketFilter
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraPacketFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraPacketFilterRules_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraPacketFilterExists("sakura_packet_filter.foobar", &filter),
					resource.TestCheckResourceAttr(resourceName, "expression.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.source_network", "192.168.2.0"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.source_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.destination_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.allow", "true"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.description", "description"),

					resource.TestCheckResourceAttr(resourceName, "expression.1.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "expression.1.source_network", "192.168.2.0"),
					resource.TestCheckResourceAttr(resourceName, "expression.1.source_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "expression.1.destination_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "expression.1.allow", "true"),

					resource.TestCheckResourceAttr(resourceName, "expression.2.protocol", "ip"),
					resource.TestCheckResourceAttr(resourceName, "expression.2.source_network", ""),
					resource.TestCheckResourceAttr(resourceName, "expression.2.source_port", ""),
					resource.TestCheckResourceAttr(resourceName, "expression.2.destination_port", ""),
					resource.TestCheckResourceAttr(resourceName, "expression.2.allow", "false"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraPacketFilterRules_update, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "expression.0.protocol", "udp"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.source_network", "192.168.2.2"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.source_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.destination_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.allow", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "expression.0.description"),

					resource.TestCheckResourceAttr(resourceName, "expression.1.protocol", "udp"),
					resource.TestCheckResourceAttr(resourceName, "expression.1.source_network", "192.168.2.2"),
					resource.TestCheckResourceAttr(resourceName, "expression.1.source_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "expression.1.destination_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "expression.1.allow", "true"),

					resource.TestCheckResourceAttr(resourceName, "expression.2.protocol", "icmp"),
					resource.TestCheckResourceAttr(resourceName, "expression.2.source_network", "0.0.0.0"),
					resource.TestCheckResourceAttr(resourceName, "expression.2.source_port", ""),
					resource.TestCheckResourceAttr(resourceName, "expression.2.destination_port", ""),
					resource.TestCheckResourceAttr(resourceName, "expression.2.allow", "true"),

					resource.TestCheckResourceAttr(resourceName, "expression.3.protocol", "ip"),
					resource.TestCheckResourceAttr(resourceName, "expression.3.source_network", ""),
					resource.TestCheckResourceAttr(resourceName, "expression.3.source_port", ""),
					resource.TestCheckResourceAttr(resourceName, "expression.3.destination_port", ""),
					resource.TestCheckResourceAttr(resourceName, "expression.3.allow", "false"),
				),
			},
		},
	})
}

var testAccSakuraPacketFilterRules_basic = `
resource "sakura_packet_filter" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_packet_filter_rules" "rules" {
  packet_filter_id = sakura_packet_filter.foobar.id
  expression = [{
 	protocol         = "tcp"
	source_network   = "192.168.2.0"
	source_port      = "80"
	destination_port = "80"
	allow            = true
    description      = "description"
  },
  {
	protocol         = "tcp"
	source_network   = "192.168.2.0"
	source_port      = "443"
	destination_port = "443"
	allow            = true
  },
  {
	protocol = "ip"
	allow    = false
  }]
}
`

var testAccSakuraPacketFilterRules_update = `
resource "sakura_packet_filter" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_packet_filter_rules" "rules" {
  packet_filter_id = sakura_packet_filter.foobar.id
  expression = [{
   	protocol         = "udp"
  	source_network   = "192.168.2.2"
  	source_port      = "80"
  	destination_port = "80"
   	allow            = true
  },
  {
   	protocol         = "udp"
  	source_network   = "192.168.2.2"
  	source_port      = "443"
  	destination_port = "443"
  	allow            = true
  },
  {
    protocol       = "icmp"
    source_network = "0.0.0.0"
    allow          = true
  },
  {
  	protocol = "ip"
	allow    = false
  }]
}
`
