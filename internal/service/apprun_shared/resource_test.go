// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_shared_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraApprunShared_basic(t *testing.T) {
	resourceName := "sakura_apprun_shared.foobar"
	rand := test.RandomName()

	var application v1.Application
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "1Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:latest"),
					resource.TestMatchResourceAttr(resourceName, "status", regexp.MustCompile(".+")),
					resource.TestMatchResourceAttr(resourceName, "public_url", regexp.MustCompile(".+")),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "2"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "2Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:tag1"),
				),
			},
		},
	})
}

func TestAccSakuraApprunShared_withCRUser(t *testing.T) {
	resourceName := "sakura_apprun_shared.foobar"
	rand := test.RandomName()

	var application v1.Application
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withCRUser, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "1Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:latest"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.server", "apprun-test.sakuracr.jp"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.username", "user"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.password", "password"),
				),
			},
		},
	})
}

func TestAccSakuraApprunShared_withEnv(t *testing.T) {
	resourceName := "sakura_apprun_shared.foobar"
	rand := test.RandomName()

	var application v1.Application
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withEnv, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "1Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:latest"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.0.key", "key"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.0.value", "value"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.1.key", "key2"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.1.value", "value2"),
				),
			},
		},
	})
}

func TestAccSakuraApprunShared_withEnvUpdate(t *testing.T) {
	resourceName := "sakura_apprun_shared.foobar"
	rand := test.RandomName()

	var application v1.Application
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withEnv, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.0.key", "key"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.0.value", "value"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.1.key", "key2"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.1.value", "value2"),
					resource.TestCheckNoResourceAttr(resourceName, "components.0.env.2"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withEnvUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.#", "2"),
					// Update
					resource.TestCheckResourceAttr(resourceName, "components.0.env.0.key", "key"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.0.value", "value-updated"),
					// Remove&Add
					resource.TestCheckResourceAttr(resourceName, "components.0.env.1.key", "key3"),
					resource.TestCheckResourceAttr(resourceName, "components.0.env.1.value", "value3"),
				),
			},
		},
	})
}

func TestAccSakuraApprunShared_withProbe(t *testing.T) {
	resourceName := "sakura_apprun_shared.foobar"
	rand := test.RandomName()

	var application v1.Application
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withProbe, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "1Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:latest"),
					resource.TestCheckResourceAttr(resourceName, "components.0.probe.http_get.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "components.0.probe.http_get.port", "80"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withProbeUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "1Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:latest"),
					resource.TestCheckResourceAttr(resourceName, "components.0.probe.http_get.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "components.0.probe.http_get.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "components.0.probe.http_get.headers.0.name", "name1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.probe.http_get.headers.0.value", "value1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.probe.http_get.headers.1.name", "name2"),
					resource.TestCheckResourceAttr(resourceName, "components.0.probe.http_get.headers.1.value", "value2"),
				),
			},
		},
	})
}

func TestAccSakuraApprunShared_withTraffic(t *testing.T) {
	resourceName := "sakura_apprun_shared.foobar"
	rand := test.RandomName()

	var application v1.Application
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withTraffic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "1Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:latest"),
					resource.TestCheckResourceAttr(resourceName, "traffics.0.version_index", "0"),
					resource.TestCheckResourceAttr(resourceName, "traffics.0.percent", "100"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withTrafficUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraApprunSharedExists(resourceName, &application),
					testCheckSakuraApprunSharedAttributes(&application),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "1Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:latest"),
					resource.TestCheckResourceAttr(resourceName, "traffics.0.version_index", "0"),
					resource.TestCheckResourceAttr(resourceName, "traffics.0.percent", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffics.1.version_index", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffics.1.percent", "99"),
				),
			},
		},
	})
}

func TestAccSakuraApprunShared_withPacketFilter(t *testing.T) {
	resourceName := "sakura_apprun_shared.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withPacketFilter, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.0.from_ip", "192.0.2.0"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.0.from_ip_prefix_length", "28"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withPacketFilterUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.0.from_ip", "192.0.2.0"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.0.from_ip_prefix_length", "28"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.1.from_ip", "192.0.2.128"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.1.from_ip_prefix_length", "28"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withPacketFilterDisabled, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.0.from_ip", "192.0.2.0"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.0.from_ip_prefix_length", "28"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.1.from_ip", "192.0.2.128"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.1.from_ip_prefix_length", "28"),
				),
			},
		},
	})
}

func TestAccImportSakuraApprunShared_basic(t *testing.T) {
	rand := test.RandomName()
	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                    rand,
			"timeout_seconds":         "90",
			"port":                    "80",
			"min_scale":               "0",
			"max_scale":               "1",
			"components.0.name":       "compo1",
			"components.0.max_cpu":    "0.5",
			"components.0.max_memory": "1Gi",
			"components.0.deploy_source.container_registry.image": "apprun-test.sakuracr.jp/test1:latest",
		}

		return test.CompareStateMulti(s[0], expects)
	}

	resourceName := "sakura_apprun_shared.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_basic, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"status",
					"public_url",
				},
			},
		},
	})
}

