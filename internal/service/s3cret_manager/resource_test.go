// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package secret_manager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraSecretManager_basic(t *testing.T) {
	resourceName := "sakura_secret_manager.foobar"
	rand := test.RandomName()

	var vault v1.Vault
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSecretManagerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecretManager_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "sakura_kms.foobar", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecretManager_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "sakura_kms.foobar", "id"),
				),
			},
		},
	})
}

func testCheckSakuraSecretManagerDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	vaultOp := sm.NewVaultOp(client.SecretManagerClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_secret_manager" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := vaultOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists SecretManager: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraSecretManagerExists(n string, vault *v1.Vault) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no SecretManager vault ID is set")
		}

		client := test.AccClientGetter()
		vaultOp := sm.NewVaultOp(client.SecretManagerClient)

		foundVault, err := vaultOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundVault.ID != rs.Primary.ID {
			return fmt.Errorf("not found SecretManager: %s", rs.Primary.ID)
		}

		*vault = *foundVault
		return nil
	}
}

//nolint:gosec
var testAccSakuraSecretManager_basic = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
}

resource "sakura_secret_manager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}`

//nolint:gosec
var testAccSakuraSecretManager_update = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
}

resource "sakura_secret_manager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag1-upd"]
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}`
