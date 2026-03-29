// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedCluster(t *testing.T) {
	t.Run("find by id", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_cluster.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedClusterConfigById,
						name,
						os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_principal_id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_lets_encrypt_email"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("created_at"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ports"), knownvalue.ListSizeExact(2)),
					},
				},
			},
		})
	})

	t.Run("find by name", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_cluster.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(
						testAccCheckSakuraDataSourceApprunDedicatedClusterConfigByName,
						name,
						os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_principal_id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_lets_encrypt_email"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("created_at"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ports"), knownvalue.ListSizeExact(2)),
					},
				},
			},
		})
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedClusterConfigById = `
resource "sakura_apprun_dedicated_cluster" "main" {
  name                 = "tfacc-{{ .arg0 }}"
  service_principal_id = "{{ .arg1 }}"

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
resource "sakura_apprun_dedicated_cluster" "main" {
  name                 = "tfacc-{{ .arg0 }}"
  service_principal_id = "{{ .arg1 }}"

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
