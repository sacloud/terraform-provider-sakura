// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package bridge

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
)

type bridgeBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Zone        types.String `tfsdk:"zone"`
}

func (model *bridgeBaseModel) updateState(bridge *iaas.Bridge, zone string) {
	model.ID = types.StringValue(bridge.ID.String())
	model.Name = types.StringValue(bridge.Name)
	model.Description = types.StringValue(bridge.Description)
	model.Zone = types.StringValue(zone)
}
