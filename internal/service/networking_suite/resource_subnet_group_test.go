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
	"github.com/sacloud/sacloud-sdk-go/srn"
	networkingsuite "github.com/sacloud/terraform-provider-sakura/internal/ns-sdk"
	v1 "github.com/sacloud/terraform-provider-sakura/internal/ns-sdk/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraNetworkingSuiteSubnetGroup_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_NETWORKING_SUITE_REGION")

	resourceName := "sakura_networking_suite_subnet_group.foobar"
	rand := test.RandomName()
	region := os.Getenv("SAKURA_NETWORKING_SUITE_REGION")

	var dashboard v1.ReadSubnetGroup
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraNetworkingSuiteSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNetworkingSuiteSubnetGroup_basic, rand, region),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraNetworkingSuiteSubnetGroupExists(resourceName, &dashboard),
					resource.TestCheckResourceAttrSet(resourceName, "srn"),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_address_range_cidr", "10.0.0.0/20"),
					resource.TestCheckResourceAttr(resourceName, "region", region),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNetworkingSuiteSubnetGroup_update, rand, region),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraNetworkingSuiteSubnetGroupExists(resourceName, &dashboard),
					resource.TestCheckResourceAttrSet(resourceName, "srn"),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_address_range_cidr", "10.0.0.0/20"),
					resource.TestCheckResourceAttr(resourceName, "region", region),
				),
			},
		},
	})
}

func TestAccImportSakuraNetworkingSuiteSubnetGroup_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_NETWORKING_SUITE_REGION")

	rand := test.RandomName()
	region := os.Getenv("SAKURA_NETWORKING_SUITE_REGION")

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                    rand,
			"description":             "description",
			"ipv4_address_range_cidr": "10.0.0.0/20",
			"region":                  region,
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0], "srn")
	}

	resourceName := "sakura_networking_suite_subnet_group.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraNetworkingSuiteSubnetGroupDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNetworkingSuiteSubnetGroup_basic, rand, region),
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

func testCheckSakuraNetworkingSuiteSubnetGroupDestroy(s *terraform.State) error {
	client, err := networkingsuite.NewClient(test.AccClientGetter().SaClient2)
	if err != nil {
		return fmt.Errorf("Read: API Client Error: failed to create networking suite API client: %s", err)
	}
	op := networkingsuite.NewSubnetGroupsOp(client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_networking_suite_subnet_group" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		parsed, _ := srn.Parse(rs.Primary.Attributes["srn"])
		_, err := op.Read(context.Background(), parsed)
		if err == nil {
			return fmt.Errorf("still exists networking suite subnet group: %s", rs.Primary.Attributes["srn"])
		}
	}
	return nil
}

func testCheckSakuraNetworkingSuiteSubnetGroupExists(n string, sg *v1.ReadSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no subnet group ID is set")
		}

		client, err := networkingsuite.NewClient(test.AccClientGetter().SaClient2)
		if err != nil {
			return fmt.Errorf("Read: API Client Error: failed to create networking suite API client: %s", err)
		}
		op := networkingsuite.NewSubnetGroupsOp(client)
		parsed, _ := srn.Parse(rs.Primary.Attributes["srn"])
		found, err := op.Read(context.Background(), parsed)
		if err != nil {
			return err
		}

		if found.SRN != rs.Primary.Attributes["srn"] {
			return fmt.Errorf("not found subnet group: %s", rs.Primary.Attributes["srn"])
		}

		*sg = *found
		return nil
	}
}

var testAccSakuraNetworkingSuiteSubnetGroup_basic = `
resource "sakura_networking_suite_subnet_group" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  ipv4_address_range_cidr = "10.0.0.0/20"
  region = "{{ .arg1 }}"
}
`

var testAccSakuraNetworkingSuiteSubnetGroup_update = `
resource "sakura_networking_suite_subnet_group" "foobar" {
  name        = "{{ .arg0 }}-updated"
  description = "description-updated"
  ipv4_address_range_cidr = "10.0.0.0/20"
  region = "{{ .arg1 }}"
}
`
