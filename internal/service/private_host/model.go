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
}
