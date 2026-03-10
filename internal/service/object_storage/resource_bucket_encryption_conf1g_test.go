// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraObjectStorageBucketEncryptionConfig_basic(t *testing.T) {
	resourceName := "sakura_object_storage_bucket_encryption_config.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(
		// testCheckSakuraObjectStorageBucketDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraObjectStorageBucketEncryptionConfig_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					// testCheckSakuraObjectStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", "tf-enc-config-"+rand),
					resource.TestCheckResourceAttrPair(
						resourceName, "site_id",
						"sakura_object_storage_bucket.foobar", "site_id",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "kms_key_id",
						"sakura_kms.foobar1", "id",
					),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraObjectStorageBucketEncryptionConfig_update, rand),
				Check: resource.ComposeTestCheckFunc(
					// testCheckSakuraObjectStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", "tf-enc-config-"+rand),
					resource.TestCheckResourceAttrPair(
						resourceName, "site_id",
						"sakura_object_storage_bucket.foobar", "site_id",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "kms_key_id",
						"sakura_kms.foobar2", "id",
					),
				),
			},
		},
	})
}

// TODO: implement check functions after update object-storage-api-go
// func testCheckSakuraObjectStorageBucketExists(n string) resource.TestCheckFunc {
// func testCheckSakuraObjectStorageBucketDestroy(s *terraform.State) error {

const testAccSakuraObjectStorageBucketEncryptionConfig_basic = `
resource "sakura_kms" "foobar1" {
  name        = "{{ .arg0 }}-1"
  description = "description-1"
  tags        = ["tag1"]
}

resource "sakura_kms" "foobar2" {
  name        = "{{ .arg0 }}-2"
  description = "description-2"
  tags        = ["tag2"]
}

data "sakura_object_storage_site" "foobar" {
  id = "tky01"
}

resource "sakura_object_storage_bucket" "foobar" {
  name    = "tf-enc-config-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_bucket_encryption_config" "foobar" {
  bucket = "tf-enc-config-{{ .arg0 }}"
  site_id = sakura_object_storage_bucket.foobar.site_id
  kms_key_id = sakura_kms.foobar1.id
}`

const testAccSakuraObjectStorageBucketEncryptionConfig_update = `
resource "sakura_kms" "foobar1" {
  name        = "{{ .arg0 }}-1"
  description = "description-1"
  tags        = ["tag1"]
}

resource "sakura_kms" "foobar2" {
  name        = "{{ .arg0 }}-2"
  description = "description-2"
  tags        = ["tag2"]
}

data "sakura_object_storage_site" "foobar" {
  id = "tky01"
}

resource "sakura_object_storage_bucket" "foobar" {
  name    = "tf-enc-config-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_bucket_encryption_config" "foobar" {
  bucket = "tf-enc-config-{{ .arg0 }}"
  site_id = sakura_object_storage_bucket.foobar.site_id
  kms_key_id = sakura_kms.foobar2.id
}`
