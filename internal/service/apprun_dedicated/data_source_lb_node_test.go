// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedLoadBalancerNode(t *testing.T) {
	resourceName := "data.sakura_apprun_dedicated_lb_node.main"
	name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		PreCheck:                 AccPreCheck(t),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedLoadBalancerNodeConfig, name, globalClusterID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_scaling_group_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("lb_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("status"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("created"), knownvalue.NotNull()),
				},
			},
		},
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedLoadBalancerNodeConfig = `
locals {
  sakura_dns = [ "133.242.0.3", "133.242.0.4" ]
}

data "sakura_zone" "is1c" {
  name = "is1c"
}

data "sakura_apprun_dedicated_worker_service_classes" "main" {}

data "sakura_apprun_dedicated_lb_service_classes" "main" {}

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
  max_nodes                 = 1
  interfaces = [
    {
      interface_index = 0
      upstream        = "shared"
      connects_to_lb  = true
    }
  ]
}

resource "sakura_apprun_dedicated_lb" "main" {
  cluster_id                = sakura_apprun_dedicated_auto_scaling_group.main.cluster_id
  auto_scaling_group_id     = sakura_apprun_dedicated_auto_scaling_group.main.id
  name                      = "tfacc-{{ .arg0 }}"
  service_class_path        = data.sakura_apprun_dedicated_lb_service_classes.main.classes[0].path
  name_servers              = local.sakura_dns
  interfaces = [
    {
      interface_index   = 0
      upstream          = sakura_internet.main.vswitch_id
      ip_pool = [
        {
          start = sakura_internet.main.min_ip_address
          end   = sakura_internet.main.max_ip_address
        }
      ]
      netmask           = sakura_internet.main.netmask
      default_gateway   = sakura_internet.main.gateway
      vip               = cidrhost("${sakura_internet.main.gateway}/${sakura_internet.main.netmask}", 9)
      virtual_router_id = 1
    }
  ]
}

data "sakura_apprun_dedicated_lb_nodes" "nodes" {
  cluster_id            = sakura_apprun_dedicated_lb.main.cluster_id
  auto_scaling_group_id = sakura_apprun_dedicated_lb.main.auto_scaling_group_id
  lb_id      = sakura_apprun_dedicated_lb.main.id
}

data "sakura_apprun_dedicated_lb_node" "main" {
  cluster_id            = sakura_apprun_dedicated_lb.main.cluster_id
  auto_scaling_group_id = sakura_apprun_dedicated_lb.main.auto_scaling_group_id
  lb_id      = sakura_apprun_dedicated_lb.main.id
  id                    = data.sakura_apprun_dedicated_lb_nodes.nodes.nodes[0].id
}
`
