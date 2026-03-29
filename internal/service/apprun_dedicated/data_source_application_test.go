// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedApplication(t *testing.T) {
	t.Run("find by id", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_application.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedApplicationConfigById, name, globalClusterID),
					Check: resource.ComposeTestCheckFunc(
						test.CheckSakuraDataSourceExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "name", "tfacc-"+name),
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttrSet(resourceName, "cluster_id"),
						resource.TestCheckResourceAttrSet(resourceName, "cluster_name"),
					),
				},
			},
		})
	})

	t.Run("find by name", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_application.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedApplicationConfigByName, name, globalClusterID),
					Check: resource.ComposeTestCheckFunc(
						test.CheckSakuraDataSourceExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "name", "tfacc-"+name),
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttrSet(resourceName, "cluster_id"),
					),
				},
			},
		})
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedApplicationSetup = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = "{{ .arg1 }}"
  name       = "tfacc-{{ .arg0 }}"
}
`

var testAccCheckSakuraDataSourceApprunDedicatedApplicationConfigById = testAccCheckSakuraDataSourceApprunDedicatedApplicationSetup + `
data "sakura_apprun_dedicated_application" "main" {
  id         = sakura_apprun_dedicated_application.main.id
  cluster_id = "{{ .arg1 }}"
}
`

var testAccCheckSakuraDataSourceApprunDedicatedApplicationConfigByName = testAccCheckSakuraDataSourceApprunDedicatedApplicationSetup + `
data "sakura_apprun_dedicated_application" "main" {
  name       = "tfacc-{{ .arg0 }}"
  cluster_id = "{{ .arg1 }}"

  depends_on = [sakura_apprun_dedicated_application.main]
}
`
