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
	secret_manager "github.com/sacloud/terraform-provider-sakura/internal/service/s3cret_manager"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraSecretManagerSecret_basic(t *testing.T) {
	resourceName := "sakura_secret_manager_secret.foobar"
	rand := test.RandomName()

	var secret v1.Secret
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSecretManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecretManagerSecret_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecretManagerSecret_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "value", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
		},
	})
}

func TestAccSakuraSecretManagerSecret_basicWithWO(t *testing.T) {
	resourceName := "sakura_secret_manager_secret.foobar"
	rand := test.RandomName()

	var secret v1.Secret
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSecretManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecretManagerSecret_basicWithWO, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "value_wo_version", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "value"),
					resource.TestCheckNoResourceAttr(resourceName, "value_wo"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecretManagerSecret_updateWithWO, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "value_wo_version", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "value"),
					resource.TestCheckNoResourceAttr(resourceName, "value_wo"),
				),
			},
		},
	})
}

func testCheckSakuraSecretManagerSecretDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_secret_manager_secret" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		secretOp := sm.NewSecretOp(client.SecretManagerClient, rs.Primary.Attributes["vault_id"])

		_, err := secret_manager.FilterSecretManagerSecretByName(ctx, secretOp, rs.Primary.Attributes["name"])
		if err == nil {
			return fmt.Errorf("still exists SecretManagerSecret: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraSecretManagerSecretExists(n string, secret *v1.Secret) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no SecretManagerSecret vault ID is set")
		}

		client := test.AccClientGetter()
		ctx := context.Background()
		secretOp := sm.NewSecretOp(client.SecretManagerClient, rs.Primary.Attributes["vault_id"])

		foundSecret, err := secret_manager.FilterSecretManagerSecretByName(ctx, secretOp, rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		*secret = *foundSecret
		return nil
	}
}

//nolint:gosec
var testAccSakuraSecretManagerSecret_basic = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_secret_manager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}

resource "sakura_secret_manager_secret" "foobar" {
  name     = "{{ .arg0 }}"
  value    = "value1"
  vault_id = sakura_secret_manager.foobar.id

  depends_on = [sakura_secret_manager.foobar]
}`

//nolint:gosec
var testAccSakuraSecretManagerSecret_update = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_secret_manager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}

resource "sakura_secret_manager_secret" "foobar" {
  name     = "{{ .arg0 }}"
  value    = "value2"
  vault_id = sakura_secret_manager.foobar.id

  depends_on = [sakura_secret_manager.foobar]
}`

//nolint:gosec
var testAccSakuraSecretManagerSecret_basicWithWO = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_secret_manager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}

resource "sakura_secret_manager_secret" "foobar" {
  name     = "{{ .arg0 }}"
  vault_id = sakura_secret_manager.foobar.id
  value_wo = "value1"
  value_wo_version = 1

  depends_on = [sakura_secret_manager.foobar]
}`

//nolint:gosec
var testAccSakuraSecretManagerSecret_updateWithWO = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_secret_manager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}

resource "sakura_secret_manager_secret" "foobar" {
  name     = "{{ .arg0 }}"
  vault_id = sakura_secret_manager.foobar.id
  value_wo = "value2"
  value_wo_version = 2

  depends_on = [sakura_secret_manager.foobar]
}`
