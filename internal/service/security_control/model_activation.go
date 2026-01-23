// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
)

type activationBaseModel struct {
	ServicePrincipalID   types.String `tfsdk:"service_principal_id"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	AutomatedActionLimit types.Int32  `tfsdk:"automated_action_limit"`
}

func (model *activationBaseModel) updateState(activation *v1.ActivationOutput) {
	model.ServicePrincipalID = types.StringValue(activation.ServicePrincipalId)
	model.Enabled = types.BoolValue(activation.IsActive)
	model.AutomatedActionLimit = types.Int32Value(int32(activation.AutomatedActionLimit))
}
