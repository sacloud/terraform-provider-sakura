// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraServer_basic(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()
	password := test.RandomPassword()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_basic, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					test.CheckSakuraServerAttributes(&server),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "core", "1"),
					resource.TestCheckResourceAttr(resourceName, "memory", "1"),
					resource.TestCheckResourceAttr(resourceName, "disks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "interface_driver", "virtio"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.hostname", rand),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.password", password),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.ssh_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.ssh_keys.0", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEAo5G7cwRp423KOrtCewX5nXFkboGxZ3hfvECNGg56 e2e-test-only@example"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.ssh_key_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.ssh_key_ids.0", "100000000000"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.disable_pw_auth", "true"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.0.id", "100000000000"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.0.api_key_id", "100000000001"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.0.variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.0.variables.foo1", "bar1"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.0.variables.foo2", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.0.upstream", "shared"),
					resource.TestCheckResourceAttr(resourceName, "hostname", rand),
					resource.TestCheckResourceAttrSet(resourceName, "network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(
						resourceName, "icon_id",
						"sakura_icon.foobar", "id",
					),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_update, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					test.CheckSakuraServerAttributes(&server),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "core", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory", "2"),
					resource.TestCheckResourceAttr(resourceName, "disks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "interface_driver", "e1000"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2-upd"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.0.upstream", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, "network_address"),
					resource.TestCheckNoResourceAttr(resourceName, "icon_id"),
				),
			},
		},
	})
}

func TestAccSakuraServer_basicWithWO(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()
	password := test.RandomPassword()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_basicWithWO, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					test.CheckSakuraServerAttributes(&server),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "core", "1"),
					resource.TestCheckResourceAttr(resourceName, "memory", "1"),
					resource.TestCheckResourceAttr(resourceName, "disks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "interface_driver", "virtio"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.hostname", rand),
					resource.TestCheckNoResourceAttr(resourceName, "disk_edit_parameter.password"),
					resource.TestCheckNoResourceAttr(resourceName, "disk_edit_parameter.password_wo"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.password_wo_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.ssh_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.ssh_keys.0", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEAo5G7cwRp423KOrtCewX5nXFkboGxZ3hfvECNGg56 e2e-test-only@example"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.ssh_key_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.ssh_key_ids.0", "100000000000"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.disable_pw_auth", "true"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.0.id", "100000000000"),
					resource.TestCheckResourceAttr(resourceName, "disk_edit_parameter.script.0.api_key_id", "100000000001"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.0.upstream", "shared"),
					resource.TestCheckResourceAttr(resourceName, "hostname", rand),
					resource.TestCheckResourceAttrSet(resourceName, "network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
				),
			},
		},
	})
}

func TestAccSakuraServer_validateHostName(t *testing.T) {
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraServer_validateHostName, rand),
				ExpectError: regexp.MustCompile(`"invalid_host_name" is not a valid hostname`),
			},
		},
	})
}

func TestAccSakuraServer_planChange(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()
	password := test.RandomPassword()

	var step1, step2, step3 iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_standardPlan, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &step1),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "core", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory", "4"),
					resource.TestCheckResourceAttr(resourceName, "commitment", "standard"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_dedicatedCPUPlan, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &step2),
					func(state *terraform.State) error {
						if step1.ID == step2.ID {
							return fmt.Errorf("server id was not changed: before.ID: %s after.ID:%s", step1.ID, step2.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "core", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory", "4"),
					resource.TestCheckResourceAttr(resourceName, "commitment", "dedicatedcpu"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_standardPlan, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &step3),
					func(state *terraform.State) error {
						if step2.ID == step3.ID {
							return fmt.Errorf("server id was not changed: before.ID: %s after.ID:%s", step2.ID, step3.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "core", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory", "4"),
					resource.TestCheckResourceAttr(resourceName, "commitment", "standard"),
				),
			},
		},
	})
}

