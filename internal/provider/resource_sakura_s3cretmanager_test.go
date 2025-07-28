// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sakura

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

func TestAccSakuraSecretManager_basic(t *testing.T) {
	resourceName := "sakura_secretmanager.foobar"
	rand := randomName()

	var vault v1.Vault
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSecretManagerDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildConfigWithArgs(testAccSakuraSecretManager_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					// 綺麗に動的にkms_key_idをテストで取得する方法があればコメントアウト
					// resource.TestCheckResourceAttr(resourceName, "kms_key_id", vault.KmsKeyID),
				),
			},
			{
				Config: buildConfigWithArgs(testAccSakuraSecretManager_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					// resource.TestCheckResourceAttr(resourceName, "kms_key_id", vault.KmsKeyID),
				),
			},
		},
	})
}

func testCheckSakuraSecretManagerDestroy(s *terraform.State) error {
	client := testAccProvider.(*sakuraProvider).client
	vaultOp := sm.NewVaultOp(client.secretmanagerClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_secretmanager" {
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

		client := testAccProvider.(*sakuraProvider).client
		vaultOp := sm.NewVaultOp(client.secretmanagerClient)

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

resource "sakura_secretmanager" "foobar" {
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

resource "sakura_secretmanager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag1-upd"]
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}`
