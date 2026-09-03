// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraNetworkingSuiteInterfaceConnection_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_NETWORKING_SUITE_INTERFACE_SRN")

	resourceName := "sakura_networking_suite_interface_connection.foobar"
	rand := test.RandomName()
	region, zone := getRegionAndZone(t, os.Getenv("SAKURA_ENDPOINTS_NETWORKING_SUITE"))
	interfaceSRN := os.Getenv("SAKURA_NETWORKING_SUITE_INTERFACE_SRN")

	// Interface Connection doesn't have Read API, so we can't check destroy and existence of the resource.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNetworkingSuiteInterfaceConnection_basic, rand, region, zone, interfaceSRN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "srn"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_srn", "sakura_networking_suite_subnet.foobar", "srn"),
					resource.TestCheckResourceAttr(resourceName, "interface_srn", interfaceSRN),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_ipv4_address", "10.0.0.5"),
				),
			},
		},
	})
}

// Interface Connection doesn't support Import, so we don't have import test for it.

var testAccSakuraNetworkingSuiteInterfaceConnection_basic = `
resource "sakura_networking_suite_subnet_group" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  ipv4_address_range = "10.0.0.0/20"
  region = "{{ .arg1 }}"
}

resource "sakura_networking_suite_subnet" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  ipv4_address_range = "10.0.0.0/24"
  zone = "{{ .arg2 }}"
  subnet_group_srn = sakura_networking_suite_subnet_group.foobar.srn
}

resource "sakura_networking_suite_interface_connection" "foobar" {
  subnet_srn = sakura_networking_suite_subnet.foobar.srn
  interface_srn = "{{ .arg3 }}"
  ephemeral_ipv4_address = "10.0.0.5"
}
`
