// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package subnet_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraSubnet_basic(t *testing.T) {
	resourceName := "sakura_subnet.foobar"
	rand := test.RandomName()

	var subnet iaas.Subnet
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			test.CheckSakuraInternetDestroy,
			testCheckSakuraSubnetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSubnet_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "ip_addresses.#", "16"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSubnet_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "ip_addresses.#", "16"),
				),
			},
		},
	})
}

func testCheckSakuraSubnetExists(n string, subnet *iaas.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Subnet ID is set")
		}

		client := test.AccClientGetter()
		subnetOp := iaas.NewSubnetOp(client)
		zone := rs.Primary.Attributes["zone"]

		foundSubnet, err := subnetOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundSubnet.ID.String() != rs.Primary.ID {
			return fmt.Errorf("not found Subnet: %s", rs.Primary.ID)
		}

		*subnet = *foundSubnet
		return nil
	}
}

func testCheckSakuraSubnetDestroy(s *terraform.State) error {
	subnetOp := iaas.NewSubnetOp(test.AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_subnet" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := subnetOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("resource Subnet[%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}

var testAccSakuraSubnet_basic = `
resource sakura_internet "foobar" {
  name = "{{ .arg0 }}"
}
resource "sakura_subnet" "foobar" {
  internet_id = sakura_internet.foobar.id
  next_hop    = sakura_internet.foobar.min_ip_address
}`

var testAccSakuraSubnet_update = `
resource sakura_internet "foobar" {
  name = "{{ .arg0 }}"
}
resource "sakura_subnet" "foobar" {
  internet_id = sakura_internet.foobar.id
  next_hop    = sakura_internet.foobar.max_ip_address
}`
