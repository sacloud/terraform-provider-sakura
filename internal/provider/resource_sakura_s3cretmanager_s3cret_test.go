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

func TestAccSakuraSecretManagerSecret_basic(t *testing.T) {
	resourceName := "sakura_secretmanager_secret.foobar"
	rand := randomName()

	var secret v1.Secret
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraSecretManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildConfigWithArgs(testAccSakuraSecretManagerSecret_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: buildConfigWithArgs(testAccSakuraSecretManagerSecret_update, rand),
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

func testCheckSakuraSecretManagerSecretDestroy(s *terraform.State) error {
	client := testAccProvider.(*sakuraProvider).client
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_secretmanager_secret" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		secretOp := sm.NewSecretOp(client.secretmanagerClient, rs.Primary.Attributes["vault_id"])

		_, err := filterSecretManagerSecretByName(ctx, secretOp, rs.Primary.Attributes["name"])
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

		client := testAccProvider.(*sakuraProvider).client
		ctx := context.Background()
		secretOp := sm.NewSecretOp(client.secretmanagerClient, rs.Primary.Attributes["vault_id"])

		foundSecret, err := filterSecretManagerSecretByName(ctx, secretOp, rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if foundSecret.Name != rs.Primary.ID {
			return fmt.Errorf("not found SecretManagerSecret: %s", rs.Primary.ID)
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

resource "sakura_secretmanager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}

resource "sakura_secretmanager_secret" "foobar" {
  name     = "{{ .arg0 }}"
  value    = "value1"
  vault_id = sakura_secretmanager.foobar.id

  depends_on = [sakura_secretmanager.foobar]
}`

//nolint:gosec
var testAccSakuraSecretManagerSecret_update = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_secretmanager" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  kms_key_id  = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}

resource "sakura_secretmanager_secret" "foobar" {
  name     = "{{ .arg0 }}"
  value    = "value2"
  vault_id = sakura_secretmanager.foobar.id

  depends_on = [sakura_secretmanager.foobar]
}`
