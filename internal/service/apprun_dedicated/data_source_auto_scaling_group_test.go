// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedAutoScalingGroup(t *testing.T) {
	t.Run("find by id", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_auto_scaling_group.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedAutoScalingGroupConfigById, name, globalClusterID),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttrSet(resourceName, "cluster_id"),
						resource.TestCheckResourceAttr(resourceName, "name", "tfacc-"+name),
						resource.TestCheckResourceAttr(resourceName, "zone", "is1c"),
						resource.TestCheckResourceAttr(resourceName, "min_nodes", "1"),
						resource.TestCheckResourceAttr(resourceName, "max_nodes", "3"),
					),
				},
			},
		})
	})

	t.Run("find by name", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_auto_scaling_group.by_name"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedAutoScalingGroupConfigByName, name, globalClusterID),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttrSet(resourceName, "cluster_id"),
						resource.TestCheckResourceAttr(resourceName, "name", "tfacc-"+name),
						resource.TestCheckResourceAttr(resourceName, "zone", "is1c"),
					),
				},
			},
		})
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedAutoScalingGroupSetup = `
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
      netmask_len     = sakura_internet.main.netmask
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

var testAccCheckSakuraDataSourceApprunDedicatedAutoScalingGroupConfigById = testAccCheckSakuraDataSourceApprunDedicatedAutoScalingGroupSetup + `
data "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id = sakura_apprun_dedicated_auto_scaling_group.main.cluster_id
  id         = sakura_apprun_dedicated_auto_scaling_group.main.id
}
`

var testAccCheckSakuraDataSourceApprunDedicatedAutoScalingGroupConfigByName = testAccCheckSakuraDataSourceApprunDedicatedAutoScalingGroupSetup + `
data "sakura_apprun_dedicated_auto_scaling_group" "by_name" {
  cluster_id = sakura_apprun_dedicated_auto_scaling_group.main.cluster_id
  name       = sakura_apprun_dedicated_auto_scaling_group.main.name
}
`
