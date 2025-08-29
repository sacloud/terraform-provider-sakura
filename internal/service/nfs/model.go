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

package nfs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

type nfsNetworkInterfaceModel struct {
	SwitchID  types.String `tfsdk:"switch_id"`
	IPAddress types.String `tfsdk:"ip_address"`
	Netmask   types.Int32  `tfsdk:"netmask"`
	Gateway   types.String `tfsdk:"gateway"`
}

type nfsBaseModel struct {
	common.SakuraBaseModel
	Zone             types.String                `tfsdk:"zone"`
	IconID           types.String                `tfsdk:"icon_id"`
	Plan             types.String                `tfsdk:"plan"`
	Size             types.Int64                 `tfsdk:"size"`
	NetworkInterface []*nfsNetworkInterfaceModel `tfsdk:"network_interface"`
}

func (model *nfsBaseModel) updateState(ctx context.Context, client *common.APIClient, nfs *iaas.NFS, zone string) (bool, error) {
	if nfs.Availability.IsFailed() {
		return true, fmt.Errorf("got unexpected state: NFS[%d].Availability is failed", nfs.ID)
	}

	model.UpdateBaseState(nfs.ID.String(), nfs.Name, nfs.Description, nfs.Tags)
	model.Zone = types.StringValue(zone)
	model.IconID = types.StringValue(nfs.IconID.String())

	plan, size, err := flattenNFSDiskPlan(ctx, client, nfs.PlanID)
	if err != nil {
		return false, err
	}
	model.Plan = types.StringValue(plan)
	model.Size = types.Int64Value(int64(size))

	var nis []*nfsNetworkInterfaceModel
	nis = append(nis, &nfsNetworkInterfaceModel{
		SwitchID:  types.StringValue(nfs.SwitchID.String()),
		IPAddress: types.StringValue(nfs.IPAddresses[0]),
		Netmask:   types.Int32Value(int32(nfs.NetworkMaskLen)),
		Gateway:   types.StringValue(nfs.DefaultRoute),
	})
	model.NetworkInterface = nis

	return false, nil
}

func flattenNFSDiskPlan(ctx context.Context, client *common.APIClient, planID iaastypes.ID) (string, int, error) {
	planInfo, err := query.GetNFSPlanInfo(ctx, iaas.NewNoteOp(client), planID)
	if err != nil {
		return "", 0, err
	}
	var planName string
	size := int(planInfo.Size)

	switch planInfo.DiskPlanID {
	case iaastypes.NFSPlans.HDD:
		planName = "hdd"
	case iaastypes.NFSPlans.SSD:
		planName = "ssd"
	}

	return planName, size, nil
}
