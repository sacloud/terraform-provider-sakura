// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	cert "github.com/sacloud/apprun-dedicated-api-go/apis/certificate"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceApprunDedicatedCertificate(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_certificate.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)
		domain := fmt.Sprintf("tfacc-%s.xn--eckwd4c7cu47r2wf.jp", name)
		cert, key, err := OreSign(domain)

		if err != nil {
			t.Fatalf("%q", err)
		}

		config := test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedCertificate_basic, name, globalClusterID, string(cert), string(key))
		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedCertificateDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_id"), knownvalue.StringExact(globalClusterID)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("common_name"), knownvalue.StringExact(domain)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("created_at"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("not_before"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("not_after"), knownvalue.NotNull()),
					},
				},
				{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("not found: %s", resourceName)
						}
						clusterID := rs.Primary.Attributes["cluster_id"]
						id := rs.Primary.Attributes["id"]
						return fmt.Sprintf("%s/%s", clusterID, id), nil
					},
					// Importing doesn't retrieve write-only/sensitive fields
					ImportStateVerifyIgnore: []string{
						"certificate_pem",
						"private_key_pem",
						"intermediate_certificate_pem",
						"timeouts",
					},
				},
			},
		})
	})

	t.Run("update", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_certificate.main"
		name := acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)
		domain := fmt.Sprintf("tfacc-%s.xn--eckwd4c7cu47r2wf.jp", name)
		cert, key, err := OreSign(domain)

		if err != nil {
			t.Fatalf("%q", err)
		}

		configBasic := test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedCertificate_basic, name, globalClusterID, string(cert), string(key))
		configUpdate := test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedCertificate_update, name, globalClusterID, string(cert), string(key))
		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedCertificateDestroy,
			Steps: []resource.TestStep{
				{
					Config: configBasic,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("intermediate_certificate_pem"), knownvalue.Null()),
					},
				},
				{
					Config: configUpdate,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s-2", name))),
					},
				},
			},
		})
	})
}

func testCheckSakuraApprunDedicatedCertificateDestroy(s *terraform.State) error {
	client := test.AccClientGetter().AppRunDedicatedClient
	if client == nil {
		return errors.New("AppRunDedicatedClient is nil")
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apprun_dedicated_certificate" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}
		clusterID, err := uuid.Parse(rs.Primary.Attributes["cluster_id"])
		if err != nil {
			return fmt.Errorf("invalid cluster ID: %s", rs.Primary.Attributes["cluster_id"])
		}
		certID, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid certificate ID: %s", rs.Primary.ID)
		}
		api := cert.NewCertificateOp(client, v1.ClusterID(clusterID))
		_, err = api.Read(context.Background(), v1.CertificateID(certID))
		if err == nil {
			return fmt.Errorf("certificate still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

var testAccSakuraResourceApprunDedicatedCertificate_basic = `
resource "sakura_apprun_dedicated_certificate" "main" {
	cluster_id         = "{{ .arg1 }}"
	name              = "tfacc-{{ .arg0 }}"
	certificate_pem   = <<EOF
{{ .arg2 }}
EOF
	private_key_pem   = <<EOF
{{ .arg3 }}
EOF
}
`

var testAccSakuraResourceApprunDedicatedCertificate_update = `
resource "sakura_apprun_dedicated_certificate" "main" {
	cluster_id                    = "{{ .arg1 }}"
	name                         = "tfacc-{{ .arg0 }}-2"
	certificate_pem              = <<EOF
{{ .arg2 }}
EOF
	private_key_pem              = <<EOF
{{ .arg3 }}
EOF
}
`
