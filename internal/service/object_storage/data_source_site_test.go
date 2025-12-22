// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceObjectStorage_basic(t *testing.T) {
	resourceName := "data.sakura_object_storage_site.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSakuraDataSourceObjectStorageSite_basic,
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "id", "isk01"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "石狩第1サイト"),
					resource.TestCheckResourceAttr(resourceName, "region", "jp-north-1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", "isk01.sakurastorage.jp"),
					resource.TestCheckResourceAttr(resourceName, "s3_endpoint", "s3.isk01.sakurastorage.jp"),
					resource.TestCheckResourceAttr(resourceName, "status.status", "ok"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceObjectStorageSite_basic = `
data "sakura_object_storage_site" "foobar" { 
  id = "isk01"
}`
