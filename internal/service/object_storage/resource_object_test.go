// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraObjectStorageObject_basic(t *testing.T) {
	resourceName := "sakura_object_storage_object.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(
		//testCheckSakuraObjectStorageObjectDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraObjectStorageObject_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					//testCheckSakuraObjectStorageObjectExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rand+"-object"),
					resource.TestCheckResourceAttr(resourceName, "content", "Hello"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/plain"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraObjectStorageObject_update, rand),
				Check: resource.ComposeTestCheckFunc(
					//testCheckSakuraObjectStorageObjectExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rand+"-object"),
					resource.TestCheckResourceAttr(resourceName, "content", "Hello World"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "text/plain"),
				),
			},
		},
	})
}

// TODO: implement check functions after update object-storage-api-go
// func testCheckSakuraObjectStorageObjectExists(n string) resource.TestCheckFunc {
// func testCheckSakuraObjectStorageObjectDestroy(s *terraform.State) error {

const testAccSakuraObjectStorageObject_basic = `
data "sakura_object_storage_site" "foobar" {
  id = "tky01"
}

resource "sakura_object_storage_bucket" "foobar" {
  name    = "{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_permission" "foobar" {
  name = "{{ .arg0 }}"
  site_id = sakura_object_storage_bucket.foobar.site_id
  bucket_controls = [{
    bucket = sakura_object_storage_bucket.foobar.name
    can_read = true
    can_write = true
  }]
}

resource "sakura_object_storage_object" "foobar" {
  region  = data.sakura_object_storage_site.foobar.region
  endpoint = data.sakura_object_storage_site.foobar.s3_endpoint
  bucket  = sakura_object_storage_bucket.foobar.name
  access_key = sakura_object_storage_permission.foobar.access_key
  secret_key = sakura_object_storage_permission.foobar.secret_key
  key     = "{{ .arg0 }}-object"
  content = "Hello"
  content_type = "text/plain"
}
`

const testAccSakuraObjectStorageObject_update = `
data "sakura_object_storage_site" "foobar" {
  id = "tky01"
}

resource "sakura_object_storage_bucket" "foobar" {
  name    = "{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_permission" "foobar" {
  name = "{{ .arg0 }}"
  site_id = sakura_object_storage_bucket.foobar.site_id
  bucket_controls = [{
    bucket = sakura_object_storage_bucket.foobar.name
    can_read = true
    can_write = true
  }]
}

resource "sakura_object_storage_object" "foobar" {
  region  = data.sakura_object_storage_site.foobar.region
  endpoint = data.sakura_object_storage_site.foobar.s3_endpoint
  bucket  = sakura_object_storage_bucket.foobar.name
  access_key = sakura_object_storage_permission.foobar.access_key
  secret_key = sakura_object_storage_permission.foobar.secret_key
  key     = "{{ .arg0 }}-object"
  content = "Hello World"
  content_type = "text/plain"
}
`
