// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraObjectStorageBucketReplicationConfig_basic(t *testing.T) {
	resourceName := "sakura_object_storage_bucket_replication_config.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(
		// testCheckSakuraObjectStorageBucketDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraObjectStorageBucketReplicationConfig_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					// testCheckSakuraObjectStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", "tf-rep-config-isk-"+rand),
					resource.TestCheckResourceAttr(resourceName, "destination_bucket", "tf-rep-config-tky-"+rand),
					resource.TestCheckResourceAttrPair(
						resourceName, "site_id",
						"sakura_object_storage_bucket.isk", "site_id",
					),
				),
			},
		},
	})
}

const testAccSakuraObjectStorageBucketReplicationConfig_basic = `
data "sakura_object_storage_site" "isk" {
  id = "isk01"
}

data "sakura_object_storage_site" "tky" {
  id = "tky01"
}

resource "sakura_object_storage_bucket" "isk" {
  name    = "tf-rep-config-isk-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.isk.id
}

resource "sakura_object_storage_bucket" "tky" {
  name    = "tf-rep-config-tky-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.tky.id
}

resource "sakura_object_storage_bucket_replication_config" "foobar" {
  site_id = data.sakura_object_storage_site.isk.id
  bucket = sakura_object_storage_bucket.isk.name
  destination_bucket = sakura_object_storage_bucket.tky.name
}`
