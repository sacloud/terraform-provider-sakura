// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedVersion(t *testing.T) {
	t.Run("find by id", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedVersionConfig, name, globalClusterID),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("application_id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cpu"), knownvalue.Int64Exact(1000)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory"), knownvalue.Int64Exact(512)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scaling_mode"), knownvalue.StringExact("manual")),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("image"), knownvalue.StringExact("nginx:latest")),
					},
				},
			},
		})
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedVersionConfig = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = "{{ .arg1 }}"
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
