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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunDedicatedClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tfacc-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "ports.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "service_principal_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "has_lets_encrypt_email"),
				),
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
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunDedicatedClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tfacc-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "ports.#", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "lets_encrypt_email"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedCluster_update, name, os.Getenv("SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID")),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunDedicatedClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tfacc-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ports.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ports.0.protocol", "http"),
					// Note: lets_encrypt_email cannot be verified directly as it's write-only for updates
				),
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

func testCheckSakuraApprunDedicatedClusterExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no cluster ID is set")
		}

		client := test.AccClientGetter().AppRunDedicatedClient

		if client == nil {
			return errors.New("AppRunDedicatedClient is nil")
		}

		clusterID, err := uuid.Parse(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("invalid cluster ID: %s", rs.Primary.ID)
		}

		api := cluster.NewClusterOp(client)
		found, err := api.Read(context.Background(), v1.ClusterID(clusterID))

		if err != nil {
			return fmt.Errorf("failed to read cluster: %s", err)
		}

		if uuid.UUID(found.ClusterID).String() != rs.Primary.ID {
			return fmt.Errorf("cluster not found: %s", rs.Primary.ID)
		}

		return nil
	}
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
