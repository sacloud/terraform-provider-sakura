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

package validator

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	IPv4 = iota
	IPv6
	Both
)

// 空文字列を許容する必要があるなど、terraform-plugin-framework-nettypesのIPAddressが使えない場合の代替
type stringIPAddressValidator struct {
	ipVersion int
}

var _ validator.String = stringIPAddressValidator{}

func (v stringIPAddressValidator) Description(_ context.Context) string {
	var format string
	switch v.ipVersion {
	case IPv4:
		format = "IPv4"
	case IPv6:
		format = "IPv6"
	default:
		format = "IP"
	}
	return "string must be a valid " + format + " address"
}

func (v stringIPAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIPAddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		// さくらの場合、空文字は許容する
		return
	}

	ipAddr, err := netip.ParseAddr(value)
	if err != nil || !ipAddr.IsValid() {
		resp.Diagnostics.AddAttributeError(req.Path, v.Description(ctx), fmt.Sprintf("%q is not a valid IP address: %s", value, err))
		return
	}

	if v.ipVersion == IPv4 && !ipAddr.Is4() {
		resp.Diagnostics.AddAttributeError(req.Path, v.Description(ctx), fmt.Sprintf("%q is not a valid IPv4 address", value))
		return
	}

	if v.ipVersion == IPv6 && !ipAddr.Is6() {
		resp.Diagnostics.AddAttributeError(req.Path, v.Description(ctx), fmt.Sprintf("%q is not a valid IPv6 address", value))
		return
	}
}

func IPAddressValidator(ipVersion int) stringIPAddressValidator {
	return stringIPAddressValidator{
		ipVersion: ipVersion,
	}
}
