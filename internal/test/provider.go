// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sakura "github.com/sacloud/terraform-provider-sakura/internal/provider"
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
		var v = reflect.ValueOf(AccProvider).Elem()
		var c = v.FieldByName("client").Elem()
		return (*common.APIClient)(unsafe.Pointer(c.UnsafeAddr())) //nolint:gosec
	}
}

func AccPreCheck(t *testing.T) {
	requiredEnvs := []string{
		"SAKURA_ACCESS_TOKEN",
		"SAKURA_ACCESS_TOKEN_SECRET",
	}

	for _, env := range requiredEnvs {
		if v := os.Getenv(env); v == "" {
			t.Fatalf("%s must be set for acceptance tests", env)
		}
	}

	if v := os.Getenv("SAKURA_ZONE"); v == "" {
		os.Setenv("SAKURA_ZONE", testDefaultTargetZone) //nolint:errcheck,gosec
	}

	if v := os.Getenv("SAKURA_RETRY_MAX"); v == "" {
		os.Setenv("SAKURA_RETRY_MAX", testDefaultAPIRetryMax) //nolint:errcheck,gosec
	}

	if v := os.Getenv("SAKURA_RATE_LIMIT"); v == "" {
		os.Setenv("SAKURA_RATE_LIMIT", testDefaultAPIRateLimit) //nolint:errcheck,gosec
	}
}
