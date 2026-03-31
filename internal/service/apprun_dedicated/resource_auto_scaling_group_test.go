// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	asg "github.com/sacloud/apprun-dedicated-api-go/apis/autoscalinggroup"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceApprunDedicatedAutoScalingGroup(t *testing.T) {
	resourceName := "sakura_apprun_dedicated_auto_scaling_group.main"
	name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		PreCheck:                 AccPreCheck(t),
		CheckDestroy:             testCheckSakuraApprunDedicatedAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedAutoScalingGroup_basic, name, globalClusterID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact("tfacc-"+name)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("zone"), knownvalue.StringExact("is1c")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("min_nodes"), knownvalue.Int32Exact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_nodes"), knownvalue.Int32Exact(3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("interfaces"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"interface_index":  knownvalue.Int32Exact(0),
							"upstream":         knownvalue.StringRegexp(regexp.MustCompile("^[0-9]+$")), // not "shared"
							"connects_to_lb":   knownvalue.Bool(false),
							"default_gateway":  knownvalue.NotNull(),
							"netmask":          knownvalue.NotNull(),
							"packet_filter_id": knownvalue.Null(),
							"ip_pool": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"start": knownvalue.NotNull(),
									"end":   knownvalue.NotNull(),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts", "current_nodes", "deleting"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					clusterID := rs.Primary.Attributes["cluster_id"]
					id := rs.Primary.Attributes["id"]
					return fmt.Sprintf("%s/%s", clusterID, id), nil
				},
			},
		},
	})
}

func testCheckSakuraApprunDedicatedAutoScalingGroupDestroy(s *terraform.State) error {
	client := test.AccClientGetter().AppRunDedicatedClient
	if client == nil {
		return errors.New("AppRunDedicatedClient is nil")
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apprun_dedicated_auto_scaling_group" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}
		clusterID, err := uuid.Parse(rs.Primary.Attributes["cluster_id"])
		if err != nil {
			return fmt.Errorf("invalid cluster ID: %s", rs.Primary.Attributes["cluster_id"])
		}
		asgID, err := uuid.Parse(rs.Primary.Attributes["id"])
		if err != nil {
			return fmt.Errorf("invalid auto scaling group ID: %s", rs.Primary.Attributes["id"])
		}
		api := asg.NewAutoScalingGroupOp(client, v1.ClusterID(clusterID))
		_, err = api.Read(context.Background(), v1.AutoScalingGroupID(asgID))
		if err == nil {
			return fmt.Errorf("auto scaling group still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

var testAccSakuraResourceApprunDedicatedAutoScalingGroup_basic = `
locals {
  sakura_dns = [ "133.242.0.3", "133.242.0.4" ]
}

data "sakura_zone" "is1c" {
  name = "is1c"
}

data "sakura_apprun_dedicated_worker_service_classes" "main" {}

resource "sakura_internet" "main" {
  name = "tfacc-{{ .arg0 }}"
  zone = data.sakura_zone.is1c.name
}

resource "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id                = "{{ .arg1 }}"
  name                      = "tfacc-{{ .arg0 }}"
  zone                      = data.sakura_zone.is1c.name
  worker_service_class_path = data.sakura_apprun_dedicated_worker_service_classes.main.classes[0].path
  name_servers              = local.sakura_dns
  min_nodes                 = 1
  max_nodes                 = 3
  interfaces = [
    {
      interface_index = 0
      upstream        = sakura_internet.main.vswitch_id
      connects_to_lb  = false
      netmask         = sakura_internet.main.netmask
      default_gateway = sakura_internet.main.gateway
      ip_pool = [
	    {
          start = sakura_internet.main.min_ip_address
          end   = sakura_internet.main.max_ip_address
        }
      ]
    }
  ]
}
`
