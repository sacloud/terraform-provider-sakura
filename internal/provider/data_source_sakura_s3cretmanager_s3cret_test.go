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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

func TestAccSakuraDataSourceSecretManagerSecret_basic(t *testing.T) {
	resourceName := "data.sakura_secretmanager_secret.foobar"
	rand := randomName()

	var secret v1.Secret
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfigWithArgs(testAccSakuraDataSourceSecretManagerSecret_byName, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerSecretExists("sakura_secretmanager_secret.foobar", &secret),
					testCheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

//nolint:gosec
var testAccSakuraDataSourceSecretManagerSecret_byName = `
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
}

data "sakura_secretmanager_secret" "foobar" {
  name     = "{{ .arg0 }}"
  vault_id = sakura_secretmanager.foobar.id

  depends_on = [sakura_secretmanager_secret.foobar]
}`
