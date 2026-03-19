// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedVersion(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_ENABLE_APPRUN_DEDICATED_TEST")
	test.SkipIfEnvIsNotSet(t, "SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")
	test.SkipIfFakeModeEnabled(t)

	spid := os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")
	if spid == "" {
		t.Fatalf("need valid SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID environment variable")
	}

	t.Run("find by id", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedVersionConfig, name, spid),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName, "version"),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
						resource.TestCheckResourceAttr(resourceName, "cpu", "1000"),
						resource.TestCheckResourceAttr(resourceName, "memory", "512"),
						resource.TestCheckResourceAttr(resourceName, "scaling_mode", "manual"),
						resource.TestCheckResourceAttr(resourceName, "image", "nginx:latest"),
					),
				},
			},
		})
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedVersionConfig = `
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

resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = sakura_apprun_dedicated_cluster.main.id
  name       = "tfacc-{{ .arg0 }}"
}

resource "sakura_apprun_dedicated_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = 1000
  memory         = 512
  image          = "nginx:latest"
  cmd            = ["/bin/sh"]
  scaling_mode   = "manual"
  fixed_scale    = 1
}

data "sakura_apprun_dedicated_version" "main" {
  application_id = sakura_apprun_dedicated_version.main.application_id
  version             = sakura_apprun_dedicated_version.main.version
}
`
