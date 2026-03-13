// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunDedicatedCertificate(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_ENABLE_APPRUN_DEDICATED_TEST")
	test.SkipIfEnvIsNotSet(t, "SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")
	test.SkipIfFakeModeEnabled(t)

	spid := os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")
	if spid == "" {
		t.Fatalf("need valid SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID environment variable")
	}

	t.Run("find by id", func(t *testing.T) {
		resourceName := "data.sakura_apprun_dedicated_certificate.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)
		cert, key, err := OreSign(fmt.Sprintf("tfacc-%s.xn--eckwd4c7cu47r2wf.jp", name))

		if err != nil {
			t.Fatalf("%q", err)
		}

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedCertificateConfigById, name, spid, string(cert), string(key)),
					Check: resource.ComposeTestCheckFunc(
						test.CheckSakuraDataSourceExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "name", "tfacc-"+name),
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttrSet(resourceName, "cluster_id"),
						resource.TestCheckResourceAttrSet(resourceName, "common_name"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					),
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

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceApprunDedicatedCertificateConfigByName, name, spid, string(cert), string(key)),
					Check: resource.ComposeTestCheckFunc(
						test.CheckSakuraDataSourceExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "name", "tfacc-"+name),
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttrSet(resourceName, "cluster_id"),
						resource.TestCheckResourceAttrSet(resourceName, "common_name"),
					),
				},
			},
		})
	})
}

var testAccCheckSakuraDataSourceApprunDedicatedCertificateSetup = `
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

resource "sakura_apprun_dedicated_certificate" "main" {
  cluster_id  = sakura_apprun_dedicated_cluster.main.id
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
  cluster_id = sakura_apprun_dedicated_cluster.main.id
}
`

var testAccCheckSakuraDataSourceApprunDedicatedCertificateConfigByName = testAccCheckSakuraDataSourceApprunDedicatedCertificateSetup + `
data "sakura_apprun_dedicated_certificate" "main" {
  name       = "tfacc-{{ .arg0 }}"
  cluster_id = sakura_apprun_dedicated_cluster.main.id

  depends_on = [sakura_apprun_dedicated_certificate.main]
}
`
