// Copyright 2016-2025 terraform-provider-sakura authors
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
	"github.com/sacloud/kms-api-go"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
)

func TestAccSakuraResourceKMS_basic(t *testing.T) {
	resourceName := "sakura_kms.foobar"
	rand := randomName()
	var key v1.Key
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraKMSDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildConfigWithArgs(testAccSakuraKMS_basic, rand),
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
				Config: buildConfigWithArgs(testAccSakuraKMS_update, rand),
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
	rand := randomName()

	var key v1.Key
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraKMSDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildConfigWithArgs(testAccSakuraKMS_imported, rand),
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
				Config: buildConfigWithArgs(testAccSakuraKMS_importedUpdate, rand),
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
	client := testAccProvider.(*sakuraProvider).client
	keyOp := kms.NewKeyOp(client.kmsClient)

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

		client := testAccProvider.(*sakuraProvider).client
		keyOp := kms.NewKeyOp(client.kmsClient)

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
