// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
)

type frontDoorOriginModel struct {
	Hostname   types.String `tfsdk:"hostname"`
	HostHeader types.String `tfsdk:"host_header"`
}

var pricingLevelMap = map[string]v1.PricingLevel{
	"Standard": v1.PricingLevel1,
	"Premium":  v1.PricingLevel2,
}
