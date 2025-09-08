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

package secret_manager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceSecretManagerSecret_basic(t *testing.T) {
	resourceName := "data.sakura_secret_manager_secret.foobar"
	rand := test.RandomName()

	var secret v1.Secret
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSecretManagerSecret_byName, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSecretManagerSecretExists("sakura_secret_manager_secret.foobar", &secret),
					test.CheckSakuraDataSourceExists(resourceName),
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
}

data "sakura_secret_manager_secret" "foobar" {
  name     = "{{ .arg0 }}"
  vault_id = sakura_secret_manager.foobar.id

  depends_on = [sakura_secret_manager_secret.foobar]
}`
