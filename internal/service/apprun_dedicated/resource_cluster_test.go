// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/sacloud/apprun-dedicated-api-go/apis/cluster"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceApprunDedicatedCluster_basic(t *testing.T) {
	resourceName := "sakura_apprun_dedicated_cluster.main"
	name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 AccPreCheck(t),
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunDedicatedClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedCluster_basic, name, os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ports"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_principal_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("created_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_lets_encrypt_email"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Importing doesn't retrieve the Let's Encrypt email (sensitive/computed), and timeouts are not part of state
				ImportStateVerifyIgnore: []string{"lets_encrypt_email", "timeouts"},
			},
		},
	})
}

func TestAccSakuraResourceApprunDedicatedCluster_update(t *testing.T) {
	resourceName := "sakura_apprun_dedicated_cluster.main"
	name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 AccPreCheck(t),
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunDedicatedClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedCluster_basic, name, os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ports"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("lets_encrypt_email"), knownvalue.Null()),
				},
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedCluster_update, name, os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ports"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ports").AtSliceIndex(0).AtMapKey("port"), knownvalue.Int64Exact(80)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ports").AtSliceIndex(0).AtMapKey("protocol"), knownvalue.StringExact("http")),
					// Note: lets_encrypt_email cannot be verified directly as it's write-only for updates
				},
			},
		},
	})
}

func testCheckSakuraApprunDedicatedClusterDestroy(s *terraform.State) error {
	client := test.AccClientGetter().AppRunDedicatedClient

	if client == nil {
		return errors.New("AppRunDedicatedClient is nil")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apprun_dedicated_cluster" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		api := cluster.NewClusterOp(client)
		clusterID, err := uuid.Parse(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("invalid cluster ID: %s", rs.Primary.ID)
		}

		_, err = api.Read(context.Background(), v1.ClusterID(clusterID))

		if err == nil {
			return fmt.Errorf("cluster still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

var testAccSakuraResourceApprunDedicatedCluster_basic = `
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
`

var testAccSakuraResourceApprunDedicatedCluster_update = `
resource "sakura_apprun_dedicated_cluster" "main" {
  name                 = "tfacc-{{ .arg0 }}"
  service_principal_id = "{{ .arg1 }}"
  lets_encrypt_email   = "admin@example.com"

  # port 80 mandatory for let's encrypt
  ports = [
    {
      port     = 80
      protocol = "http"
    }
  ]
}
`