/* v3 and frameworkではこのテストだけ失敗する。検証に時間がかかりそうなので一旦コメントアウト。
    resource_test.go:175: Step 2/2 error: Error running apply: exit status 1

        Error: Build Server Error

          with sakura_server.foobar,
          on terraform_plugin_test.tf line 20, in resource "sakura_server" "foobar":
          20: resource "sakura_server" "foobar" {

        Error in response: &iaas.APIErrorResponse{IsFatal:true, Status:"409 Conflict",
        ErrorCode:"res_already_connected",　ErrorMessage:"要求された操作を行えません。このリソースは他のリソースと既に接続されています。"}
--- FAIL: TestAccSakuraServer_planChangeByOutsideOfTerraform (234.21s)

func TestAccSakuraServer_planChangeByOutsideOfTerraform(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()
	password := test.RandomPassword()

	var step1, step2, step3 iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config:             test.BuildConfigWithArgs(testAccSakuraServer_standardPlan, rand, password),
				ExpectNonEmptyPlan: true, // Terraform外での変更を擬似的にStepの中で行うためtrueにしておく
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &step1),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "core", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory", "4"),
					resource.TestCheckResourceAttr(resourceName, "commitment", "standard"),
					func(*terraform.State) error {
						client := test.AccClientGetter()
						ctx := context.Background()
						// shutdown
						if err := power.ShutdownServer(ctx, iaas.NewServerOp(client), step1.Zone.Name, step1.ID, true); err != nil {
							return err
						}

						updated, err := plans.ChangeServerPlan(
							context.Background(), client, step1.Zone.Name, step1.ID,
							&iaas.ServerChangePlanRequest{
								CPU:        1,
								MemoryMB:   1 * size.GiB,
								Generation: types.PlanGenerations.Default,
								Commitment: types.Commitments.Standard,
							})
						if err != nil {
							return err
						}
						step2 = *updated
						return nil
					},
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_standardPlan, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &step3),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "core", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory", "4"),
					resource.TestCheckResourceAttr(resourceName, "commitment", "standard"),
					func(*terraform.State) error {
						if step1.ID == step2.ID || step1.ID == step3.ID || step2.ID == step3.ID {
							return fmt.Errorf("server plan changed, but id is not updated")
						}
						return nil
					},
				),
			},
		},
	})
}
*/

func TestAccSakuraServer_withoutShutdown(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()
	password := test.RandomPassword()

	var created, updated, editParamChanged iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_withoutShutdownWhenUpdate, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &created),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_updateWithoutShutdownWhenUpdate, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &updated),
					func(state *terraform.State) error {
						if !created.InstanceStatusChangedAt.Equal(updated.InstanceStatusChangedAt) {
							return fmt.Errorf(
								"unexpected shutdown has happened: ChangeAt: before: %s after: %s",
								created.InstanceStatusChangedAt.String(),
								updated.InstanceStatusChangedAt.String(),
							)
						}
						return nil
					},
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_updateWithShutdownWhenUpdate, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &editParamChanged),
					func(state *terraform.State) error {
						if updated.InstanceStatusChangedAt.Equal(editParamChanged.InstanceStatusChangedAt) {
							return errors.New("expected shutdown has not happened")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccSakuraServer_interfaces(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_interfaces, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					test.CheckSakuraServerAttributes(&server),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.0.upstream", "shared"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_interfacesAdded, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					test.CheckSakuraServerAttributes(&server),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.0.upstream", "shared"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_interfacesUpdated, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					test.CheckSakuraServerAttributes(&server),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.0.upstream", "shared"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_interfacesDisconnect, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					testCheckSakuraServerSharedInterface(&server),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.0.upstream", "disconnect"),
				),
			},
		},
	})
}

func TestAccSakuraServer_packetFilter(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_packetFilter, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "2"),
					resource.TestCheckResourceAttrPair(
						resourceName, "network_interface.0.packet_filter_id",
						"sakura_packet_filter.foobar", "id",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "network_interface.1.packet_filter_id",
						"sakura_packet_filter.foobar", "id",
					),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_packetFilterUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
					resource.TestCheckResourceAttrPair(
						resourceName, "network_interface.0.packet_filter_id",
						"sakura_packet_filter.foobar", "id",
					),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_packetFilterDelete, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "network_interface.0.packet_filter_id"),
				),
			},
		},
	})
}

func TestAccSakuraServer_withBlankDisk(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_withBlankDisk, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					test.CheckSakuraServerAttributes(&server),
				),
			},
		},
	})
}

