// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedCluster(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_ENABLE_APPRUN_DEDICATED_TEST")
	test.SkipIfFakeModeEnabled(t)

	t.Run("find by id", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_cluster.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.Test(t, resource.TestCase{
			//PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedClusterConfigById, name),
					Check: resource.ComposeTestCheckFunc(
						test.CheckSakuraDataSourceExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tfacc-%s", name)),
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttrSet(resourceName, "service_principal_id"),
						resource.TestCheckResourceAttrSet(resourceName, "has_lets_encrypt_email"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttr(resourceName, "ports.#", "2"),
					),
				},
			},
		})
	})

	t.Run("find by name", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_cluster.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.Test(t, resource.TestCase{
			//PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedClusterConfigByName, name),
					Check: resource.ComposeTestCheckFunc(
						test.CheckSakuraDataSourceExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tfacc-%s", name)),
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttrSet(resourceName, "service_principal_id"),
						resource.TestCheckResourceAttrSet(resourceName, "has_lets_encrypt_email"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttr(resourceName, "ports.#", "2"),
					),
				},
			},
		})
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedClusterConfigById = `
resource "sakura_iam_project" "main" {
  name = "tfacc-{{ .arg0 }}"
  code = "code{{ .arg0 }}"
  description = "Terraform Apprun Dedicated Test Project"
}

resource "sakura_iam_service_principal" "main" {
  project_id  = sakura_iam_project.main.id
  name        = "tfacc-{{ .arg0 }}"
  description = "Service Principal for Apprun Dedicated API integration tests"
}

resource "sakura_iam_policy" "main" {
  target    = "project"
  target_id = sakura_iam_project.main.id

  bindings = [
    {
      role = {
        type = "preset"
        id   = "resource-creator"
      }
      principals = [
        {
          type = "service-principal"
          id   = sakura_iam_service_principal.main.id
        }
      ]
    }
  ]
}

resource "sakura_apprun_dedicated_cluster" "main" {
  name                 = "tfacc-{{ .arg0 }}"
  service_principal_id = sakura_iam_service_principal.main.id

  ports = [
    {
      port     = 443
      protocol = "https"
    },
    {
      port     = 80
      protocol = "http"
    }
  ]
}

data "sakura_apprun_dedicated_cluster" "main" {
  id = sakura_apprun_dedicated_cluster.main.id
}
`

var testAccCheckSakuraDataSourceApprunDedicatedClusterConfigByName = `
resource "sakura_iam_project" "main" {
  name = "tfacc-{{ .arg0 }}"
  code = "code{{ .arg0 }}"
  description = "Terraform Apprun Dedicated Test Project"
}

resource "sakura_iam_service_principal" "main" {
  project_id  = sakura_iam_project.main.id
  name        = "tfacc-{{ .arg0 }}"
  description = "Service Principal for Apprun Dedicated API integration tests"
}

resource "sakura_iam_policy" "main" {
  target    = "project"
  target_id = sakura_iam_project.main.id

  bindings = [
    {
      role = {
        type = "preset"
        id   = "resource-creator"
      }
      principals = [
        {
          type = "service-principal"
          id   = sakura_iam_service_principal.main.id
        }
      ]
    }
  ]
}

resource "sakura_apprun_dedicated_cluster" "main" {
  name                 = "tfacc-{{ .arg0 }}"
  service_principal_id = sakura_iam_service_principal.main.id

  ports = [
    {
      port     = 443
      protocol = "https"
    },
    {
      port     = 80
      protocol = "http"
    }
  ]
}

data "sakura_apprun_dedicated_cluster" "main" {
  name       = "tfacc-{{ .arg0 }}"
  depends_on = [sakura_apprun_dedicated_cluster.main]
}
`
