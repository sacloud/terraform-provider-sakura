// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enhanced_lb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

const (
	envEnhancedLBACMEDomain = "SAKURACLOUD_ENHANCED_LB_ACME_DOMAIN"
)

var elbDomain string

func TestAccSakuraEnhancedLBACME_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envEnhancedLBACMEDomain)

	rand := test.RandomName()
	subDomain := "acme-acctest1" + test.RandStringFromCharSet(5, "")

	elbDomain = os.Getenv(envEnhancedLBACMEDomain)

	var elb iaas.ProxyLB
	resourceName := "sakura_enhanced_lb_acme.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			test.CheckSakuraCloudDiskDestroy,
			test.CheckSakuraCloudDNSRecordDestroy,
			testCheckSakuraEnhancedLBDestroy,
			test.CheckSakuraCloudServerDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraEnhancedLBACME_basic, rand, elbDomain, subDomain),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraEnhancedLBExists("sakura_enhanced_lb.foobar", &elb),
					resource.TestCheckResourceAttr("sakura_enhanced_lb.foobar", "gzip", "true"),
					resource.TestCheckResourceAttr("sakura_enhanced_lb.foobar", "proxy_protocol", "true"),
					resource.TestCheckResourceAttr("sakura_enhanced_lb.foobar", "backend_http_keep_alive", "aggressive"),
					resource.TestCheckResourceAttr("sakura_enhanced_lb.foobar", "rule.#", "1"),
					//resource.TestCheckResourceAttr(resourceName, "certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate.common_name", subDomain+"."+elbDomain),
					resource.TestCheckResourceAttr(resourceName,
						"certificate.subject_alt_names",
						fmt.Sprintf("%s.%s, acme-acctest2.%s, acme-acctest3.%s", subDomain, elbDomain, elbDomain, elbDomain),
					),
				),
			},
		},
	})
}

var testAccSakuraEnhancedLBACME_basic = `
resource "sakura_enhanced_lb" "foobar" {
  name           = "{{ .arg0 }}"
  plan           = 100
  vip_failover   = true
  gzip           = true
  proxy_protocol = true

  backend_http_keep_alive = "aggressive"

  health_check = {
    protocol    = "http"
    delay_loop  = 10
    host_header = "usacloud.jp"
    path        = "/"
  }
  bind_port = [{
    proxy_mode = "http"
    port       = 80
  },
  {
    proxy_mode = "https"
    port       = 443
  }]
  server = [{
    ip_address = sakura_server.foobar.ip_address
    port       = 80
    group      = "group1"
  }]
  rule = [{
    host  = "www.usacloud.com"
    path  = "/"
    group = "group1"
  }]
}

resource "sakura_enhanced_lb_acme" "foobar" {
  enhanced_lb_id               = sakura_enhanced_lb.foobar.id
  accept_tos                   = true
  common_name                  = "{{ .arg2 }}.{{ .arg1 }}"
  subject_alt_names            = ["acme-acctest2.{{ .arg1 }}", "acme-acctest3.{{ .arg1 }}"]
  update_delay_sec             = 120
  get_certificates_timeout_sec = 300
}

data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}

resource "sakura_disk" "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource "sakura_server" "foobar" {
  name  = "{{ .arg0 }}"
  disks = [sakura_disk.foobar.id]
  network_interface {
    upstream = "shared"
  }
}

data "sakura_dns" "zone" {
    name = "{{ .arg1 }}"
}

resource "sakura_dns_record" "record" {
  dns_id = data.sakura_dns.zone.id
  name   = "{{ .arg2 }}"
  type   = "CNAME"
  value  = "${sakura_enhanced_lb.foobar.fqdn}."
  ttl    = 10
}
resource "sakura_dns_record" "record2" {
  dns_id = data.sakura_dns.zone.id
  name   = "acme-acctest2"
  type   = "CNAME"
  value  = "${sakura_enhanced_lb.foobar.fqdn}."
  ttl    = 10
}
resource "sakura_dns_record" "record3" {
  dns_id = data.sakura_dns.zone.id
  name   = "acme-acctest3"
  type   = "CNAME"
  value  = "${sakura_enhanced_lb.foobar.fqdn}."
  ttl    = 10
}
`