func TestAccImportSakuraApprunShared_withCRUser(t *testing.T) {
	rand := test.RandomName()
	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                    rand,
			"timeout_seconds":         "90",
			"port":                    "80",
			"min_scale":               "0",
			"max_scale":               "1",
			"components.0.name":       "compo1",
			"components.0.max_cpu":    "0.5",
			"components.0.max_memory": "1Gi",
			"components.0.deploy_source.container_registry.image":    "apprun-test.sakuracr.jp/test1:latest",
			"components.0.deploy_source.container_registry.username": "user",
			"components.0.deploy_source.container_registry.password": "",
		}

		return test.CompareStateMulti(s[0], expects)
	}

	resourceName := "sakura_apprun_shared.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withCRUser, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"components.0.deploy_source.container_registry.password",
					"status",
					"public_url",
				},
			},
		},
	})
}

func TestAccImportSakuraApprunShared_withEnv(t *testing.T) {
	rand := test.RandomName()
	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                    rand,
			"timeout_seconds":         "90",
			"port":                    "80",
			"min_scale":               "0",
			"max_scale":               "1",
			"components.0.name":       "compo1",
			"components.0.max_cpu":    "0.5",
			"components.0.max_memory": "1Gi",
			"components.0.deploy_source.container_registry.image": "apprun-test.sakuracr.jp/test1:latest",
			"components.0.env.#":       "2",
			"components.0.env.0.key":   "key",
			"components.0.env.0.value": "value",
			"components.0.env.1.key":   "key2",
			"components.0.env.1.value": "value2",
		}

		return test.CompareStateMulti(s[0], expects)
	}

	resourceName := "sakura_apprun_shared.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraApprunSharedDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraApprunShared_withEnv, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"status",
					"public_url",
				},
			},
		},
	})
}

func testCheckSakuraApprunSharedExists(n string, application *v1.Application) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no AppRun Application ID is set")
		}

		client := test.AccClientGetter()
		appOp := apprun.NewApplicationOp(client.AppRunClient)

		found, err := appOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("not found AppRun Application: %s", rs.Primary.ID)
		}

		*application = *found
		return nil
	}
}

func testCheckSakuraApprunSharedAttributes(application *v1.Application) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(application.Components) == 0 {
			return errors.New("unexpected application components: components is nil")
		}

		c := (application.Components)[0]
		if c.DeploySource.ContainerRegistry == nil {
			return errors.New("unexpected application components: container_registry is nil")
		}

		return nil
	}
}

func testCheckSakuraApprunSharedDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	appOp := apprun.NewApplicationOp(client.AppRunClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apprun_shared" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := appOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists AppRun Application:%s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraApprunShared_basic = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
  }]
}
`

const testAccSakuraApprunShared_update = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 8080
  min_scale       = 0
  max_scale       = 2
  components = [{
    name       = "compo1"
    max_cpu    = "1"
    max_memory = "2Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:tag1"
      }
    }
  }]
}
`

const testAccSakuraApprunShared_withCRUser = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image    = "apprun-test.sakuracr.jp/test1:latest"
        server   = "apprun-test.sakuracr.jp"
        username = "user"
        password = "password"
      }
    }
  }]
}
`

const testAccSakuraApprunShared_withEnv = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
    env = [{
      key   = "key"
      value = "value"
    },
	{
      key   = "key2"
      value = "value2"
    },
    {
      key   = "key2"
      value = "value2"
    }]
  }]
}
`

const testAccSakuraApprunShared_withEnvUpdate = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
	// Updated
    env = [{
      key   = "key"
      value = "value-updated"
    },
	// Removed
    // env {
    //   key   = "key2"
    //   value = "value2"
    // }
	// Added
    {
      key   = "key3"
      value = "value3"
    }]
  }]
}
`

const testAccSakuraApprunShared_withProbe = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
    probe = {
      http_get = {
        path = "/"
        port = 80
      }
    }
  }]
}
`

const testAccSakuraApprunShared_withProbeUpdate = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
    probe = {
      http_get = {
        path = "/"
        port = 80
        headers = [{
          name  = "name1"
          value = "value1"
        },
        {
          name  = "name2"
          value = "value2"
        }]
      }
    }
  }]
}
`

const testAccSakuraApprunShared_withTraffic = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
  }]
  traffics = [{
    version_index = 0
    percent       = 100
  }]
}
`

const testAccSakuraApprunShared_withTrafficUpdate = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 10
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
  }]
  traffics = [{
    version_index = 0
    percent       = 1
  },
  {
    version_index = 1
    percent       = 99
  }]
}
`

const testAccSakuraApprunShared_withPacketFilter = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
  }]
  packet_filter = {
	enabled = true
	settings = [{
	  from_ip               = "192.0.2.0"
      from_ip_prefix_length = "28"
	}]
  }
}
`

const testAccSakuraApprunShared_withPacketFilterUpdate = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
  }]
  packet_filter = {
	enabled = true
	settings = [{
	  from_ip               = "192.0.2.0"
      from_ip_prefix_length = "28"
	},
	{
	  from_ip               = "192.0.2.128"
      from_ip_prefix_length = "28"
	}]
  }
}
`

const testAccSakuraApprunShared_withPacketFilterDisabled = `
resource "sakura_apprun_shared" "foobar" {
  name            = "{{ .arg0 }}"
  timeout_seconds = 90
  port            = 80
  min_scale       = 0
  max_scale       = 1
  components = [{
    name       = "compo1"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image = "apprun-test.sakuracr.jp/test1:latest"
      }
    }
  }]
  packet_filter = {
	enabled = false
	settings = [{
	  from_ip               = "192.0.2.0"
      from_ip_prefix_length = "28"
	},
	{
	  from_ip               = "192.0.2.128"
      from_ip_prefix_length = "28"
	}]
  }
}
`
