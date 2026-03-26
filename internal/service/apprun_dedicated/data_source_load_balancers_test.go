// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedLoadBalancers(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_ENABLE_APPRUN_DEDICATED_TEST")
	test.SkipIfEnvIsNotSet(t, "SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")
	test.SkipIfFakeModeEnabled(t)

	spid := os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")
	if spid == "" {
		t.Fatalf("need valid SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID environment variable")
	}

	resourceName := "data.sakura_apprun_dedicated_load_balancers.main"
	name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedLoadBalancersConfig, name, spid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "cluster_id"),
					resource.TestCheckResourceAttrSet(resourceName, "auto_scaling_group_id"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancers.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancers.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancers.0.service_class_path"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancers.0.created"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedLoadBalancersConfig = `
locals {
  sakura_dns = [ "133.242.0.3", "133.242.0.4" ]
}

data "sakura_zone" "is1c" {
  name = "is1c"
}

data "sakura_apprun_dedicated_worker_service_classes" "main" {}

data "sakura_apprun_dedicated_load_balancer_service_classes" "main" {}

resource "sakura_apprun_dedicated_cluster" "main" {
  name                 = "tfacc-{{ .arg0 }}"
  service_principal_id = "{{ .arg1 }}"
}

resource "sakura_internet" "main" {
  name = "tfacc-{{ .arg0 }}"
  zone = data.sakura_zone.is1c.name
}

resource "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id                = sakura_apprun_dedicated_cluster.main.id
  name                      = "tfacc-{{ .arg0 }}"
  zone                      = data.sakura_zone.is1c.name
  worker_service_class_path = data.sakura_apprun_dedicated_worker_service_classes.main.classes[0].path
  name_servers              = local.sakura_dns
  min_nodes                 = 1
  max_nodes                 = 3
  interfaces = [
    {
      interface_index = 0
      upstream        = "shared"
      connects_to_lb  = true
    }
  ]
}

resource "sakura_apprun_dedicated_load_balancer" "main" {
  cluster_id                = sakura_apprun_dedicated_auto_scaling_group.main.cluster_id
  auto_scaling_group_id     = sakura_apprun_dedicated_auto_scaling_group.main.id
  name                      = "tfacc-{{ .arg0 }}"
  service_class_path        = data.sakura_apprun_dedicated_load_balancer_service_classes.main.classes[0].path
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
      netmask_len       = sakura_internet.main.netmask
      default_gateway   = sakura_internet.main.gateway
      vip               = cidrhost("${sakura_internet.main.gateway}/${sakura_internet.main.netmask}", 9)
      virtual_router_id = 1
    }
  ]
}

data "sakura_apprun_dedicated_load_balancers" "main" {
  cluster_id            = sakura_apprun_dedicated_auto_scaling_group.main.cluster_id
  auto_scaling_group_id = sakura_apprun_dedicated_auto_scaling_group.main.id
  depends_on            = [sakura_apprun_dedicated_load_balancer.main]
}
`
