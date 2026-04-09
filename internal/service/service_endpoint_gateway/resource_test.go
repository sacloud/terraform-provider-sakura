// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package service_endpoint_gateway_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	service_endpoint_gateway "github.com/sacloud/service-endpoint-gateway-api-go"
	v1 "github.com/sacloud/service-endpoint-gateway-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

const (
	envSEG_OBJECT_STORAGE_ENDPOINT1    = "SAKURA_SEG_ENDPOINT_OBJECT_STORAGE_ENDPOINT_1"
	envSEG_OBJECT_STORAGE_ENDPOINT2    = "SAKURA_SEG_ENDPOINT_OBJECT_STORAGE_ENDPOINT_2"
	envSEG_MONITORING_SUITE_ENDPOINT   = "SAKURA_SEG_MONITORING_SUITE_ENDPOINT"
	envSEG_CONTAINER_REGISTRY_ENDPOINT = "SAKURA_SEG_CONTAINER_REGISTRY_ENDPOINT"
	envSEG_AI_ENGINE_ENDPOINT          = "SAKURA_SEG_AI_ENGINE_ENDPOINT"
	envSEG_DNS_PRIVATE_HOSTED_ZONE     = "SAKURA_SEG_DNS_PRIVATE_HOSTED_ZONE"
	envSEG_DNS_UPSTREAM_SERVER_1       = "SAKURA_SEG_DNS_UPSTREAM_SERVER_1"
	envSEG_DNS_UPSTREAM_SERVER_2       = "SAKURA_SEG_DNS_UPSTREAM_SERVER_2"
)

func TestAccSakuraSEG_basic(t *testing.T) {
	resourceName := "sakura_service_endpoint_gateway.foobar"

	test.SkipIfEnvIsNotSet(t,
		envSEG_OBJECT_STORAGE_ENDPOINT1, envSEG_OBJECT_STORAGE_ENDPOINT2, envSEG_MONITORING_SUITE_ENDPOINT, envSEG_CONTAINER_REGISTRY_ENDPOINT, envSEG_AI_ENGINE_ENDPOINT,
		envSEG_DNS_PRIVATE_HOSTED_ZONE, envSEG_DNS_UPSTREAM_SERVER_1, envSEG_DNS_UPSTREAM_SERVER_2,
	)
	rand := test.RandomName()
	object_storageEndpoint1 := os.Getenv(envSEG_OBJECT_STORAGE_ENDPOINT1)
	object_storageEndpoint2 := os.Getenv(envSEG_OBJECT_STORAGE_ENDPOINT2)
	monitoring_suiteEndpoint := os.Getenv(envSEG_MONITORING_SUITE_ENDPOINT)
	container_registryEndpoint := os.Getenv(envSEG_CONTAINER_REGISTRY_ENDPOINT)
	ai_engineEndpoint := os.Getenv(envSEG_AI_ENGINE_ENDPOINT)
	dns_private_hostedzone := os.Getenv(envSEG_DNS_PRIVATE_HOSTED_ZONE)
	dns_upstream_server_1 := os.Getenv(envSEG_DNS_UPSTREAM_SERVER_1)
	dns_upstream_server_2 := os.Getenv(envSEG_DNS_UPSTREAM_SERVER_2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraSEGDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSEG_basic, rand, object_storageEndpoint1, object_storageEndpoint2, monitoring_suiteEndpoint, container_registryEndpoint, ai_engineEndpoint, dns_private_hostedzone, dns_upstream_server_1, dns_upstream_server_2),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSEGExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "zone", "tk1b"),
					resource.TestCheckResourceAttrPair(resourceName, "vswitch_id", "sakura_vswitch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "server_ip_addresses.0", "192.168.128.31"),
					resource.TestCheckResourceAttr(resourceName, "netmask", "28"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.object_storage_endpoints.0", object_storageEndpoint1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.object_storage_endpoints.1", object_storageEndpoint2),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.monitoring_suite_endpoints.0", monitoring_suiteEndpoint),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.container_registry_endpoints.0", container_registryEndpoint),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.ai_engine_endpoints.0", ai_engineEndpoint),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.app_run_dedicated_control_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_suite_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "dns_forwarding.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "dns_forwarding.private_hosted_zone", dns_private_hostedzone),
					resource.TestCheckResourceAttr(resourceName, "dns_forwarding.upstream_dns_1", dns_upstream_server_1),
					resource.TestCheckResourceAttr(resourceName, "dns_forwarding.upstream_dns_2", dns_upstream_server_2),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSEG_update, rand, object_storageEndpoint1),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSEGExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "zone", "tk1b"),
					resource.TestCheckResourceAttrPair(resourceName, "vswitch_id", "sakura_vswitch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "server_ip_addresses.0", "192.168.128.129"),
					resource.TestCheckResourceAttr(resourceName, "netmask", "28"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.object_storage_endpoints.0", object_storageEndpoint1),
					resource.TestCheckResourceAttr(resourceName, "monitoring_suite_enable", "false"),
				),
			},
		},
	})
}

func testCheckSakuraSEGExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Service Endpoint Gateway ID is set")
		}

		client, err := testGetClientFromState(s)
		if err != nil {
			return err
		}

		segOp := service_endpoint_gateway.NewServiceEndpointGatewayOp(client)
		found, err := segOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Appliance.ID != rs.Primary.ID {
			return fmt.Errorf("not found Service Endpoint Gateway: %s", rs.Primary.ID)
		}
		return nil
	}
}

func testCheckSakuraSEGDestroy(s *terraform.State) error {
	client, err := testGetClientFromState(s)
	if err != nil {
		return err
	}
	segOp := service_endpoint_gateway.NewServiceEndpointGatewayOp(client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_service_endpoint_gateway" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}
		_, err = segOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists Service Endpoint Gateway: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testGetClientFromState(s *terraform.State) (*v1.Client, error) {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_service_endpoint_gateway" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		client := test.AccClientGetter()
		zone := rs.Primary.Attributes["zone"]
		apiRoot := fmt.Sprintf("https://secure.sakura.ad.jp/cloud/zone/%s/api/cloud/1.1", zone)
		return service_endpoint_gateway.NewClientWithAPIRootURL(client.SaClient, apiRoot)
	}
	return nil, errors.New("Service Endpoint Gateway resource not found in state")
}

const testAccSakuraSEG_basic = `
resource "sakura_vswitch" "foobar" {
	name = "{{ .arg0 }}"
	zone = "tk1b"
}

resource "sakura_service_endpoint_gateway" "foobar" {
	zone        = "tk1b"
	vswitch_id  = sakura_vswitch.foobar.id
	server_ip_addresses = ["192.168.128.31"]
	netmask     = 28
	endpoint_setting = {
		object_storage_endpoints = ["{{ .arg1 }}", "{{ .arg2 }}"]
		monitoring_suite_endpoints = ["{{ .arg3 }}"]
		container_registry_endpoints = ["{{ .arg4 }}"]
		ai_engine_endpoints = ["{{ .arg5 }}"]
		app_run_dedicated_control_enabled = false
	}
	monitoring_suite_enable = true
	dns_forwarding = {
		enabled = true
		private_hosted_zone = "{{ .arg6 }}"
		upstream_dns_1 = "{{ .arg7 }}"
		upstream_dns_2 = "{{ .arg8 }}"
	}
}
`

const testAccSakuraSEG_update = `
resource "sakura_vswitch" "foobar" {
	name = "{{ .arg0 }}"
	zone = "tk1b"
}

resource "sakura_service_endpoint_gateway" "foobar" {
	zone        = "tk1b"
	vswitch_id  = sakura_vswitch.foobar.id
	server_ip_addresses = ["192.168.128.129"]
	netmask     = 28
	endpoint_setting = {
		object_storage_endpoints = ["{{ .arg1 }}"]
	}
	monitoring_suite_enable = false
}
`
