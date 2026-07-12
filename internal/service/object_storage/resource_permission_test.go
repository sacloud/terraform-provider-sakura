// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	objectstorage "github.com/sacloud/object-storage-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraObjectStoragePermission_basic(t *testing.T) {
	resourceName := "sakura_object_storage_permission.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraObjectStoragePermissionDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraObjectStoragePermission_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraObjectStoragePermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-permission-test-"+rand),
					resource.TestCheckResourceAttrPair(resourceName, "site_id", "data.sakura_object_storage_site.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "access_key"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.0.bucket", "tf-permission-test-tky1-"+rand),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.0.can_read", "true"),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.0.can_write", "true"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraObjectStoragePermission_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraObjectStoragePermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-permission-test-"+rand),
					resource.TestCheckResourceAttrPair(resourceName, "site_id", "data.sakura_object_storage_site.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "access_key"),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.0.bucket", "tf-permission-test-tky1-"+rand),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.0.can_read", "true"),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.0.can_write", "true"),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.1.bucket", "tf-permission-test-tky2-"+rand),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.1.can_read", "true"),
					resource.TestCheckResourceAttr(resourceName, "bucket_controls.1.can_write", "false"),
				),
			},
		},
	})
}

func TestAccImportSakuraObjectStoragePermission_basic(t *testing.T) {
	rand := test.RandomName()

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                        "tf-permission-test-" + rand,
			"site_id":                     "tky01",
			"bucket_controls.#":           "1",
			"bucket_controls.0.bucket":    "tf-permission-test-tky1-" + rand,
			"bucket_controls.0.can_read":  "true",
			"bucket_controls.0.can_write": "true",
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0])
	}

	resourceName := "sakura_object_storage_permission.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraObjectStoragePermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraObjectStoragePermission_import, rand),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateCheck:        checkFn,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_key", "secret_key"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["site_id"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func testCheckSakuraObjectStoragePermissionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Object Storage Permission ID is set")
		}

		client := test.AccClientGetter()
		siteClient, err := objectstorage.NewSiteClient(client.SaClient, rs.Primary.Attributes["site_id"])
		if err != nil {
			return fmt.Errorf("failed to create Object Storage Site client: %w", err)
		}

		permissionOp := objectstorage.NewPermissionOp(siteClient)
		permission, err := permissionOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if utils.ItoA(permission.ID.Value) != rs.Primary.ID {
			return fmt.Errorf("resource Object Storage Permission[%s] not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckSakuraObjectStoragePermissionDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_object_storage_permission" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		siteClient, err := objectstorage.NewSiteClient(client.SaClient, rs.Primary.Attributes["site_id"])
		if err != nil {
			return fmt.Errorf("failed to create Object Storage Site client: %w", err)
		}

		permissionOp := objectstorage.NewPermissionOp(siteClient)
		_, err = permissionOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("resource Object Storage Permission[%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraObjectStoragePermission_basic = `
data "sakura_object_storage_site" "foobar" {
  id = "tky01"
}

resource "sakura_object_storage_bucket" "foobar1" {
  name    = "tf-permission-test-tky1-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_bucket" "foobar2" {
  name    = "tf-permission-test-tky2-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_permission" "foobar" {
  name = "tf-permission-test-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
  bucket_controls = [{
    bucket = sakura_object_storage_bucket.foobar1.name
    can_read = true
    can_write = true
  }]
}`

const testAccSakuraObjectStoragePermission_update = `
data "sakura_object_storage_site" "foobar" {
  id = "tky01"
}

resource "sakura_object_storage_bucket" "foobar1" {
  name    = "tf-permission-test-tky1-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_bucket" "foobar2" {
  name    = "tf-permission-test-tky2-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_permission" "foobar" {
  name = "tf-permission-test-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
  bucket_controls = [{
    bucket = sakura_object_storage_bucket.foobar1.name
    can_read = true
    can_write = true
  },
  {
    bucket = sakura_object_storage_bucket.foobar2.name
    can_read = true
    can_write = false
  }]
}`

const testAccSakuraObjectStoragePermission_import = `
data "sakura_object_storage_site" "foobar" {
  id = "tky01"
}

resource "sakura_object_storage_bucket" "foobar1" {
  name    = "tf-permission-test-tky1-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
}

resource "sakura_object_storage_permission" "foobar" {
  name = "tf-permission-test-{{ .arg0 }}"
  site_id = data.sakura_object_storage_site.foobar.id
  bucket_controls = [{
    bucket = sakura_object_storage_bucket.foobar1.name
    can_read = true
    can_write = true
  }]
}`
