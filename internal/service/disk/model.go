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

package disk

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

type diskBaseModel struct {
	common.SakuraBaseModel
	IconID              types.String `tfsdk:"icon_id"`
	Zone                types.String `tfsdk:"zone"`
	Plan                types.String `tfsdk:"plan"`
	Size                types.Int64  `tfsdk:"size"`
	Connector           types.String `tfsdk:"connector"`
	EncryptionAlgorithm types.String `tfsdk:"encryption_algorithm"`
	SourceArchiveID     types.String `tfsdk:"source_archive_id"`
	SourceDiskID        types.String `tfsdk:"source_disk_id"`
	ServerID            types.String `tfsdk:"server_id"`
}

func (model *diskBaseModel) updateState(disk *iaas.Disk, zone string) {
	model.UpdateBaseState(disk.ID.String(), disk.Name, disk.Description, disk.Tags)
	model.Zone = types.StringValue(zone)
	model.Plan = types.StringValue(iaastypes.DiskPlanNameMap[disk.DiskPlanID])
	model.Size = types.Int64Value(int64(disk.GetSizeGB()))
	model.Connector = types.StringValue(disk.Connection.String())
	model.EncryptionAlgorithm = types.StringValue(string(disk.EncryptionAlgorithm.String()))
	model.SourceArchiveID = types.StringValue(disk.SourceArchiveID.String())
	model.SourceDiskID = types.StringValue(disk.SourceDiskID.String())
	model.ServerID = types.StringValue(disk.ServerID.String())
}
