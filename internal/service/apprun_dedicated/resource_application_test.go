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
	app "github.com/sacloud/apprun-dedicated-api-go/apis/application"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceApprunDedicatedApplication_basic(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_application.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedApplicationDestroy,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedApplication_basic, name, globalClusterID),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_id"), knownvalue.StringExact(globalClusterID)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_name"), knownvalue.NotNull()),
					},
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"timeouts"},
				},
			},
		})
	})
	t.Run("update", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_application.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			Steps: []resource.TestStep{
				// create app & version
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedApplication_version, name, globalClusterID),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("tfacc-%s", name))),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cluster_id"), knownvalue.StringExact(globalClusterID)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("active_version"), knownvalue.Null()),
					},
				},
				// set version
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedApplication_update, name, globalClusterID),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("active_version"), knownvalue.Int32Exact(1)),
					},
				},
				// deactivate version (necessary for proper teardown)
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedApplication_teardown, name, globalClusterID),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("active_version"), knownvalue.Null()),
					},
				},
			},
		})
	})
}

func testCheckSakuraApprunDedicatedApplicationDestroy(s *terraform.State) error {
	client := test.AccClientGetter().AppRunDedicatedClient

	if client == nil {
		return errors.New("AppRunDedicatedClient is nil")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apprun_dedicated_application" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		api := app.NewApplicationOp(client)
		appID, err := uuid.Parse(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("invalid application ID: %s", rs.Primary.ID)
		}

		_, err = api.Read(context.Background(), v1.ApplicationID(appID))

		if err == nil {
			return fmt.Errorf("application still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

var testAccSakuraResourceApprunDedicatedApplication_basic = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id     = "{{ .arg1 }}"
  name           = "tfacc-{{ .arg0 }}"
  active_version = null
}
`

var testAccSakuraResourceApprunDedicatedApplication_version = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id     = "{{ .arg1 }}"
  name           = "tfacc-{{ .arg0 }}"
  active_version = null
}

resource "sakura_apprun_dedicated_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = 100
  memory         = 256
  scaling_mode   = "manual"
  fixed_scale    = 1
  image          = "nginx:latest"

  exposed_ports = [
    {
      target_port        = 80
	  lb_port            = null
      health_check       = {
        path             = "/"
        interval_seconds = 10
        timeout_seconds  = 5
      }
    }
  ]
}
`

var testAccSakuraResourceApprunDedicatedApplication_update = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id     = "{{ .arg1 }}"
  name           = "tfacc-{{ .arg0 }}"
  active_version = 1
}

resource "sakura_apprun_dedicated_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = 100
  memory         = 256
  scaling_mode   = "manual"
  fixed_scale    = 1
  image          = "nginx:latest"

  exposed_ports = [
    {
      target_port        = 80
	  lb_port            = null
      health_check       = {
        path             = "/"
        interval_seconds = 10
        timeout_seconds  = 5
      }
    }
  ]
}
`

var testAccSakuraResourceApprunDedicatedApplication_teardown = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id     = "{{ .arg1 }}"
  name           = "tfacc-{{ .arg0 }}"
  active_version = null
}

resource "sakura_apprun_dedicated_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = 100
  memory         = 256
  scaling_mode   = "manual"
  fixed_scale    = 1
  image          = "nginx:latest"

  exposed_ports = [
    {
      target_port        = 80
	  lb_port            = null
      health_check       = {
        path             = "/"
        interval_seconds = 10
        timeout_seconds  = 5
      }
    }
  ]
}
`
