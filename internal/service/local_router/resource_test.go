// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package local_router_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraLocalRouter_basic(t *testing.T) {
	if !test.IsResourceRequiredTest() {
		t.Skip("This test only run if SAKURA_RESOURCE_REQUIRED_TEST environment variable is set")
	}

	resourceName := "sakura_local_router.foobar"
	rand := test.RandomName()
	var localRouter iaas.LocalRouter
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			test.CheckSakuraIconDestroy,
			testCheckSakuraLocalRouterDestroy,
			test.CheckSakuravSwitchDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraLocalRouter_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraLocalRouterExists(resourceName, &localRouter),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttrPair(
						resourceName, "switch.code",
						"sakura_vswitch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "switch.category", "cloud"),
					resource.TestCheckResourceAttrPair(
						resourceName, "switch.zone",
						"data.sakura_zone.current", "name"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.vip", "192.168.11.1"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.0", "192.168.11.11"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.1", "192.168.11.12"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.netmask", "24"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.vrid", "1"),
					resource.TestCheckResourceAttr(resourceName, "static_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "static_route.0.prefix", "10.0.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "static_route.0.next_hop", "192.168.11.2"),
					resource.TestCheckResourceAttrPair(
						resourceName, "icon_id",
						"sakura_icon.foobar", "id",
					),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraLocalRouter_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraLocalRouterExists(resourceName, &localRouter),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2-upd"),
					resource.TestCheckResourceAttrPair(
						resourceName, "switch.code",
						"sakura_vswitch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "switch.category", "cloud"),
					resource.TestCheckResourceAttrPair(
						resourceName, "switch.zone",
						"data.sakura_zone.current", "name"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.vip", "192.168.21.1"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.0", "192.168.21.11"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.ip_addresses.1", "192.168.21.12"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.netmask", "24"),
					resource.TestCheckResourceAttr(resourceName, "network_interface.vrid", "2"),
					resource.TestCheckResourceAttr(resourceName, "static_route.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "static_route.0.prefix", "10.0.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "static_route.0.next_hop", "192.168.21.2"),
					resource.TestCheckResourceAttr(resourceName, "static_route.1.prefix", "172.16.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "static_route.1.next_hop", "192.168.21.3"),
				),
			},
		},
	})
}

func testCheckSakuraLocalRouterExists(n string, localRouter *iaas.LocalRouter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no LocalRouter ID is set")
		}

		client := test.AccClientGetter()
		lrOp := iaas.NewLocalRouterOp(client)

		foundLocalRouter, err := lrOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundLocalRouter.ID.String() != rs.Primary.ID {
			return fmt.Errorf("not found LocalRouter: %s", rs.Primary.ID)
		}

		*localRouter = *foundLocalRouter
		return nil
	}
}

func testCheckSakuraLocalRouterDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	lrOp := iaas.NewLocalRouterOp(client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_local_router" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := lrOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists LocalRouter: %s", rs.Primary.ID)
		}
	}

	return nil
}

func TestAccImportSakuraLocalRouter_basic(t *testing.T) {
	if !test.IsResourceRequiredTest() {
		t.Skip("This test only run if SAKURA_RESOURCE_REQUIRED_TEST environment variable is set")
	}

	rand := test.RandomName()
	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                             rand,
			"description":                      "description",
			"tags.0":                           "tag1",
			"tags.1":                           "tag2",
			"switch.category":                  "cloud",
			"switch.zone":                      os.Getenv("SAKURACLOUD_ZONE"),
			"network_interface.vip":            "192.168.11.1",
			"network_interface.ip_addresses.0": "192.168.11.11",
			"network_interface.ip_addresses.1": "192.168.11.12",
			"network_interface.netmask":        "24",
			"network_interface.vrid":           "1",
			"static_route.0.prefix":            "10.0.0.0/24",
			"static_route.0.next_hop":          "192.168.11.2",
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0], "icon_id", "switch.code")
	}

	resourceName := "sakura_local_router.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraLocalRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraLocalRouter_basic, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccSakuraLocalRouter_basic = `
resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_zone" "current" {}

resource "sakura_local_router" "foobar" {
  switch = {
    code     = sakura_vswitch.foobar.id
    category = "cloud"
    zone     = data.sakura_zone.current.name
  }
  network_interface = {
    vip          = "192.168.11.1"
    ip_addresses = ["192.168.11.11", "192.168.11.12"]
    netmask      = 24
    vrid         = 1
  }
  static_route = [{
    prefix   = "10.0.0.0/24"
    next_hop = "192.168.11.2"
  }]

  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
  icon_id     = sakura_icon.foobar.id
}

