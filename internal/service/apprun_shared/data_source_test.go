// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_shared_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceApprunShared_basic(t *testing.T) {
	resourceName := "data.sakura_apprun_shared.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceApprunShared_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "timeout_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "min_scale", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.name", "compo1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_cpu", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "components.0.max_memory", "1Gi"),
					resource.TestCheckResourceAttr(resourceName, "components.0.deploy_source.container_registry.image", "apprun-test.sakuracr.jp/test1:latest"),
				),
			},
		},
	})
}

func TestAccSakuraDataSourceApprunShared_withCRUser(t *testing.T) {
	resourceName := "data.sakura_apprun_shared.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceApprunShared_withCRUser, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
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
				),
			},
		},
	})
}

func TestAccSakuraDataSourceApprunShared_withProbe(t *testing.T) {
	resourceName := "data.sakura_apprun_shared.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceApprunShared_withProbe, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
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

func TestAccSakuraDataSourceApprunShared_withTraffic(t *testing.T) {
	// Use resource to check traffics setting
	resourceName := "sakura_apprun_shared.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceApprunShared_withTraffic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
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
		},
	})
}

func TestAccSakuraDataSourceApprunShared_withPacketFilter(t *testing.T) {
	resourceName := "data.sakura_apprun_shared.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceApprunShared_withPacketFilter, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.0.from_ip", "192.0.2.0"),
					resource.TestCheckResourceAttr(resourceName, "packet_filter.settings.0.from_ip_prefix_length", "24"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceApprunShared_basic = `
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

data "sakura_apprun_shared" "foobar" {
  name = sakura_apprun_shared.foobar.name

  depends_on = [
    sakura_apprun_shared.foobar
  ]
}
`

var testAccSakuraDataSourceApprunShared_withCRUser = `
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

data "sakura_apprun_shared" "foobar" {
  name = sakura_apprun_shared.foobar.name
}
`

var testAccSakuraDataSourceApprunShared_withProbe = `
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

data "sakura_apprun_shared" "foobar" {
  name = sakura_apprun_shared.foobar.name
}
`

var testAccSakuraDataSourceApprunShared_withTraffic = `
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

data "sakura_apprun_shared" "foobar" {
  name = sakura_apprun_shared.foobar.name
}
`

var testAccSakuraDataSourceApprunShared_withPacketFilter = `
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
      from_ip_prefix_length = "24"
	}]
  }
}

data "sakura_apprun_shared" "foobar" {
  name = sakura_apprun_shared.foobar.name
}
`
