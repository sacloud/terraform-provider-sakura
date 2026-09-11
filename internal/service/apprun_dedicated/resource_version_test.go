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
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/v1"
	ver "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/version"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceApprunDedicatedVersion(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedVersionDestroy,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedVersion_basic, name, globalClusterID),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("application_id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cpu"), knownvalue.Int64Exact(100)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory"), knownvalue.Int64Exact(256)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scaling_mode"), knownvalue.StringExact("manual")),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("fixed_scale"), knownvalue.Int32Exact(1)),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("image"), knownvalue.StringExact("nginx:latest")),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("created_at"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exposed_ports"), knownvalue.NotNull()),
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
								"key":    knownvalue.StringExact("ENV_VAR2"),
								"value":  knownvalue.StringExact("value2"),
								"secret": knownvalue.Bool(true),
							}),
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

					// env_vars is randomized.  Not equal to the config order by nature.
					// import puts secrets in secret_vars
					ImportStateVerifyIgnore: []string{"timeouts", "registry_password", "env_vars", "secret_vars"},
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("not found: %s", resourceName)
						}
						appID := rs.Primary.Attributes["application_id"]
						version := rs.Primary.Attributes["version"]
						return fmt.Sprintf("%s/%s", appID, version), nil
					},
				},
			},
		})
	})

	t.Run("secret_vars", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedVersionDestroy,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedVersion_secretVars, name, globalClusterID, "1", "s3cr3t"),
					Check:  testCheckSakuraApprunDedicatedVersionEnvVars(resourceName, map[string]bool{"SECRET_VAR": true, "PLAIN_VAR": false}),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.NotNull()),
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
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedVersion_secretVars, name, globalClusterID, "2", "r0tated"),
					Check:  testCheckSakuraApprunDedicatedVersionEnvVars(resourceName, map[string]bool{"SECRET_VAR": true, "PLAIN_VAR": false}),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("version"), knownvalue.NotNull()),
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

					// env_vars and secret_vars are randomized, and import cannot restore value_wo_version
					ImportStateVerifyIgnore: []string{"timeouts", "registry_password", "env_vars", "secret_vars"},
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("not found: %s", resourceName)
						}
						appID := rs.Primary.Attributes["application_id"]
						version := rs.Primary.Attributes["version"]
						return fmt.Sprintf("%s/%s", appID, version), nil
					},
				},
			},
		})
	})

	t.Run("without_health_check", func(t *testing.T) {
		resourceName := "sakura_apprun_dedicated_version.main"
		name := acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum)

		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			PreCheck:                 AccPreCheck(t),
			CheckDestroy:             testCheckSakuraApprunDedicatedVersionDestroy,
			Steps: []resource.TestStep{
				{
					Config: test.BuildConfigWithArgs(testAccSakuraResourceApprunDedicatedVersion_noHealthCheck, name, globalClusterID),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exposed_ports"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"target_port":      knownvalue.Int32Exact(80),
								"lb_port":          knownvalue.Null(),
								"use_lets_encrypt": knownvalue.Bool(false),
								"host":             knownvalue.Null(),
								"health_check":     knownvalue.Null(),
							}),
						})),
					},
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"timeouts", "registry_password"},
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("not found: %s", resourceName)
						}
						appID := rs.Primary.Attributes["application_id"]
						version := rs.Primary.Attributes["version"]
						return fmt.Sprintf("%s/%s", appID, version), nil
					},
				},
			},
		})
	})
}

func testCheckSakuraApprunDedicatedVersionDestroy(s *terraform.State) error {
	client := test.AccClientGetter().AppRunDedicatedClient

	if client == nil {
		return errors.New("AppRunDedicatedClient is nil")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apprun_dedicated_version" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		appID, err := uuid.Parse(rs.Primary.Attributes["application_id"])
		if err != nil {
			return fmt.Errorf("invalid application ID: %s", rs.Primary.Attributes["application_id"])
		}

		versionNum := int32(0)
		_, err = fmt.Sscanf(rs.Primary.Attributes["version"], "%d", &versionNum)
		if err != nil {
			return fmt.Errorf("invalid version number: %s", rs.Primary.Attributes["version"])
		}

		api := ver.NewVersionOp(client, v1.ApplicationID(appID))
		_, err = api.Read(context.Background(), v1.ApplicationVersionNumber(versionNum))

		if err == nil {
			return fmt.Errorf("version still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

// write-only values cannot be checked via the state; ask the API
func testCheckSakuraApprunDedicatedVersionEnvVars(n string, expects map[string]bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		appID, err := uuid.Parse(rs.Primary.Attributes["application_id"])
		if err != nil {
			return fmt.Errorf("invalid application ID: %s", rs.Primary.Attributes["application_id"])
		}

		versionNum := int32(0)
		_, err = fmt.Sscanf(rs.Primary.Attributes["version"], "%d", &versionNum)
		if err != nil {
			return fmt.Errorf("invalid version number: %s", rs.Primary.Attributes["version"])
		}

		api := ver.NewVersionOp(test.AccClientGetter().AppRunDedicatedClient, v1.ApplicationID(appID))
		detail, err := api.Read(context.Background(), v1.ApplicationVersionNumber(versionNum))
		if err != nil {
			return err
		}

		got := make(map[string]bool, len(detail.EnvVars))
		for _, e := range detail.EnvVars {
			got[e.Key] = e.Secret
		}

		for key, secret := range expects {
			actual, ok := got[key]
			if !ok {
				return fmt.Errorf("env var %q is missing on the API side", key)
			}
			if actual != secret {
				return fmt.Errorf("env var %q: secret = %t on the API side, want %t", key, actual, secret)
			}
		}
		return nil
	}
}

var testAccSakuraResourceApprunDedicatedVersion_basic = `
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
      key    = "ENV_VAR2"
      value  = "value2"
      secret = true
    },
    {
      key    = "ENV_VAR1"
      value  = "value1"
      secret = false
    }
  ]
}
`

//nolint:gosec
var testAccSakuraResourceApprunDedicatedVersion_secretVars = `
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

var testAccSakuraResourceApprunDedicatedVersion_noHealthCheck = `
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
      target_port = 80
      lb_port     = null
    }
  ]
}
`
