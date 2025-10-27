// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package kms_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/kms-api-go"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceKMS_basic(t *testing.T) {
	resourceName := "sakura_kms.foobar"
	rand := test.RandomName()
	var key v1.Key
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraKMSDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraKMS_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraKMSExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "key_origin", "generated"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraKMS_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraKMSExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					resource.TestCheckResourceAttr(resourceName, "key_origin", "generated"),
				),
			},
		},
	})
}

func TestAccSakuraResourceKMS_imported(t *testing.T) {
	resourceName := "sakura_kms.foobar2"
	rand := test.RandomName()

	var key v1.Key
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraKMSDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraKMS_imported, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraKMSExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description with plain key"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "key_origin", "imported"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraKMS_importedUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraKMSExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description with plain key updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "key_origin", "imported"),
				),
			},
		},
	})
}

func testCheckSakuraKMSDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	keyOp := kms.NewKeyOp(client.KmsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_kms" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := keyOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists KMS: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraKMSExists(n string, key *v1.Key) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no KMS ID is set")
		}

		client := test.AccClientGetter()
		keyOp := kms.NewKeyOp(client.KmsClient)

		foundKey, err := keyOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundKey.ID != rs.Primary.ID {
			return fmt.Errorf("not found KMS: %s", rs.Primary.ID)
		}

		*key = *foundKey
		return nil
	}
}

var testAccSakuraKMS_basic = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
}`

var testAccSakuraKMS_update = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag1-upd"]
}`

var testAccSakuraKMS_imported = `
resource "sakura_kms" "foobar2" {
  name        = "{{ .arg0 }}"
  description = "description with plain key"
  tags        = ["tag1", "tag2"]
  key_origin  = "imported"
  plain_key   = "AfL5zzjD4RgeFQm3vvAADwPNrurNUc616877wsa8v4w="
}`

var testAccSakuraKMS_importedUpdate = `
resource "sakura_kms" "foobar2" {
  name        = "{{ .arg0 }}"
  description = "description with plain key updated"
  tags        = ["tag1"]
  key_origin  = "imported"
}`
