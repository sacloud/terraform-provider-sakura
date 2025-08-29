// Copyright 2016-2025 terraform-provider-sakuracloud authors
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

package test

import (
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	sakura "github.com/sacloud/terraform-provider-sakuracloud/internal/provider"
)

var (
	testDefaultTargetZone   = "is1b"
	testDefaultAPIRetryMax  = "30"
	testDefaultAPIRateLimit = "5"
)

var (
	AccProvider                 provider.Provider
	AccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
	AccClientGetter             func() *common.APIClient
)

func init() {
	// APIClientをテストで利用するためのAccProvider
	AccProvider = sakura.New("test")()
	AccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"sakura": providerserver.NewProtocol6WithError(AccProvider),
	}
	AccClientGetter = func() *common.APIClient {
		var v reflect.Value = reflect.ValueOf(AccProvider).Elem()
		var c reflect.Value = v.FieldByName("client").Elem()
		return (*common.APIClient)(unsafe.Pointer(c.UnsafeAddr()))
	}
}

func AccPreCheck(t *testing.T) {
	requiredEnvs := []string{
		"SAKURACLOUD_ACCESS_TOKEN",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET",
	}

	for _, env := range requiredEnvs {
		if v := os.Getenv(env); v == "" {
			t.Fatalf("%s must be set for acceptance tests", env)
		}
	}

	if v := os.Getenv("SAKURACLOUD_ZONE"); v == "" {
		os.Setenv("SAKURACLOUD_ZONE", testDefaultTargetZone) //nolint:errcheck,gosec
	}

	if v := os.Getenv("SAKURACLOUD_RETRY_MAX"); v == "" {
		os.Setenv("SAKURACLOUD_RETRY_MAX", testDefaultAPIRetryMax) //nolint:errcheck,gosec
	}

	if v := os.Getenv("SAKURACLOUD_RATE_LIMIT"); v == "" {
		os.Setenv("SAKURACLOUD_RATE_LIMIT", testDefaultAPIRateLimit) //nolint:errcheck,gosec
	}
}