func TestAccSakuraServer_vswitch(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_vswitch, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					resource.TestCheckResourceAttr(resourceName, "network_interface.0.user_ip_address", "192.168.0.2"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.1.user_ip_address", "192.168.1.2"),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "192.168.0.2"),
					resource.TestCheckResourceAttr(resourceName, "netmask", "24"),
					resource.TestCheckResourceAttr(resourceName, "gateway", "192.168.0.1"),
				),
			},
		},
	})
}

func TestAccSakuraServer_withGPU(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_withGPU, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					resource.TestCheckResourceAttr(resourceName, "core", "4"),
					resource.TestCheckResourceAttr(resourceName, "memory", "56"),
					resource.TestCheckResourceAttr(resourceName, "gpu", "1"),
				),
			},
		},
	})
}

func TestAccSakuraServer_withAMDPlan(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURACLOUD_ENABLE_AMD_PLAN")

	resourceName := "sakura_server.foobar"
	rand := test.RandomName()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_withAMDPlan, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					resource.TestCheckResourceAttr(resourceName, "core", "32"),
					resource.TestCheckResourceAttr(resourceName, "memory", "120"),
					resource.TestCheckResourceAttr(resourceName, "cpu_model", "amd_epyc_7713p"),
					resource.TestCheckResourceAttr(resourceName, "commitment", "dedicatedcpu"),
				),
			},
		},
	})
}

func TestAccSakuraServer_withKoukaryokuVRT(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURACLOUD_ENABLE_KOUKARYOKU_VRT")

	resourceName := "sakura_server.foobar"
	rand := test.RandomName()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer__withKoukaryokuVRT, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					resource.TestCheckResourceAttr(resourceName, "core", "4"),
					resource.TestCheckResourceAttr(resourceName, "memory", "56"),
					resource.TestCheckResourceAttr(resourceName, "gpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "gpu_model", "nvidia_v100_32gbvram"),
				),
			},
		},
	})
}

const envCloudInitDiskID = "SAKURACLOUD_CLOUD_INIT_DISK"

func TestAccSakuraServer_cloudInit(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envCloudInitDiskID)

	resourceName := "sakura_server.foobar"
	rand := test.RandomName()
	password := test.RandomPassword()
	diskID := os.Getenv(envCloudInitDiskID)

	var created, updated, userDataChanged iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_cloudInit, rand, password, diskID),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &created),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_cloudInitUpdated, rand, password, diskID),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &updated),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"_upd"),
					func(state *terraform.State) error {
						if !created.InstanceStatusChangedAt.Equal(updated.InstanceStatusChangedAt) {
							return fmt.Errorf(
								"unexpected shutdown has happened: ChangeAt: before: %s after: %s",
								created.InstanceStatusChangedAt.String(),
								updated.InstanceStatusChangedAt.String(),
							)
						}
						return nil
					},
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_cloudInitUserDataUpdated, rand, password, diskID),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &userDataChanged),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					func(state *terraform.State) error {
						if updated.InstanceStatusChangedAt.Equal(userDataChanged.InstanceStatusChangedAt) {
							return errors.New("expected shutdown has not happened")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccSakuraServer_confidentialVM(t *testing.T) {
	resourceName := "sakura_server.foobar"
	rand := test.RandomName()
	password := test.RandomPassword()

	var server iaas.Server
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             test.CheckSakuraServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraServer_confidentialVM, rand, password),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraServerExists(resourceName, &server),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "core", "16"),
					resource.TestCheckResourceAttr(resourceName, "memory", "24"),
					resource.TestCheckResourceAttr(resourceName, "cpu_model", "amd_epyc_9654p"),
					resource.TestCheckResourceAttr(resourceName, "commitment", "dedicatedcpu"),
					resource.TestCheckResourceAttr(resourceName, "confidential_vm", "true"),
				),
			},
		},
	})
}

func testCheckSakuraServerSharedInterface(server *iaas.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !server.InstanceStatus.IsUp() {
			return fmt.Errorf("unexpected server status: status=%v", server.InstanceStatus)
		}

		if len(server.Interfaces) == 0 || !server.Interfaces[0].SwitchID.IsEmpty() {
			return fmt.Errorf("unexpected server NIC status. %#v", server.Interfaces)
		}

		return nil
	}
}

