// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v1 "github.com/sacloud/sacloud-sdk-go/api/networking-suite/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceNetworkingSuiteSubnet_basic(t *testing.T) {
	resourceName := "data.sakura_networking_suite_subnet.foobar"
	rand := test.RandomName()
	region, zone := getRegionAndZone(t, os.Getenv("SAKURA_ENDPOINTS_NETWORKING_SUITE"))

	var dashboard v1.ReadSubnet
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraNetworkingSuiteSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceNetworkingSuiteSubnet_basic, rand, region, zone),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraNetworkingSuiteSubnetExists(resourceName, &dashboard),
					resource.TestCheckResourceAttrSet(resourceName, "srn"),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_address_range", "10.0.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "zone", zone),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_group_srn", "sakura_networking_suite_subnet_group.foobar", "srn"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceNetworkingSuiteSubnet_basic = `
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

data "sakura_networking_suite_subnet" "foobar" {
  srn = sakura_networking_suite_subnet.foobar.srn
}
`
