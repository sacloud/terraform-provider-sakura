// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package seg_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/saclient-go"
	service_endpoint_gateway "github.com/sacloud/service-endpoint-gateway-api-go"
	v1 "github.com/sacloud/service-endpoint-gateway-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

const (
	envSEGObjectStorageEndpoint1    = "SAKURA_SEG_OBJECT_STORAGE_ENDPOINT_1"
	envSEGObjectStorageEndpoint2    = "SAKURA_SEG_OBJECT_STORAGE_ENDPOINT_2"
	envSEGMonitoringSuiteEndpoint   = "SAKURA_SEG_MONITORING_SUITE_ENDPOINT"
	envSEGContainerRegistryEndpoint = "SAKURA_SEG_CONTAINER_REGISTRY_ENDPOINT"
	envSEGAIEngineEndpoint          = "SAKURA_SEG_AI_ENGINE_ENDPOINT"
	envSEGDNSPrivateHostedZone      = "SAKURA_SEG_DNS_PRIVATE_HOSTED_ZONE"
	envSEGDNSUpstreamServer1        = "SAKURA_SEG_DNS_UPSTREAM_SERVER_1"
	envSEGDNSUpstreamServer2        = "SAKURA_SEG_DNS_UPSTREAM_SERVER_2"
)

func TestAccSakuraSEG_basic(t *testing.T) {
	resourceName := "sakura_seg.foobar"

	test.SkipIfEnvIsNotSet(t,
		envSEGObjectStorageEndpoint1, envSEGObjectStorageEndpoint2, envSEGMonitoringSuiteEndpoint, envSEGContainerRegistryEndpoint, envSEGAIEngineEndpoint,
		envSEGDNSPrivateHostedZone, envSEGDNSUpstreamServer1, envSEGDNSUpstreamServer2,
	)
	rand := test.RandomName()
	objectStorageEndpoint1 := os.Getenv(envSEGObjectStorageEndpoint1)
	objectStorageEndpoint2 := os.Getenv(envSEGObjectStorageEndpoint2)
	monitoringSuiteEndpoint := os.Getenv(envSEGMonitoringSuiteEndpoint)
	containerRegistryEndpoint := os.Getenv(envSEGContainerRegistryEndpoint)
	aiEngineEndpoint := os.Getenv(envSEGAIEngineEndpoint)
	dnsPrivateHostedZone := os.Getenv(envSEGDNSPrivateHostedZone)
	dnsUpstreamServer1 := os.Getenv(envSEGDNSUpstreamServer1)
	dnsUpstreamServer2 := os.Getenv(envSEGDNSUpstreamServer2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraSEGDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSEGBasic, rand, objectStorageEndpoint1, objectStorageEndpoint2, monitoringSuiteEndpoint, containerRegistryEndpoint, aiEngineEndpoint, dnsPrivateHostedZone, dnsUpstreamServer1, dnsUpstreamServer2),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSEGExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "zone", "tk1b"),
					resource.TestCheckResourceAttrPair(resourceName, "vswitch_id", "sakura_vswitch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "server_ip_addresses.0", "192.168.128.31"),
					resource.TestCheckResourceAttr(resourceName, "netmask", "28"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.object_storage_endpoints.0", objectStorageEndpoint1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.object_storage_endpoints.1", objectStorageEndpoint2),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.monitoring_suite_endpoints.0", monitoringSuiteEndpoint),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.container_registry_endpoints.0", containerRegistryEndpoint),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.ai_engine_endpoints.0", aiEngineEndpoint),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.apprun_dedicated_control_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_suite_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "dns_forwarding.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "dns_forwarding.private_hosted_zone", dnsPrivateHostedZone),
					resource.TestCheckResourceAttr(resourceName, "dns_forwarding.upstream_dns_1", dnsUpstreamServer1),
					resource.TestCheckResourceAttr(resourceName, "dns_forwarding.upstream_dns_2", dnsUpstreamServer2),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSEGUpdate, rand, objectStorageEndpoint1),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSEGExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "zone", "tk1b"),
					resource.TestCheckResourceAttrPair(resourceName, "vswitch_id", "sakura_vswitch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "server_ip_addresses.0", "192.168.128.129"),
					resource.TestCheckResourceAttr(resourceName, "netmask", "28"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_setting.object_storage_endpoints.0", objectStorageEndpoint1),
					resource.TestCheckResourceAttr(resourceName, "monitoring_suite_enabled", "false"),
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
		if rs.Type != "sakura_seg" {
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
		if rs.Type != "sakura_seg" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		client := test.AccClientGetter()
		zone := rs.Primary.Attributes["zone"]
		clientAPI, err := client.SaClient.DupWith(saclient.WithZone(zone))
		if err != nil {
			return nil, err
		}

		return service_endpoint_gateway.NewClient(clientAPI)
	}
	return nil, errors.New("Service Endpoint Gateway resource not found in state")
}

const testAccSakuraSEGBasic = `
resource "sakura_vswitch" "foobar" {
	name = "{{ .arg0 }}"
	zone = "tk1b"
}

resource "sakura_seg" "foobar" {
	zone        = "tk1b"
	vswitch_id  = sakura_vswitch.foobar.id
	server_ip_addresses = ["192.168.128.31"]
	netmask     = 28
	endpoint_setting = {
		object_storage_endpoints = ["{{ .arg1 }}", "{{ .arg2 }}"]
		monitoring_suite_endpoints = ["{{ .arg3 }}"]
		container_registry_endpoints = ["{{ .arg4 }}"]
		ai_engine_endpoints = ["{{ .arg5 }}"]
		apprun_dedicated_control_enabled = false
	}
	monitoring_suite_enabled = true
	dns_forwarding = {
		enabled = true
		private_hosted_zone = "{{ .arg6 }}"
		upstream_dns_1 = "{{ .arg7 }}"
		upstream_dns_2 = "{{ .arg8 }}"
	}
}
`

const testAccSakuraSEGUpdate = `
resource "sakura_vswitch" "foobar" {
	name = "{{ .arg0 }}"
	zone = "tk1b"
}

resource "sakura_seg" "foobar" {
	zone        = "tk1b"
	vswitch_id  = sakura_vswitch.foobar.id
	server_ip_addresses = ["192.168.128.129"]
	netmask     = 28
	endpoint_setting = {
		object_storage_endpoints = ["{{ .arg1 }}"]
	}
	monitoring_suite_enabled = false
}
`