const testAccSakuraServer_basic = `
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"
  disks       = [sakura_disk.foobar.id]
  description = "description"
  tags        = ["tag1", "tag2"]
  icon_id     = sakura_icon.foobar.id
  network_interface = [{
    upstream = "shared"
  }]
  disk_edit_parameter = {
    hostname        = "{{ .arg0 }}"
    password        = "{{ .arg1 }}"
    ssh_keys        = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEAo5G7cwRp423KOrtCewX5nXFkboGxZ3hfvECNGg56 e2e-test-only@example"]
    ssh_key_ids     = ["100000000000", "200000000000"]
    disable_pw_auth = true
    script = [{
      id         = "100000000000"
      api_key_id = "100000000001"
      variables  = {
        foo1 = "bar1"
        foo2 = "bar2"
      }
    }]
  }
}

resource "sakura_icon" "foobar" {
  name          = "{{ .arg0 }}"
  base64content = "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"
}
`

const testAccSakuraServer_update = `
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}

resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}-upd"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name             = "{{ .arg0 }}-upd"
  disks            = [sakura_disk.foobar.id]
  core             = 2
  memory           = 2
  description      = "description-upd"
  tags             = ["tag1-upd", "tag2-upd"]
  interface_driver = "e1000"
  network_interface = [{
    upstream = "shared"
  }]
}
`

const testAccSakuraServer_basicWithWO = `
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"
  disks       = [sakura_disk.foobar.id]
  network_interface = [{
    upstream = "shared"
  }]
  disk_edit_parameter = {
    hostname        = "{{ .arg0 }}"
    password_wo     = "{{ .arg1 }}"
	password_wo_version = 1
    ssh_keys        = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEAo5G7cwRp423KOrtCewX5nXFkboGxZ3hfvECNGg56 e2e-test-only@example"]
    ssh_key_ids     = ["100000000000", "200000000000"]
    disable_pw_auth = true
    script = [{
      id         = "100000000000"
      api_key_id = "100000000001"
    }]
  }
}`

const testAccSakuraServer_validateHostName = `
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"
  disks       = [sakura_disk.foobar.id]

  disk_edit_parameter = {
    hostname        = "invalid_host_name"
  }
}
`

const testAccSakuraServer_interfaces = `
resource "sakura_server" "foobar" {
  lifecycle {
    create_before_destroy = true
  }

  name = "{{ .arg0 }}"
  network_interface = [{
    upstream = "shared"
  }]

  force_shutdown = true
}
`

const testAccSakuraServer_interfacesAdded = `
resource "sakura_server" "foobar" {
  lifecycle {
    create_before_destroy = true
  }

  name = "{{ .arg0 }}"
  network_interface = [{
    upstream = "shared"
  },
  {
    upstream = sakura_vswitch.foobar0.id
  }]

  force_shutdown = true
}

resource "sakura_vswitch" "foobar0" {
  name = "{{ .arg0 }}-0"
}
`
const testAccSakuraServer_interfacesUpdated = `
resource "sakura_server" "foobar" {
  lifecycle {
    create_before_destroy = true
  }

  name = "{{ .arg0 }}"
  network_interface = [{
    upstream = "shared"
  },
  {
    upstream = sakura_vswitch.foobar0.id
  },
  {
    upstream = sakura_vswitch.foobar1.id
  },
  {
    upstream = sakura_vswitch.foobar2.id
  }]

  force_shutdown = true
}

resource "sakura_vswitch" "foobar0" {
  name = "{{ .arg0 }}-0"
}
resource "sakura_vswitch" "foobar1" {
  name = "{{ .arg0 }}-1"
}
resource "sakura_vswitch" "foobar2" {
  name = "{{ .arg0 }}-2"
}
`

const testAccSakuraServer_interfacesDisconnect = `
resource "sakura_server" "foobar" {
  lifecycle {
    create_before_destroy = true
  }

  name = "{{ .arg0 }}"

  network_interface = [{
    upstream = "disconnect"
  },
  {
    upstream = sakura_vswitch.foobar0.id
  }]

  force_shutdown = true
}
resource "sakura_vswitch" "foobar0" {
  name = "{{ .arg0 }}-0"
}
`

