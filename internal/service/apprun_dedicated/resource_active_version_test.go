// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	app "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/application"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/v1"
	ver "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/version"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceApprunDedicatedActiveVersion(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_active_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedActiveVersionDestroy,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_basic, name, globalClusterID, "100", "nginx:latest"),
					Check:  testCheckSakuraApprunDedicatedActiveVersionActivated(resourceName),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.CompareValuePairs(resourceName, tfjsonpath.New("id"), "sakura_apprun_dedicated_application.main", tfjsonpath.New("id"), compare.ValuesSame()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("application_id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cpu"), knownvalue.Int64Exact(100)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory"), knownvalue.Int64Exact(256)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scaling_mode"), knownvalue.StringExact("manual")),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("fixed_scale"), knownvalue.Int32Exact(1)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("image"), knownvalue.StringExact("nginx:latest")),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("created_at"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exposed_ports"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"target_port":      knownvalue.Int32Exact(80),
								"lb_port":          knownvalue.Null(),
								"use_lets_encrypt": knownvalue.Bool(false),
								"host":             knownvalue.Null(),
								"health_check": knownvalue.ObjectExact(map[string]knownvalue.Check{
									"path":             knownvalue.StringExact("/"),
									"interval_seconds": knownvalue.Int32Exact(10),
									"timeout_seconds":  knownvalue.Int32Exact(5),
								}),
							}),
						})),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("env_vars"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"key":    knownvalue.StringExact("ENV_VAR1"),
								"value":  knownvalue.StringExact("value1"),
								"secret": knownvalue.Bool(false),
							}),
						})),
					},
				},
				{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,

					// env_vars is randomized, import puts secrets in secret_vars, and
					// active_node_count changes while the containers of the new version start
					ImportStateVerifyIgnore: []string{"timeouts", "registry_password", "env_vars", "secret_vars", "active_node_count"},
				},
			},
		})
	})

	t.Run("update", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_active_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedActiveVersionDestroy,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_basic, name, globalClusterID, "100", "nginx:latest"),
					Check:  testCheckSakuraApprunDedicatedActiveVersionActivated(resourceName),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.Int32Exact(1)),
					},
				},
				{
					// a change is an in-place update that creates and activates a new version
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_basic, name, globalClusterID, "200", "nginx:alpine"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
					},
					Check: resource.ComposeTestCheckFunc(
						testCheckSakuraApprunDedicatedActiveVersionActivated(resourceName),
						testCheckSakuraApprunDedicatedActiveVersionRetains(resourceName, 1),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.Int32Exact(2)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cpu"), knownvalue.Int64Exact(200)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("image"), knownvalue.StringExact("nginx:alpine")),
					},
				},
				{
					// a change of timeouts alone is an update that keeps the active version
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_timeouts, name, globalClusterID, "200", "nginx:alpine"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
					},
					Check: testCheckSakuraApprunDedicatedActiveVersionActivated(resourceName),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.Int32Exact(2)),
					},
				},
			},
		})
	})

	t.Run("external_activation", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_active_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)
		var appID string

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedActiveVersionDestroy,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_basic, name, globalClusterID, "100", "nginx:latest"),
					Check: resource.ComposeTestCheckFunc(
						testCheckSakuraApprunDedicatedActiveVersionActivated(resourceName),
						testCheckSakuraApprunDedicatedResourceID(resourceName, &appID),
					),
				},
				{
					// external activation: the refresh follows it (env_vars dropped), the plan reconciles
					PreConfig: func() {
						version := testAccSakuraApprunDedicatedCreateVersion(t, appID, testAccSakuraApprunDedicatedActiveVersionParams(200, "nginx:alpine"))
						testAccSakuraApprunDedicatedActivateVersion(t, appID, version)
					},
					RefreshState: true,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "version", "2"),
						resource.TestCheckResourceAttr(resourceName, "cpu", "200"),
						resource.TestCheckResourceAttr(resourceName, "image", "nginx:alpine"),
						resource.TestCheckResourceAttr(resourceName, "env_vars.#", "0"),
					),
					RefreshPlanChecks: resource.RefreshPlanChecks{
						PostRefresh: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
							plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("cpu"), knownvalue.Int64Exact(100)),
							plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("image"), knownvalue.StringExact("nginx:latest")),
							plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("env_vars"), knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"key":    knownvalue.StringExact("ENV_VAR1"),
									"value":  knownvalue.StringExact("value1"),
									"secret": knownvalue.Bool(false),
								}),
							})),
						},
					},
					ExpectNonEmptyPlan: true,
				},
				{
					// external deactivation: the resource is gone, the next apply creates a new version
					PreConfig: func() { testAccSakuraApprunDedicatedDeactivateApplication(t, appID) },
					Config:    test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_basic, name, globalClusterID, "100", "nginx:latest"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate)},
					},
					Check: testCheckSakuraApprunDedicatedActiveVersionActivated(resourceName),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cpu"), knownvalue.Int64Exact(100)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("image"), knownvalue.StringExact("nginx:latest")),
					},
				},
			},
		})
	})

	t.Run("secret_vars", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_active_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedActiveVersionDestroy,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_secretVars, name, globalClusterID, "1", "s3cr3t"),
					Check:  testCheckSakuraApprunDedicatedVersionEnvVars(resourceName, map[string]bool{"SECRET_VAR": true, "PLAIN_VAR": false}),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.Int32Exact(1)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("env_vars"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"key":    knownvalue.StringExact("PLAIN_VAR"),
								"value":  knownvalue.StringExact("plain"),
								"secret": knownvalue.Bool(false),
							}),
						})),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("secret_vars"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"key":              knownvalue.StringExact("SECRET_VAR"),
								"value_wo":         knownvalue.Null(),
								"value_wo_version": knownvalue.Int32Exact(1),
							}),
						})),
					},
				},
				{
					// bump value_wo_version to roll out a new value_wo
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_secretVars, name, globalClusterID, "2", "r0tated"),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
					},
					Check: resource.ComposeTestCheckFunc(
						testCheckSakuraApprunDedicatedActiveVersionActivated(resourceName),
						testCheckSakuraApprunDedicatedVersionEnvVars(resourceName, map[string]bool{"SECRET_VAR": true, "PLAIN_VAR": false}),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.Int32Exact(2)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("secret_vars"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"key":              knownvalue.StringExact("SECRET_VAR"),
								"value_wo":         knownvalue.Null(),
								"value_wo_version": knownvalue.Int32Exact(2),
							}),
						})),
					},
				},
				{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,

					// env_vars and secret_vars are randomized, import cannot restore value_wo_version,
					// and active_node_count changes while the containers of the new version start
					ImportStateVerifyIgnore: []string{"timeouts", "registry_password", "env_vars", "secret_vars", "active_node_count"},
				},
			},
		})
	})

	t.Run("empty_lists", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_active_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedActiveVersionDestroy,
			Steps: []resource.TestStep{
				{
					// empty lists must stay empty, not become null (a diff on every plan)
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedActiveVersion_emptyLists, name, globalClusterID),
					Check:  testCheckSakuraApprunDedicatedActiveVersionActivated(resourceName),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cmd"), knownvalue.ListExact([]knownvalue.Check{})),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("env_vars"), knownvalue.ListExact([]knownvalue.Check{})),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("secret_vars"), knownvalue.ListExact([]knownvalue.Check{})),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exposed_ports").AtSliceIndex(0).AtMapKey("host"), knownvalue.SetExact([]knownvalue.Check{})),
					},
				},
			},
		})
	})
}

