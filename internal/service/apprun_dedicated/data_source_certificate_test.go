// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedCertificate(t *testing.T) {
	t.Run("find by id", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_certificate.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)
		cert, key, err := OreSign(fmt.Sprintf("tfacc-%s.xn--eckwd4c7cu47r2wf.jp", name))

		if err != nil {
			t.Fatalf("%q", err)
		}

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedCertificateConfigById, name, globalClusterID, string(cert), string(key)),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact("tfacc-"+name)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("common_name"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("created_at"), knownvalue.NotNull()),
					},
				},
			},
		})
	})

	t.Run("find by name", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_certificate.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)
		cert, key, err := OreSign(fmt.Sprintf("tfacc-%s.xn--eckwd4c7cu47r2wf.jp", name))

		if err != nil {
			t.Fatalf("%q", err)
		}

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedCertificateConfigByName, name, globalClusterID, string(cert), string(key)),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact("tfacc-"+name)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("common_name"), knownvalue.NotNull()),
					},
				},
			},
		})
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedCertificateSetup = `
resource "sakura_apprun_dedicated_certificate" "main" {
  cluster_id  = "{{ .arg1 }}"
  name        = "tfacc-{{ .arg0 }}"

  certificate_pem = <<EOF
{{ .arg2 }}
EOF

  private_key_pem = <<EOF
{{ .arg3 }}
EOF
}
`

var testAccCheckSakuraDataSourceApprunDedicatedCertificateConfigById = testAccCheckSakuraDataSourceApprunDedicatedCertificateSetup + `
data "sakura_apprun_dedicated_certificate" "main" {
  id         = sakura_apprun_dedicated_certificate.main.id
  cluster_id = "{{ .arg1 }}"
}
`

var testAccCheckSakuraDataSourceApprunDedicatedCertificateConfigByName = testAccCheckSakuraDataSourceApprunDedicatedCertificateSetup + `
data "sakura_apprun_dedicated_certificate" "main" {
  name       = "tfacc-{{ .arg0 }}"
  cluster_id = "{{ .arg1 }}"

  depends_on = [sakura_apprun_dedicated_certificate.main]
}
`
