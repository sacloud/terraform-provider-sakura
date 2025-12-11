// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package private_host

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type privateHostBaseModel struct {
	common.SakuraBaseModel
	Zone           types.String `tfsdk:"zone"`
	IconID         types.String `tfsdk:"icon_id"`
	Class          types.String `tfsdk:"class"`
	Hostname       types.String `tfsdk:"hostname"`
	AssignedCore   types.Int32  `tfsdk:"assigned_core"`
	AssignedMemory types.Int32  `tfsdk:"assigned_memory"`
}

func (model *privateHostBaseModel) updateState(ph *iaas.PrivateHost, zone string) {
	model.UpdateBaseState(ph.ID.String(), ph.Name, ph.Description, ph.Tags)
	model.Zone = types.StringValue(zone)
	model.Class = types.StringValue(ph.PlanClass)
	model.Hostname = types.StringValue(ph.GetHostName())
	model.AssignedCore = types.Int32Value(int32(ph.GetAssignedCPU()))
	model.AssignedMemory = types.Int32Value(int32(ph.GetAssignedMemoryGB()))
	if ph.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(ph.IconID.String())
	}
}