// destroy deactivates; the versions stay
func testCheckSakuraApprunDedicatedActiveVersionDestroy(s *terraform.State) error {
	client := test.AccClientGetter().AppRunDedicatedClient

	if client == nil {
		return errors.New("AppRunDedicatedClient is nil")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apprun_dedicated_active_version" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		appID, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid application ID: %s", rs.Primary.ID)
		}

		detail, err := app.NewApplicationOp(client).Read(context.Background(), v1.ApplicationID(appID))

		if saclient.IsNotFoundError(err) {
			continue // the application is gone with it
		}

		if err != nil {
			return err
		}

		if detail.ActiveVersion != nil {
			return fmt.Errorf("application %s still has an active version: %d", rs.Primary.ID, *detail.ActiveVersion)
		}
	}
	return nil
}

// the API side must agree on which version is active
func testCheckSakuraApprunDedicatedActiveVersionActivated(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		appID, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid application ID: %s", rs.Primary.ID)
		}

		detail, err := app.NewApplicationOp(test.AccClientGetter().AppRunDedicatedClient).Read(context.Background(), v1.ApplicationID(appID))
		if err != nil {
			return err
		}

		if detail.ActiveVersion == nil {
			return fmt.Errorf("application %s has no active version on the API side", rs.Primary.ID)
		}

		if got := fmt.Sprint(*detail.ActiveVersion); got != rs.Primary.Attributes["version"] {
			return fmt.Errorf("active version = %s on the API side, want %s", got, rs.Primary.Attributes["version"])
		}
		return nil
	}
}

