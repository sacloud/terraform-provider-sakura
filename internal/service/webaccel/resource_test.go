// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	"github.com/sacloud/webaccel-api-go"
)

const (
	envWebAccelSiteName             = "SAKURA_WEBACCEL_SITE_NAME"
	envWebAccelDomainName           = "SAKURA_WEBACCEL_DOMAIN_NAME"
	envWebAccelOrigin               = "SAKURA_WEBACCEL_ORIGIN"
	envObjectStorageEndpoint        = "SAKURA_OBJECT_STORAGE_ENDPOINT"
	envObjectStorageRegion          = "SAKURA_OBJECT_STORAGE_REGION"
	envObjectStorageBucketName      = "SAKURA_OBJECT_STORAGE_BUCKET_NAME"
	envObjectStorageAccessKeyId     = "SAKURA_OBJECT_STORAGE_ACCESS_KEY"
	envObjectStorageSecretAccessKey = "SAKURA_OBJECT_STORAGE_ACCESS_SECRET"
)

func TestAccSakuraResourceWebAccel_WebOrigin(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	envKeys := []string{
		envWebAccelOrigin,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := "your-site-name"
	// domainName := os.Getenv(envWebAccelDomainName)
	origin := os.Getenv(envWebAccelOrigin)
	regexpNotEmpty := regexp.MustCompile(".+")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelWebOriginConfigBasic(siteName, origin),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "request_protocol", "https-redirect"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "web"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.origin", origin),
					resource.TestMatchResourceAttr("sakura_webaccel.foobar", "cname_record_value", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel.foobar", "txt_record_value", regexpNotEmpty),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "normalize_ae", "br+gzip"),
				),
			},
		},
	})
}

func testCheckSakuraWebAccelDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := webaccel.NewOp(client.WebaccelClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_webaccel" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists WebAccel site: %s", rs.Primary.ID)
		}
	}
	return nil
}

