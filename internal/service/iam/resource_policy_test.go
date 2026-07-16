// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraIAMPolicy_basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "sakura_iam_policy.foobar"
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
				Config: test.BuildConfigWithArgs(testAccSakuraIAMPolicy_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target", "project"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "sakura_iam_project.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "bindings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bindings.0.role.id", "owner"),
					resource.TestCheckResourceAttr(resourceName, "bindings.0.role.type", "preset"),
					resource.TestCheckResourceAttr(resourceName, "bindings.0.principals.0.type", "service-principal"),
					resource.TestCheckResourceAttrPair(resourceName, "bindings.0.principals.0.id", "sakura_iam_service_principal.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "bindings.1.role.id", "organization-admin"),
					resource.TestCheckResourceAttr(resourceName, "bindings.1.role.type", "preset"),
					resource.TestCheckResourceAttr(resourceName, "bindings.1.principals.0.type", "service-principal"),
					resource.TestCheckResourceAttrPair(resourceName, "bindings.1.principals.0.id", "sakura_iam_service_principal.foobar", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMPolicy_update, rand),
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

func TestAccImportSakuraIAMPolicy_basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "sakura_iam_policy.foobar"
	rand := test.RandomName()

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"target":                       "project",
			"bindings.#":                   "2",
			"bindings.0.role.id":           "owner",
			"bindings.0.role.type":         "preset",
			"bindings.0.principals.0.type": "service-principal",
			"bindings.1.role.id":           "organization-admin",
			"bindings.1.role.type":         "preset",
			"bindings.1.principals.0.type": "service-principal",
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0], "target_id")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraIAMProjectDestroy,
			testCheckSakuraIAMServicePrincipalDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMPolicy_basic, rand),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateCheck:                     checkFn,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "target_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["target"], rs.Primary.Attributes["target_id"]), nil
				},
			},
		},
	})
}

const testAccSakuraIAMPolicy_basic = `
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
  },
  {
    principals = [{
      id   = sakura_iam_service_principal.foobar.id
      type = "service-principal"
    }],
    role = {
      id   = "organization-admin"
      type = "preset"
    }
  }]
}
`

const testAccSakuraIAMPolicy_update = `
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
`

const testAccSakuraIAMPolicy_import = `
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
  },
  {
    principals = [{
      id   = sakura_iam_service_principal.foobar.id
      type = "service-principal"
    }],
    role = {
      id   = "organization-admin"
      type = "preset"
    }
  }]
}
`