func testCheckSakuraApprunDedicatedActiveVersionRetains(n string, version int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		appID, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid application ID: %s", rs.Primary.ID)
		}

		api := ver.NewVersionOp(test.AccClientGetter().AppRunDedicatedClient, v1.ApplicationID(appID))
		if _, err := api.Read(context.Background(), v1.ApplicationVersionNumber(version)); err != nil {
			return fmt.Errorf("version %d of application %s: %w", version, appID, err)
		}
		return nil
	}
}

func testAccSakuraApprunDedicatedCreateVersion(t *testing.T, appID string, params ver.CreateParams) int32 {
	t.Helper()

	api := ver.NewVersionOp(test.AccClientGetter().AppRunDedicatedClient, v1.ApplicationID(uuid.MustParse(appID)))
	created, err := api.Create(t.Context(), params)

	if err != nil {
		t.Fatalf("failed to create a version of application %s: %s", appID, err)
	}

	return int32(created.Version)
}

// the _basic config as SDK params, minus env_vars so that the refresh has something to drop
func testAccSakuraApprunDedicatedActiveVersionParams(cpu int64, image string) ver.CreateParams {
	return ver.CreateParams{
		CPU:                    cpu,
		Memory:                 256,
		ScalingMode:            v1.ScalingModeManual,
		FixedScale:             new(int32(1)),
		Image:                  image,
		RegistryPasswordAction: v1.RegistryPasswordActionRemove,
		ExposedPorts: []ver.ExposedPort{
			{
				TargetPort:  80,
				HealthCheck: &v1.HealthCheck{Path: "/", IntervalSeconds: 10, TimeoutSeconds: 5},
			},
		},
	}
}

var testAccSakuraResourceApprunDedicatedActiveVersion_basic = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = "{{ .arg1 }}"
  name       = "tfacc-{{ .arg0 }}"
}

resource "sakura_apprun_dedicated_active_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = {{ .arg2 }}
  memory         = 256
  scaling_mode   = "manual"
  fixed_scale    = 1
  image          = "{{ .arg3 }}"

  exposed_ports = [
    {
      target_port  = 80
      lb_port      = null
      health_check = {
        path             = "/"
        interval_seconds = 10
        timeout_seconds  = 5
      }
    }
  ]

  env_vars = [
    {
      key   = "ENV_VAR1"
      value = "value1"
    }
  ]
}
`

var testAccSakuraResourceApprunDedicatedActiveVersion_timeouts = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = "{{ .arg1 }}"
  name       = "tfacc-{{ .arg0 }}"
}

resource "sakura_apprun_dedicated_active_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = {{ .arg2 }}
  memory         = 256
  scaling_mode   = "manual"
  fixed_scale    = 1
  image          = "{{ .arg3 }}"

  exposed_ports = [
    {
      target_port  = 80
      lb_port      = null
      health_check = {
        path             = "/"
        interval_seconds = 10
        timeout_seconds  = 5
      }
    }
  ]

  env_vars = [
    {
      key   = "ENV_VAR1"
      value = "value1"
    }
  ]

  timeouts = {
    update = "10m"
  }
}
`

//nolint:gosec
var testAccSakuraResourceApprunDedicatedActiveVersion_secretVars = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = "{{ .arg1 }}"
  name       = "tfacc-{{ .arg0 }}"
}

resource "sakura_apprun_dedicated_active_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = 100
  memory         = 256
  scaling_mode   = "manual"
  fixed_scale    = 1
  image          = "nginx:latest"

  exposed_ports = [
    {
      target_port = 80
      lb_port     = null
    }
  ]

  env_vars = [
    {
      key   = "PLAIN_VAR"
      value = "plain"
    }
  ]

  secret_vars = [
    {
      key              = "SECRET_VAR"
      value_wo         = "{{ .arg3 }}"
      value_wo_version = {{ .arg2 }}
    }
  ]
}
`

var testAccSakuraResourceApprunDedicatedActiveVersion_emptyLists = `
resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = "{{ .arg1 }}"
  name       = "tfacc-{{ .arg0 }}"
}

resource "sakura_apprun_dedicated_active_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = 100
  memory         = 256
  scaling_mode   = "manual"
  fixed_scale    = 1
  image          = "nginx:latest"
  cmd            = []

  exposed_ports = [
    {
      target_port = 80
      host        = []
    }
  ]

  env_vars    = []
  secret_vars = []
}
`
