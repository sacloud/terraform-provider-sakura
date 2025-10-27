// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package disk

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
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