resource "sakura_icon" "foobar" {
  name          = "{{ .arg0 }}"
  base64content = "iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAIAAADYYG7QAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAAAZiS0dEAP8A/wD/oL2nkwAAAAlwSFlzAAALEwAACxMBAJqcGAAACdBJREFUWMPNmHtw1NUVx8+5v9/+9rfJPpJNNslisgmIiCCgDQZR5GWnilUDPlpUqjOB2mp4qGM7tVOn/yCWh4AOVUprHRVB2+lMa0l88Kq10iYpNYPWkdeAmFjyEJPN7v5+v83ec/rH3Q1J2A2Z1hnYvz755ZzzvXPPveeee/GbC24FJmZGIYD5QgPpTBIAAICJLgJAwUQMAIDMfOEBUQchgJmAEC8CINLPThpfFCAG5orhogCBQiAAEyF8PQCATEQyxQzMzFIi4Ojdv86UEVF/f38ymezv7yciANR0zXAZhuHSdR0RRxNHZyJEBERmQvhfAAABIJlMJhIJt9t9TXX11GlTffleQGhvbz/4YeuRw4c13ZWfnycQR9ACQEShAyIxAxEKMXoAIVQ6VCzHcSzLmj937qqVK8aNrYKhv4bGxue3bvu8rc3n9+ualisyMzOltMjYccBqWanKdD5gBgAppZNMJhKJvlgs1heLxWL3fPfutU8/VVhYoGx7e3uJyOVyAcCEyy6bN2d266FDbW3thsuFI0gA4qy589PTOJC7EYEBbNu2ElYg4J9e/Y3p1dWBgN+l67csWKBC/mrbth07dnafOSMQp0y58pEVK2tm1ABAW9vn93zvgYRl5+XlAXMuCbxh3o3MDMyIguE8wADRaJ/H7Vp873119y8JBALDsrN8xcpXX3utoKDQNE1iiEV7ieSzmzYuXrwYAH7z4m83bNocDAZ1Tc8hQThrzjwYxY8BmCjaF/P78n+xZs0Ns64f+Ndnn53yevOLioo2btq8bsOGsvAYn9eHAoFZStnR0aFpWsObfxw/fvzp06fvXnyvZVmmx4M5hHQa3S4DwIRlm4Zr7dNPz7r+OgDo6el5bsuWtxrf6u7u9njygsHC9i/+U1Ia9ubnMzATA7MQIlRS8tnJk3/e1fDoI6vKysoqK8pbP/q323RDdi2hq/0ysHGyAwopU4lEfNXKlWo0Hx069MDSZcePHy8MBk3Tk0ylTnd1+wsKTNMERLUGlLtA1A3jyNEjagIKgsFk0gEM5NCSOst0+wEjAEvHtktKSuoeWAIAX3311f11Szs7OydcPtFwGYDp0sagWhoa7K4G5/f71TfHskEVdHXMn6M16CzLDcRkWfaM6dWm6QGAjZs2t7W1X1JeYRgGMzERMxOnNYa5O8mkrmkzr50JAKlUqq29Le2VQ0sACmYmIvU1OwAmLKt6ejUAyJTcu3dfQTCoaZqUkgEoY0ODvKRMSWbLsjo6O2fPmbuw9nYAOHjw4KdHjhqGoRqgLFpS6oNOE84JRDLVX1FeDgBd3V0pIrfLxZn5GGLMrE40y7YTCcula7W3167++c+UzfNbtzGRK+ObxR1RZyJARPUpNxBzPBYDAE3ThCYkETMjIPMQdwCwbNttGItqb6uqrJo2deqMGTVK8qWXX969+92SsjAi5hRF1BkQKJ3REUDXtE+PHL3ppptCoVBpcXFXVzdJqerFWWNmKaVt2T9YWldf//Dg6rL52efWrV/vCxQYLhdJmV2LmaUUkEkZZGbvXGBm0+P563vvqT/vW7LEcRwnmUxv7wFjZiYyDJdabQCQSsnt27d/6+YFT61Z4/UHBvZadi1mQBRERMwEMAIwkdttNh/8V2trKwB85647a2tv7+npTfb3y6HGKLREIvHKK6+my66ubd/x+p69+0KlZf5AQKV+BC0G0MaURwZGlxMAiam9vf3YsWNL7rsXAL694Oa2tvZPPvnEZRiozBABAIE1XfvggwMfffzxnXcsAoBrZ8zYs3+/pmm6ECNJIKrto4UvueQ8pxiRZduxWKympuauRQsnT56saRoAlIRCbzbsYmYhxGB7TdPcHk9LS3O4LHz1VVcFg8HmpubjJ0643W44/w8FS6kqW1YgKROW5VjWivr6P/3h93V1dYZhKNeD/2zp7elVjfAQLyKP2+0PFG5/NZ242XNm25bNRCNrKUjfy5gIzwXE/mQyEYs98dMnHnrw+yr6hx+2/qOp6djRo43vvGu4XJquZ3X3mO7OL8+cOnUqEolURSpUx53LeDDolDlE+ByQRNG+vlmzZ6vROI69fMWqN954Ix5PBAoLC4PBfK+XMqfSEHdEQJRS2ratyl1KSmLG3FoDoKcXFCIQDQOZTCLAQ8uWKtNlD/5w546dkaqqKq8XERDFQIkb7g6QSqUK/f5wOAwA0WgUiM+u/WxaChBRJxSgzsXhK5+sZDISiVxTUwMAjY2Nu3Y1RMZd6vXmAzCAIOB0uHP2SyqVisViCxcu9Pl8ANDc0oK6xswkxMg7mon0dGHMUqkg6Tjh0lLTdAPABwf+niKZ5zFRtRmQ8RrqyACyv783Gi0vL390eb0qqm+/szvPNNMzNGIFRnUvA0SAzOwNAiLJmU4zHo8DCgAgZgAETtswyX4pk8lkehP0pywrUTV27JaNGyqrKgHgha1bT548WRYOMwDk1hrIna46gbTAUBBCUwcqAFw6frwuRCqV0nUdmFB1MCRtx9E0bWwkEresRDzu9/nm3Th/Vf3DoVAIAJqbmtauXZfv9WpCpBd7Dq00EOGkKdNylCi0EgkhxP4971ZUVJw8ceK2RXd0dX9ZUFCgCaFyYTtOrC/22CMrf/LjH3V0dvX1RSsjEVemUDU3NS1d9uAXHR2lpaVqV4+iMIJWXFKKiEpgCCAKxI6OjuLioutmziwoLBxTFn7r7Xei0WhKSsdxYvF4PJ649Zabn1m/DhC93vxgMKiKuGUlntm46bHHHz/T0xsqKdEEZpYKZ9caJIpXTJmWfuVDofpPBcAMKKLRXoHwl727x106HgAOHDiw5ZcvHD5ymBiCwcJFtbXLM21GQ0ODZVm90ej77/9t3779XV2dBcEifyCgIcLQyCMBMU6cNCX3wQIkqbOzY+LlE373+s6KSER97untdSy7tKx0wHD16tVPPvkkAIDQvV6fz+fNz/emXzyAYVS5yqSsqLh4UM8GwwAFmqZ54sSJXY2NJSUlkyZNAgDTNL1er/Jvb29/uL7+1y++VFQcKg2PCYVCfr/XND1C01QnnytydkDECVdcqdpqtXGGgcqulHTmy+54PH71VdNunD+/sqoSEaPRaEtzy569exO2UxQM5nm9ynpQgrIEPA8w42UTJ6dLEkNWUI0KMTu2E4v3xftiSccGAKHpnrw8v8/vyfPoug4Zv1xxRgOIoDNJQAEMmfo9HNT9DxFN03QbRrCwCNQjHAp1gVc2mQKbM86oAFCA0GDQnSEXqMcGwPQjmND1zGgEAFBmNOeNMzIQSZ0GXvJHuJedPXRkLhiN+2hAVxUdz77yXWDQUdMGFUa40DC4Y/ya5vz/BMEkmVm9dl94QPwvNJB+oilXgHEAAAAldEVYdGRhdGU6Y3JlYXRlADIwMTYtMDItMTBUMjE6MDg6MzMtMDg6MDB4P0OtAAAAJXRFWHRkYXRlOm1vZGlmeQAyMDE2LTAyLTEwVDIxOjA4OjMzLTA4OjAwCWL7EQAAAABJRU5ErkJggg=="
}
`

const testAccSakuraLocalRouter_update = `
resource "sakura_vswitch" "foobar" {
  name = "{{ .arg0 }}"
}