func TestAccSakuraResourceWebAccel_OwnDomain(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	envKeys := []string{
		envWebAccelOrigin,
		envWebAccelDomainName,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := "your-site-name"
	origin := os.Getenv(envWebAccelOrigin)
	domainName := os.Getenv(envWebAccelDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelOwnDomainConfig(siteName, origin, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "domain_type", "own_domain"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "domain", domainName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "request_protocol", "http+https"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "web"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.origin", origin),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.host_header", origin),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.protocol", "https"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWebAccel_WebOriginWithOneTimeUrlSecrets(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	envKeys := []string{
		envWebAccelOrigin,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := "your-site-name"
	origin := os.Getenv(envWebAccelOrigin)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelWebOriginConfigWithOneTimeUrlSecrets(siteName, origin),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "web"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.origin", origin),
					resource.TestCheckNoResourceAttr("sakura_webaccel.foobar", "onetime_url_secrets"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "onetime_url_secrets_version", "1"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWebAccel_WebOriginWithCORS(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	envKeys := []string{
		envWebAccelOrigin,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := "your-site-name"
	// domainName := os.Getenv(envWebAccelDomainName)
	origin := os.Getenv(envWebAccelOrigin)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelWebOriginConfigWithCors(siteName, origin),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "web"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.origin", origin),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "cors_rules.0.allow_all", "false"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "cors_rules.0.allowed_origins.0", "https://apps.example.com"),
					resource.TestCheckNoResourceAttr("sakura_webaccel.foobar", "onetime_url_secrets"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "onetime_url_secrets_version", "1"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "normalize_ae", "gzip"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWebAccel_Update(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	envKeys := []string{
		envWebAccelOrigin,
		envObjectStorageEndpoint,
		envObjectStorageRegion,
		envObjectStorageBucketName,
		envObjectStorageAccessKeyId,
		envObjectStorageSecretAccessKey,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := "your-site-name"
	// domainName := os.Getenv(envWebAccelDomainName)
	origin := os.Getenv(envWebAccelOrigin)
	endpoint, _ := strings.CutPrefix(os.Getenv(envObjectStorageEndpoint), "https://")
	region := os.Getenv(envObjectStorageRegion)
	bucketName := os.Getenv(envObjectStorageBucketName)
	accessKey := os.Getenv(envObjectStorageAccessKeyId)
	secretKey := os.Getenv(envObjectStorageSecretAccessKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelWebOriginConfigBasic(siteName, origin),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "web"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.origin", origin),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "request_protocol", "https-redirect"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "vary_support", "true"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "normalize_ae", "br+gzip"),
				),
			},
			{
				Config: testAccCheckSakuraWebAccelWebOriginConfigWithCors(siteName, origin),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "web"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.origin", origin),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "request_protocol", "http+https"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "cors_rules.0.allow_all", "false"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "cors_rules.0.allowed_origins.0", "https://apps.example.com"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "normalize_ae", "gzip"),
				),
			},
			{
				Config: testAccCheckSakuraWebAccelWebOriginLoggingConfig(siteName, origin, endpoint, region, bucketName, accessKey, secretKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "web"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.origin", origin),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "logging.bucket_name", bucketName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "vary_support", "true"),
				),
			},
			{
				Config: testAccCheckSakuraWebAccelBucketOriginConfig(siteName, endpoint, region, bucketName, accessKey, secretKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "bucket"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.endpoint", endpoint),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.region", region),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.bucket_name", bucketName),
					resource.TestCheckNoResourceAttr("sakura_webaccel.foobar", "origin_parameters.access_key"),
					resource.TestCheckNoResourceAttr("sakura_webaccel.foobar", "origin_parameters.secret_access_key"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.credentials_version", "1"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWebAccel_BucketOrigin(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	envKeys := []string{
		envWebAccelOrigin,
		envObjectStorageEndpoint,
		envObjectStorageRegion,
		envObjectStorageBucketName,
		envObjectStorageAccessKeyId,
		envObjectStorageSecretAccessKey,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := "your-site-name"
	// domainName := os.Getenv(envWebAccelDomainName)
	endpoint, _ := strings.CutPrefix(os.Getenv(envObjectStorageEndpoint), "https://")
	region := os.Getenv(envObjectStorageRegion)
	bucketName := os.Getenv(envObjectStorageBucketName)
	accessKey := os.Getenv(envObjectStorageAccessKeyId)
	secretKey := os.Getenv(envObjectStorageSecretAccessKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelBucketOriginConfig(siteName, endpoint, region, bucketName, accessKey, secretKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "bucket"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.endpoint", endpoint),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.region", region),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.bucket_name", bucketName),
					resource.TestCheckNoResourceAttr("sakura_webaccel.foobar", "origin_parameters.access_key"),
					resource.TestCheckNoResourceAttr("sakura_webaccel.foobar", "origin_parameters.secret_access_key"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.credentials_version", "1"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.use_document_index", "true"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWebAccel_Logging(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	envKeys := []string{
		envWebAccelOrigin,
		envObjectStorageBucketName,
		envObjectStorageAccessKeyId,
		envObjectStorageSecretAccessKey,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := "your-site-name"
	origin := os.Getenv(envWebAccelOrigin)
	bucketName := os.Getenv(envObjectStorageBucketName)
	accessKey := os.Getenv(envObjectStorageAccessKeyId)
	secretKey := os.Getenv(envObjectStorageSecretAccessKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelWebOriginLoggingConfig(siteName, origin, "s3.isk01.sakurastorage.jp", "jp-north-1", bucketName, accessKey, secretKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "name", siteName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.type", "web"),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "origin_parameters.origin", origin),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "logging.bucket_name", bucketName),
					resource.TestCheckResourceAttr("sakura_webaccel.foobar", "vary_support", "true"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWebAccel_InvalidConfigurations(t *testing.T) {
	if os.Getenv(envWebAccelOrigin) == "" {
		t.Skipf("ENV %q is required. skip", envWebAccelOrigin)
		return
	}
	origin := os.Getenv(envWebAccelOrigin)
	for name, tc := range testAccCheckSakuraWebAccelInvalidConfigs(origin) {
		t.Logf("test for invalid configuration: %s", name)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
			CheckDestroy: func(*terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config: tc,
					ExpectError: func() *regexp.Regexp {
						return regexp.MustCompile(".")
					}(),
				},
			},
		})
	}
}

func testAccCheckSakuraWebAccelWebOriginConfigBasic(siteName string, origin string) string {
	tmpl := `
resource sakura_webaccel "foobar" {
  name = "%s"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "%s"
    protocol = "https"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`
	return fmt.Sprintf(tmpl, siteName, origin, origin)
}

func testAccCheckSakuraWebAccelOwnDomainConfig(siteName string, origin string, domain string) string {
	tmpl := `
resource sakura_webaccel "foobar" {
  name = "%s"
  domain_type = "own_domain"
  domain = "%s"
  request_protocol = "http+https"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "%s"
    protocol = "https"
  }
}
`
	return fmt.Sprintf(tmpl, siteName, domain, origin, origin)
}

func testAccCheckSakuraWebAccelWebOriginConfigWithOneTimeUrlSecrets(siteName string, origin string) string {
	tmpl := `
resource sakura_webaccel "foobar" {
  name = "%s"
  domain_type = "subdomain"
  request_protocol = "http+https"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "%s"
    protocol = "https"
  }
  onetime_url_secrets = [
    "sample-secret"
  ]
  onetime_url_secrets_version = 1
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "gzip"
}
`
	return fmt.Sprintf(tmpl, siteName, origin, origin)
}

func testAccCheckSakuraWebAccelWebOriginConfigWithCors(siteName string, origin string) string {
	tmpl := `
resource sakura_webaccel "foobar" {
  name = "%s"
  domain_type = "subdomain"
  request_protocol = "http+https"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "%s"
    protocol = "https"
  }
  cors_rules = [{
    allow_all = false
    allowed_origins = [
       "https://apps.example.com",
       "https://platform.example.com"
    ]
  }]
  onetime_url_secrets = [
    "sample-secret"
  ]
  onetime_url_secrets_version = 1
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "gzip"
}
`
	return fmt.Sprintf(tmpl, siteName, origin, origin)
}

func testAccCheckSakuraWebAccelBucketOriginConfig(siteName string, s3Endpoint string, region string, bucketName string, accessKey string, accessSecret string) string {
	tmpl := `
resource sakura_webaccel "foobar" {
  name = "%s"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "bucket"
    endpoint = "%s"
    region = "%s"
    bucket_name = "%s"
    access_key = "%s"
    secret_access_key = "%s"
	credentials_version = 1
    use_document_index = true
  }
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`
	return fmt.Sprintf(tmpl, siteName, s3Endpoint, region, bucketName, accessKey, accessSecret)
}

func testAccCheckSakuraWebAccelWebOriginLoggingConfig(siteName string, origin string, endpoint string, region string, bucketName string, accessKey string, secretKey string) string {
	tmpl := `
resource sakura_webaccel "foobar" {
  name = "%s"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "%s"
    protocol = "https"
  }
  logging = {
    enabled = true
	endpoint = "%s"
	region = "%s"
    bucket_name = "%s"
    access_key = "%s"
    secret_access_key = "%s"
	credentials_version = 1
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`
	return fmt.Sprintf(tmpl, siteName, origin, origin, endpoint, region, bucketName, accessKey, secretKey)
}

func testAccCheckSakuraWebAccelInvalidConfigs(origin string) map[string]string {
	confUnknownArgument := `
resource sakura_webaccel "foobar" {
  invalid = true
  name = "dummy1"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "dummy.example.com"
    protocol = "https"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	confInvalidDomainType := `
resource sakura_webaccel "foobar" {
  name = "dummy2"
  domain_type = "INVALID"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "dummy.example.com"
    protocol = "https"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	confInvalidRequestProtocol := `
resource sakura_webaccel "foobar" {
  name = "dummy3"
  domain_type = "subdomain"
  request_protocol = "http"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "dummy.example.com"
    protocol = "https"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`
	confWithoutOriginParameters := `
resource sakura_webaccel "foobar" {
  name = "dummy4"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	confInvalidOriginType := `
resource sakura_webaccel "foobar" {
  name = "dummy5"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "INVALID"
    origin = "%s"
    host_header = "dummy.example.com"
    protocol = "https"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	confLackingWebOriginParameters := `
resource sakura_webaccel "foobar" {
  name = "dummy6"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    host_header = "dummy.example.com"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	confMismatchedOriginParameters := `
resource sakura_webaccel "foobar" {
  name = "dummy7"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "bucket"
    host_header = "dummy.example.com"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	// config without the object storage's endpoint parameter
	confLackingBucketOriginParameters := `
resource sakura_webaccel "foobar" {
  name = "dummy8"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "bucket"
    region = "jp-sample-1"
    access_key = "sample"
    secret_access_key = "sample"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	confInvalidNormalizeAE := `
resource sakura_webaccel "foobar" {
  name = "dummy9"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "dummy.example.com"
    protocol = "https"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "INVALID"
}
`
	// config without the S3 secret access key for logging
	confMissingLoggingParameters := `
resource sakura_webaccel "foobar" {
  name = "dummy10"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    origin = "docs.usacloud.jp"
    protocol = "https"
  }
  logging = {
    bucket_name = "example-bucket"
    access_key = "sample"
  }
}
`
	// allow_all and allowed_origins should not be specified together
	confInvalidCorsConfiguration := `
resource sakura_webaccel "foobar" {
  name = "dummy11"
  domain_type = "subdomain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "dummy.example.com"
    protocol = "https"
  }
  cors_rules = [{
    allow_all = true
    allowed_origins = [
      "https://www2.example.com",
      "https://app.example.com"
    ]
  }]
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	confOwnDomainWithoutDomain := `
resource sakura_webaccel "foobar" {
  name = "dummy12"
  domain_type = "own_domain"
  request_protocol = "https-redirect"
  origin_parameters = {
    type = "web"
    origin = "%s"
    host_header = "dummy.example.com"
    protocol = "https"
  }
  vary_support = true
  default_cache_ttl = 3600
  normalize_ae = "br+gzip"
}
`

	tt := map[string]string{
		"unknown-argument":                         confUnknownArgument,
		"invalid-request-protocol":                 confInvalidRequestProtocol,
		"invalid-domain-type":                      confInvalidDomainType,
		"no-origin-params":                         confWithoutOriginParameters,
		"invalid-origin-type":                      confInvalidOriginType,
		"lacking-web-origin-params":                confLackingWebOriginParameters,
		"mismatched-bucket-origin-type-and-params": confMismatchedOriginParameters,
		"lacking-bucket-origin-params":             confLackingBucketOriginParameters,
		"invalid-compression":                      confInvalidNormalizeAE,
		"missing-logging-bucket-secret":            confMissingLoggingParameters,
		"invalid-cors-configuration":               confInvalidCorsConfiguration,
		"own-domain-without-domain":                confOwnDomainWithoutDomain,
	}
	for k, v := range tt {
		if strings.Contains(v, "%s") {
			tt[k] = fmt.Sprintf(tt[k], origin)
		}
	}

	return tt
}