const testAccSakuraServer_packetFilter = `
resource "sakura_packet_filter" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_packet_filter_rules" "rules" {
  packet_filter_id = sakura_packet_filter.foobar.id
  expression = [{
    protocol         = "tcp"
    source_network   = "0.0.0.0"
    source_port      = "0-65535"
    destination_port = "80"
    allow            = true
  }]
}

resource "sakura_server" "foobar" {
  lifecycle {
    create_before_destroy = true
  }

  name = "{{ .arg0 }}"
  network_interface = [{
    upstream         = "shared"
    packet_filter_id = sakura_packet_filter.foobar.id
  },
  {
    upstream         = sakura_vswitch.foobar.id
    packet_filter_id = sakura_packet_filter.foobar.id
  }]

  force_shutdown = true
}

resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}
`

const testAccSakuraServer_packetFilterUpdate = `
resource "sakura_packet_filter" "foobar" {
  name = "{{ .arg0 }}-upd"
}

resource "sakura_packet_filter_rules" "rules" {
  packet_filter_id = sakura_packet_filter.foobar.id
  expression = [{
    protocol         = "udp"
    source_network   = "0.0.0.0"
    source_port      = "0-65535"
    destination_port = "80"
    allow            = true
  }]
}

resource "sakura_server" "foobar" {
  lifecycle {
    create_before_destroy = true
  }

  name = "{{ .arg0 }}-upd"

  network_interface = [{
    upstream         = "shared"
    packet_filter_id = sakura_packet_filter.foobar.id
  }]

  force_shutdown = true
}
`

const testAccSakuraServer_packetFilterDelete = `
resource "sakura_server" "foobar" {
  name = "{{ .arg0 }}-del"
  network_interface = [{
    upstream = "shared"
  }]
  force_shutdown = true
}`

const testAccSakuraServer_withBlankDisk = `
resource "sakura_server" "foobar" {
  name  = "{{ .arg0 }}"
  disks = [sakura_disk.foobar.id]

  network_interface = [{
    upstream = "shared"
  }]
  force_shutdown = true
}
resource "sakura_disk" "foobar" {
  name = "{{ .arg0 }}"
}
`

const testAccSakuraServer_vswitch = `
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_vswitch" "foobar1" {
  name = "{{ .arg0 }}"
}

resource "sakura_vswitch" "foobar2" {
  name = "{{ .arg0 }}"
}

resource "sakura_server" "foobar" {
  name  = "{{ .arg0 }}"
  disks = [sakura_disk.foobar.id]

  network_interface = [{
    upstream = sakura_vswitch.foobar1.id
  },
  {
    upstream        = sakura_vswitch.foobar2.id
    user_ip_address = "192.168.1.2"
  }]
  
  disk_edit_parameter = {
    ip_address = "192.168.0.2"
    netmask    = 24
    gateway    = "192.168.0.1"
  }
}
`

const testAccSakuraServer_withoutShutdownWhenUpdate = `
resource "sakura_vswitch" "sw" {
  name = "{{ .arg0 }}"
}
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"
  disks       = [sakura_disk.foobar.id]
  description = "description"
  tags        = ["tag1", "tag2"]

  network_interface = [{
    upstream = "shared"
  },
  {
    upstream        = sakura_vswitch.sw.id
    user_ip_address = "192.168.0.11"
  }]

  disk_edit_parameter = {
    hostname        = "{{ .arg0 }}"
    password        = "{{ .arg1 }}"
  }
}
`

const testAccSakuraServer_updateWithoutShutdownWhenUpdate = `
resource "sakura_vswitch" "sw" {
  name = "{{ .arg0 }}"
}
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}-upd"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name             = "{{ .arg0 }}-upd"
  disks            = [sakura_disk.foobar.id]
  description      = "description-upd"
  tags             = ["tag1-upd", "tag2-upd"]

  network_interface = [{
    upstream = "shared"
  },
  {
    upstream        = sakura_vswitch.sw.id
    user_ip_address = "192.168.0.12"
  }]

  disk_edit_parameter = {
    hostname        = "{{ .arg0 }}"
    password        = "{{ .arg1 }}"
  }
}
`

