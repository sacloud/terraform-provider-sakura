// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package seg_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceSEG_basic(t *testing.T) {
	resourceName := "data.sakura_seg.foobar"
	rand := test.RandomName()
	objectStorageEndpointDatasource := "s3.isk01.sakurastorage.jp"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSEGBasic, rand, objectStorageEndpointDatasource),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "vswitch_id", "sakura_vswitch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "zone", "tk1b"),
					resource.TestCheckResourceAttr(resourceName, "server_ip_addresses.0", "192.168.100.10"),
					resource.TestCheckResourceAttr(resourceName, "netmask", "28"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.object_storage_endpoints.0", objectStorageEndpointDatasource),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.simple_ai_endpoints.0", "simpleai.is1.api.sacloud.jp"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceSEGBasic = `
resource "sakura_vswitch" "foobar" {
	name = "{{ .arg0 }}"
	zone = "tk1b"
}
resource "sakura_seg" "foobar" {
	zone        = "tk1b"
	vswitch_id  = sakura_vswitch.foobar.id
	server_ip_addresses = ["192.168.100.10"]
	netmask     = 28
	endpoint_setting = {
		object_storage_endpoints = ["{{ .arg1 }}"]
		simple_ai_endpoints = ["simpleai.is1.api.sacloud.jp"]
	}
}
data "sakura_seg" "foobar" {
	id = sakura_seg.foobar.id
	zone = "tk1b"
}`