data "sakura_zone" "current" {}

resource "sakura_local_router" "foobar" {
  switch = {
    code     = sakura_vswitch.foobar.id
    category = "cloud"
    zone     = data.sakura_zone.current.name
  }
  network_interface = {
    vip          = "192.168.21.1"
    ip_addresses = ["192.168.21.11", "192.168.21.12"]
    netmask      = 24
    vrid         = 2
  }
  static_route = [{
    prefix   = "10.0.0.0/24"
    next_hop = "192.168.21.2"
  },
  {
    prefix   = "172.16.0.0/16"
    next_hop = "192.168.21.3"
  }]

  name        = "{{ .arg0 }}-upd"
  description = "description-upd"
  tags        = ["tag1-upd", "tag2-upd"]
}
`

func TestAccSakuraLocalRouter_peering(t *testing.T) {
	if !test.IsResourceRequiredTest() {
		t.Skip("This test only run if SAKURA_RESOURCE_REQUIRED_TEST environment variable is set")
	}

	resourceName1 := "sakura_local_router.foobar1"
	resourceName2 := "sakura_local_router.foobar2"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			test.CheckSakuraIconDestroy,
			testCheckSakuraLocalRouterDestroy,
			test.CheckSakuravSwitchDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraLocalRouter_peering, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName1, "id", resourceName2, "peer.0.peer_id"),
					resource.TestCheckResourceAttrPair(resourceName1, "secret_keys.0", resourceName2, "peer.0.secret_key"),
					resource.TestCheckResourceAttr(resourceName2, "peer.0.description", "description"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraLocalRouter_peeringDisconnect, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "peer.#", "0"),
					resource.TestCheckResourceAttr(resourceName2, "peer.#", "0"),
				),
			},
		},
	})
}

const testAccSakuraLocalRouter_peering = `
resource "sakura_vswitch" "foobar1" {
  name = "{{ .arg0 }}"
}
resource "sakura_vswitch" "foobar2" {
  name = "{{ .arg0 }}"
}

