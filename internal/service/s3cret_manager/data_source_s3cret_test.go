// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

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
