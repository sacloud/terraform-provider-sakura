// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMPolicy_Basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "data.sakura_iam_policy.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraIAMProjectDestroy,
			testCheckSakuraIAMServicePrincipalDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMPolicyConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target", "project"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "sakura_iam_project.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "bindings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "bindings.0.role.id", "owner"),
					resource.TestCheckResourceAttr(resourceName, "bindings.0.role.type", "preset"),
					resource.TestCheckResourceAttr(resourceName, "bindings.0.principals.0.type", "service-principal"),
					resource.TestCheckResourceAttrPair(resourceName, "bindings.0.principals.0.id", "sakura_iam_service_principal.foobar", "id"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceIAMPolicyConfig = `
resource "sakura_iam_project" "foobar" {
  name = "{{ .arg0 }}"
  code = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_iam_service_principal" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  project_id = sakura_iam_project.foobar.id
}

resource "sakura_iam_policy" "foobar" {
  target = "project"
  target_id = sakura_iam_project.foobar.id
  bindings = [{
    principals = [{
      id   = sakura_iam_service_principal.foobar.id
      type = "service-principal"
    }],
    role = {
      id   = "owner"
      type = "preset"
    }
  }]
}

data "sakura_iam_policy" "foobar" {
  target = "project"
  target_id = sakura_iam_project.foobar.id

  depends_on = [sakura_iam_policy.foobar]
}`
