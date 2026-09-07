// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	networkingsuite "github.com/sacloud/sacloud-sdk-go/api/networking-suite"
	v1 "github.com/sacloud/sacloud-sdk-go/api/networking-suite/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/srn"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraNetworkingSuiteSubnet_basic(t *testing.T) {
	resourceName := "sakura_networking_suite_subnet.foobar"
	rand := test.RandomName()
	region, zone := getRegionAndZone(t, os.Getenv("SAKURA_ENDPOINTS_NETWORKING_SUITE"))

	var dashboard v1.ReadSubnet
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraNetworkingSuiteSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNetworkingSuiteSubnet_basic, rand, region, zone),
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
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNetworkingSuiteSubnet_update, rand, region, zone),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraNetworkingSuiteSubnetExists(resourceName, &dashboard),
					resource.TestCheckResourceAttrSet(resourceName, "srn"),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_address_range", "10.0.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "zone", zone),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_group_srn", "sakura_networking_suite_subnet_group.foobar", "srn"),
				),
			},
		},
	})
}

func TestAccImportSakuraNetworkingSuiteSubnet_basic(t *testing.T) {
	rand := test.RandomName()
	region, zone := getRegionAndZone(t, os.Getenv("SAKURA_ENDPOINTS_NETWORKING_SUITE"))

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":               rand,
			"description":        "description",
			"ipv4_address_range": "10.0.0.0/24",
			"zone":               zone,
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0], "srn")
	}

	resourceName := "sakura_networking_suite_subnet.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraNetworkingSuiteSubnetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNetworkingSuiteSubnet_basic, rand, region, zone),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateCheck:                     checkFn,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "srn",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return rs.Primary.Attributes["srn"], nil
				},
			},
		},
	})
}

func testCheckSakuraNetworkingSuiteSubnetDestroy(s *terraform.State) error {
	client, err := networkingsuite.NewClient(test.AccClientGetter().SaClient2)
	if err != nil {
		return fmt.Errorf("Read: API Client Error: failed to create networking suite API client: %s", err)
	}
	op := networkingsuite.NewSubnetsOp(client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_networking_suite_subnet" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		parsed, _ := srn.Parse(rs.Primary.Attributes["srn"])
		_, err := op.Read(context.Background(), parsed)
		if err == nil {
			return fmt.Errorf("still exists networking suite subnet: %s", rs.Primary.Attributes["srn"])
		}
	}
	return nil
}

func testCheckSakuraNetworkingSuiteSubnetExists(n string, subnet *v1.ReadSubnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no subnet ID is set")
		}

		client, err := networkingsuite.NewClient(test.AccClientGetter().SaClient2)
		if err != nil {
			return fmt.Errorf("Read: API Client Error: failed to create networking suite API client: %s", err)
		}
		op := networkingsuite.NewSubnetsOp(client)
		parsed, _ := srn.Parse(rs.Primary.Attributes["srn"])
		found, err := op.Read(context.Background(), parsed)
		if err != nil {
			return err
		}

		if found.SRN != rs.Primary.Attributes["srn"] {
			return fmt.Errorf("not found subnet: %s", rs.Primary.Attributes["srn"])
		}

		*subnet = *found
		return nil
	}
}

var testAccSakuraNetworkingSuiteSubnet_basic = `
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
`

var testAccSakuraNetworkingSuiteSubnet_update = `
resource "sakura_networking_suite_subnet_group" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  ipv4_address_range = "10.0.0.0/20"
  region = "{{ .arg1 }}"
}

resource "sakura_networking_suite_subnet" "foobar" {
  name        = "{{ .arg0 }}-updated"
  description = "description-updated"
  ipv4_address_range = "10.0.0.0/24"
  zone = "{{ .arg2 }}"
  subnet_group_srn = sakura_networking_suite_subnet_group.foobar.srn
}
`