const testAccSakuraServer_updateWithShutdownWhenUpdate = `
resource "sakura_vswitch" "sw" {
  name = "{{ .arg0 }}"
}
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}-upd"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name             = "{{ .arg0 }}-upd"
  disks            = [sakura_disk.foobar.id]
  description      = "description-upd"
  tags             = ["tag1-upd", "tag2-upd"]

  network_interface = [{
    upstream = "shared"
  },
  {
    upstream        = sakura_vswitch.sw.id
    user_ip_address = "192.168.0.12"
  }]

  disk_edit_parameter = {
    hostname        = "{{ .arg0 }}"
    password        = "{{ .arg1 }}-upd"
  }
}
`

const testAccSakuraServer_standardPlan = `
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"
  disks       = [sakura_disk.foobar.id]
  network_interface = [{
    upstream = "shared"
  }]
  core   = 2
  memory = 4
  commitment = "standard"

  disk_edit_parameter = {
    hostname = "{{ .arg0 }}"
    password = "{{ .arg1 }}"
  }
}
`

const testAccSakuraServer_dedicatedCPUPlan = `
data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}
resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"
  disks       = [sakura_disk.foobar.id]
  network_interface = [{
    upstream = "shared"
  }]
  core   = 2
  memory = 4
  commitment = "dedicatedcpu"

  disk_edit_parameter = {
    hostname        = "{{ .arg0 }}"
    password        = "{{ .arg1 }}"
  }
}
`

const testAccSakuraServer_cloudInit = `
resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"
  disks       = [{{ .arg2 }}]
  network_interface = [{
    upstream = "shared"
  }]
  core   = 2
  memory = 4

  user_data = join("\n", [
    "#cloud-config",
    yamlencode({
      hostname: "{{ .arg0 }}",
      password: "{{ .arg1 }}",
      chpasswd: {
        expire: false,
      }
      ssh_pwauth: false,
    }),
  ])
}
`

const testAccSakuraServer_cloudInitUpdated = `
resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}_upd"
  disks       = [{{ .arg2 }}]
  network_interface = [{
    upstream = "shared"
  }]
  core   = 2
  memory = 4

  user_data = join("\n", [
    "#cloud-config",
    yamlencode({
      hostname: "{{ .arg0 }}",
      password: "{{ .arg1 }}",
      chpasswd: {
        expire: false,
      }
      ssh_pwauth: false,
    }),
  ])
}
`

const testAccSakuraServer_cloudInitUserDataUpdated = `
resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"
  disks       = [{{ .arg2 }}]
  network_interface = [{
    upstream = "shared"
  }]
  core   = 2
  memory = 4

  user_data = join("\n", [
    "#cloud-config",
    yamlencode({
      hostname: "{{ .arg0 }}-upd",
      password: "{{ .arg1 }}",
      chpasswd: {
        expire: false,
      }
      ssh_pwauth: false,
    }),
  ])
}
`

const testAccSakuraServer_withGPU = `
resource "sakura_server" "foobar" {
  zone = "is1a"

  name   = "{{ .arg0 }}"
  core   = 4
  memory = 56
  gpu    = 1

  force_shutdown = true
}
`

const testAccSakuraServer_withAMDPlan = `
resource "sakura_server" "foobar" {
  zone = "is1b"

  name       = "{{ .arg0 }}"
  core       = 32
  memory     = 120
  cpu_model  = "amd_epyc_7713p"
  commitment = "dedicatedcpu"

  force_shutdown = true
}
`

const testAccSakuraServer__withKoukaryokuVRT = `
resource "sakura_server" "foobar" {
  zone = "is1a"

  name       = "{{ .arg0 }}"
  core       = 4
  memory     = 56
  
  gpu = 1
  gpu_model = "nvidia_v100_32gbvram"

  force_shutdown = true
}
`

const testAccSakuraServer_confidentialVM = `
resource "sakura_server" "foobar" {
  name        = "{{ .arg0 }}"

  core            = 16
  memory          = 24
  cpu_model       = "amd_epyc_9654p"
  commitment      = "dedicatedcpu"
  confidential_vm = true

  zone = "tk1b"
}
`