data "sakura_zone" "current" {}

resource "sakura_local_router" "foobar1" {
  switch = {
    code     = sakura_vswitch.foobar1.id
    category = "cloud"
    zone     = data.sakura_zone.current.name
  }
  network_interface = {
    vip          = "192.168.11.1"
    ip_addresses = ["192.168.11.11", "192.168.11.12"]
    netmask      = 24
    vrid         = 1
  }

  name        = "{{ .arg0 }}"
}

resource "sakura_local_router" "foobar2" {
  switch = {
    code     = sakura_vswitch.foobar2.id
    category = "cloud"
    zone     = data.sakura_zone.current.name
  }
  network_interface = {
    vip          = "192.168.12.1"
    ip_addresses = ["192.168.12.11", "192.168.12.12"]
    netmask      = 24
    vrid         = 1
  }
  peer = [{
    peer_id     = sakura_local_router.foobar1.id
    secret_key  = sakura_local_router.foobar1.secret_keys.0
    description = "description"
  }]

  name        = "{{ .arg0 }}"
}
`

const testAccSakuraLocalRouter_peeringDisconnect = `
resource "sakura_vswitch" "foobar1" {
  name = "{{ .arg0 }}"
}
resource "sakura_vswitch" "foobar2" {
  name = "{{ .arg0 }}"
}

data "sakura_zone" "current" {}

resource "sakura_local_router" "foobar1" {
  switch = {
    code     = sakura_vswitch.foobar1.id
    category = "cloud"
    zone     = data.sakura_zone.current.name
  }
  network_interface = {
    vip          = "192.168.11.1"
    ip_addresses = ["192.168.11.11", "192.168.11.12"]
    netmask      = 24
    vrid         = 1
  }

  name        = "{{ .arg0 }}"
}

resource "sakura_local_router" "foobar2" {
  switch = {
    code     = sakura_vswitch.foobar2.id
    category = "cloud"
    zone     = data.sakura_zone.current.name
  }
  network_interface = {
    vip          = "192.168.12.1"
    ip_addresses = ["192.168.12.11", "192.168.12.12"]
    netmask      = 24
    vrid         = 1
  }

  name        = "{{ .arg0 }}"
}
`
