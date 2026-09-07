// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	networkingsuite "github.com/sacloud/sacloud-sdk-go/api/networking-suite"
	v1 "github.com/sacloud/sacloud-sdk-go/api/networking-suite/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/common/packages/envvar"
	"github.com/sacloud/sacloud-sdk-go/srn"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraNetworkingSuiteSubnetGroup_basic(t *testing.T) {
	resourceName := "sakura_networking_suite_subnet_group.foobar"
	rand := test.RandomName()
	region, _ := getRegionAndZone(t, os.Getenv("SAKURA_ENDPOINTS_NETWORKING_SUITE"))

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
					resource.TestCheckResourceAttr(resourceName, "ipv4_address_range", "10.0.0.0/20"),
					resource.TestCheckResourceAttr(resourceName, "region", region),
					resource.TestCheckResourceAttrSet(resourceName, "zone"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraNetworkingSuiteSubnetGroup_update, rand, region),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraNetworkingSuiteSubnetGroupExists(resourceName, &dashboard),
					resource.TestCheckResourceAttrSet(resourceName, "srn"),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_address_range", "10.0.0.0/20"),
					resource.TestCheckResourceAttr(resourceName, "region", region),
					resource.TestCheckResourceAttrSet(resourceName, "zone"),
				),
			},
		},
	})
}

func TestAccImportSakuraNetworkingSuiteSubnetGroup_basic(t *testing.T) {
	rand := test.RandomName()
	region, _ := getRegionAndZone(t, os.Getenv("SAKURA_ENDPOINTS_NETWORKING_SUITE"))

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":               rand,
			"description":        "description",
			"ipv4_address_range": "10.0.0.0/20",
			"region":             region,
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0], "srn", "zone")
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

func getRegionAndZone(t *testing.T, urlStr string) (string, string) {
	var zone string
	if urlStr == "" {
		zone = envvar.StringFromEnv("SAKURA_ZONE", networkingsuite.DefaultZone)
	} else {
		u, err := url.ParseRequestURI(urlStr)
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}
		parts := strings.Split(u.Path, "/")
		if len(parts) < 4 {
			t.Fatalf("unexpected URL path format: %s", u.Path)
		}
		zone = parts[3]
	}
	return zone[:len(zone)-1], zone
}

var testAccSakuraNetworkingSuiteSubnetGroup_basic = `
resource "sakura_networking_suite_subnet_group" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  ipv4_address_range = "10.0.0.0/20"
  region = "{{ .arg1 }}"
}
`

var testAccSakuraNetworkingSuiteSubnetGroup_update = `
resource "sakura_networking_suite_subnet_group" "foobar" {
  name        = "{{ .arg0 }}-updated"
  description = "description-updated"
  ipv4_address_range = "10.0.0.0/20"
  region = "{{ .arg1 }}"
}
`
