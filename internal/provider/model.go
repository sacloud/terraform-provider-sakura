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

package sakura

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/simplemq-api-go"
	"github.com/sacloud/simplemq-api-go/apis/v1/queue"
)

type sakuraBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
}

func (m *sakuraBaseModel) updateBaseState(id string, name string, desc string, tags []string) {
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(name)
	m.Description = types.StringValue(desc)
	m.Tags = stringsToTset(tags)
}

type sakuraDiskBaseModel struct {
	sakuraBaseModel
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

func (m *sakuraDiskBaseModel) updateState(disk *iaas.Disk, zone string) {
	m.updateBaseState(disk.ID.String(), disk.Name, disk.Description, disk.Tags)
	m.IconID = types.StringValue(disk.IconID.String())
	m.Zone = types.StringValue(zone)
	m.Plan = types.StringValue(iaastypes.DiskPlanNameMap[disk.DiskPlanID])
	m.Size = types.Int64Value(int64(disk.GetSizeGB()))
	m.Connector = types.StringValue(disk.Connection.String())
	m.EncryptionAlgorithm = types.StringValue(string(disk.EncryptionAlgorithm.String()))
	m.SourceArchiveID = types.StringValue(disk.SourceArchiveID.String())
	m.SourceDiskID = types.StringValue(disk.SourceDiskID.String())
	m.ServerID = types.StringValue(disk.ServerID.String())
}

type sakuraNFSNetworkInterfaceModel struct {
	SwitchID  types.String `tfsdk:"switch_id"`
	IPAddress types.String `tfsdk:"ip_address"`
	Netmask   types.Int32  `tfsdk:"netmask"`
	Gateway   types.String `tfsdk:"gateway"`
}

type sakuraNFSBaseModel struct {
	sakuraBaseModel
	Zone             types.String                      `tfsdk:"zone"`
	IconID           types.String                      `tfsdk:"icon_id"`
	Plan             types.String                      `tfsdk:"plan"`
	Size             types.Int64                       `tfsdk:"size"`
	NetworkInterface []*sakuraNFSNetworkInterfaceModel `tfsdk:"network_interface"`
}

func (model *sakuraNFSBaseModel) updateState(ctx context.Context, client *APIClient, nfs *iaas.NFS, zone string) (bool, error) {
	if nfs.Availability.IsFailed() {
		return true, fmt.Errorf("got unexpected state: NFS[%d].Availability is failed", nfs.ID)
	}

	model.updateBaseState(nfs.ID.String(), nfs.Name, nfs.Description, nfs.Tags)
	model.Zone = types.StringValue(zone)
	model.IconID = types.StringValue(nfs.IconID.String())

	plan, size, err := flattenNFSDiskPlan(ctx, client, nfs.PlanID)
	if err != nil {
		return false, err
	}
	model.Plan = types.StringValue(plan)
	model.Size = types.Int64Value(int64(size))

	var nis []*sakuraNFSNetworkInterfaceModel
	nis = append(nis, &sakuraNFSNetworkInterfaceModel{
		SwitchID:  types.StringValue(nfs.SwitchID.String()),
		IPAddress: types.StringValue(nfs.IPAddresses[0]),
		Netmask:   types.Int32Value(int32(nfs.NetworkMaskLen)),
		Gateway:   types.StringValue(nfs.DefaultRoute),
	})
	model.NetworkInterface = nis

	return false, nil
}

func flattenNFSDiskPlan(ctx context.Context, client *APIClient, planID iaastypes.ID) (string, int, error) {
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

type sakuraNoteBaseModel struct {
	sakuraBaseModel
	IconID  types.String `tfsdk:"icon_id"`
	Class   types.String `tfsdk:"class"`
	Content types.String `tfsdk:"content"`
}

func (model *sakuraNoteBaseModel) updateState(note *iaas.Note) {
	model.updateBaseState(note.ID.String(), note.Name, note.Description, note.Tags)
	model.IconID = types.StringValue(note.IconID.String())
	model.Class = types.StringValue(note.Class)
	model.Content = types.StringValue(note.Content)
}

type sakuraPacketFilterExpressionModel struct {
	Protocol        types.String `tfsdk:"protocol"`
	SourceNetwork   types.String `tfsdk:"source_network"`
	SourcePort      types.String `tfsdk:"source_port"`
	DestinationPort types.String `tfsdk:"destination_port"`
	Allow           types.Bool   `tfsdk:"allow"`
	Description     types.String `tfsdk:"description"`
}

type sakuraPacketFilterBaseModel struct {
	ID          types.String                         `tfsdk:"id"`
	Name        types.String                         `tfsdk:"name"`
	Description types.String                         `tfsdk:"description"`
	Zone        types.String                         `tfsdk:"zone"`
	Expression  []*sakuraPacketFilterExpressionModel `tfsdk:"expression"`
}

func (model *sakuraPacketFilterBaseModel) updateState(pf *iaas.PacketFilter, zone string) {
	model.ID = types.StringValue(pf.ID.String())
	model.Name = types.StringValue(pf.Name)
	model.Description = types.StringValue(pf.Description)
	model.Zone = types.StringValue(zone)
	model.Expression = flattenPacketFilterExpressions(pf)
}

func flattenPacketFilterExpressions(pf *iaas.PacketFilter) []*sakuraPacketFilterExpressionModel {
	var result []*sakuraPacketFilterExpressionModel
	for _, e := range pf.Expression {
		result = append(result, flattenPacketFilterExpression(e))
	}
	return result
}

func flattenPacketFilterExpression(exp *iaas.PacketFilterExpression) *sakuraPacketFilterExpressionModel {
	expression := &sakuraPacketFilterExpressionModel{
		Protocol:    types.StringValue(string(exp.Protocol)),
		Allow:       types.BoolValue(exp.Action.IsAllow()),
		Description: types.StringValue(exp.Description),
	}
	switch exp.Protocol {
	case iaastypes.Protocols.TCP, iaastypes.Protocols.UDP:
		expression.SourceNetwork = types.StringValue(string(exp.SourceNetwork))
		expression.SourcePort = types.StringValue(string(exp.SourcePort))
		expression.DestinationPort = types.StringValue(string(exp.DestinationPort))
	case iaastypes.Protocols.ICMP, iaastypes.Protocols.Fragment, iaastypes.Protocols.IP:
		expression.SourceNetwork = types.StringValue(string(exp.SourceNetwork))
		expression.SourcePort = types.StringValue("") // Optional/Computedのため空文字列にする。nullを渡すとエラーになる
		expression.DestinationPort = types.StringValue("")
	}

	return expression
}

type sakuraPrivateHostBaseModel struct {
	sakuraBaseModel
	Zone           types.String `tfsdk:"zone"`
	IconID         types.String `tfsdk:"icon_id"`
	Class          types.String `tfsdk:"class"`
	Hostname       types.String `tfsdk:"hostname"`
	AssignedCore   types.Int32  `tfsdk:"assigned_core"`
	AssignedMemory types.Int32  `tfsdk:"assigned_memory"`
}

func (model *sakuraPrivateHostBaseModel) updateState(ph *iaas.PrivateHost, zone string) {
	model.updateBaseState(ph.ID.String(), ph.Name, ph.Description, ph.Tags)
	model.Zone = types.StringValue(zone)
	model.IconID = types.StringValue(ph.IconID.String())
	model.Class = types.StringValue(ph.PlanClass)
	model.Hostname = types.StringValue(ph.GetHostName())
	model.AssignedCore = types.Int32Value(int32(ph.GetAssignedCPU()))
	model.AssignedMemory = types.Int32Value(int32(ph.GetAssignedMemoryGB()))
}

type sakuraSwitchBaseModel struct {
	sakuraBaseModel
	IconID    types.String `tfsdk:"icon_id"`
	BridgeID  types.String `tfsdk:"bridge_id"`
	ServerIDs types.Set    `tfsdk:"server_ids"`
	Zone      types.String `tfsdk:"zone"`
}

func (model *sakuraSwitchBaseModel) updateState(ctx context.Context, client *APIClient, sw *iaas.Switch, zone string) error {
	model.updateBaseState(sw.ID.String(), sw.Name, sw.Description, sw.Tags)

	model.IconID = types.StringValue(sw.IconID.String())
	model.BridgeID = types.StringValue(sw.BridgeID.String())
	model.Zone = types.StringValue(zone)

	var serverIDs []string
	if sw.ServerCount > 0 {
		swOp := iaas.NewSwitchOp(client)
		searched, err := swOp.GetServers(ctx, zone, sw.ID)
		if err != nil {
			return fmt.Errorf("could not find SakuraCloud Servers: switch[%s]", err)
		}
		for _, s := range searched.Servers {
			serverIDs = append(serverIDs, s.ID.String())
		}
	}
	model.ServerIDs = stringsToTset(serverIDs)

	return nil
}

type sakuraSimpleMQBaseModel struct {
	sakuraBaseModel
	IconID                   types.String `tfsdk:"icon_id"`
	VisibilityTimeoutSeconds types.Int64  `tfsdk:"visibility_timeout_seconds"`
	ExpireSeconds            types.Int64  `tfsdk:"expire_seconds"`
}

func (model *sakuraSimpleMQBaseModel) updateState(data *queue.CommonServiceItem) {
	model.ID = types.StringValue(simplemq.GetQueueID(data))
	model.Name = types.StringValue(simplemq.GetQueueName(data))
	model.VisibilityTimeoutSeconds = types.Int64Value(int64(data.Settings.VisibilityTimeoutSeconds))
	model.ExpireSeconds = types.Int64Value(int64(data.Settings.ExpireSeconds))
	if v, ok := data.Description.Value.GetString(); ok {
		model.Description = types.StringValue(v)
	}
	if iconID, ok := data.Icon.Value.Icon1.ID.Get(); ok {
		id, ok := iconID.GetString()
		if !ok {
			id = strconv.Itoa(iconID.Int)
		}
		model.IconID = types.StringValue(id)
	} else {
		model.IconID = types.StringValue("")
	}
	model.Tags = stringsToTset(data.Tags)
}

type sakuraSSHKeyBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	PublicKey   types.String `tfsdk:"public_key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

func (model *sakuraSSHKeyBaseModel) updateState(key *iaas.SSHKey) {
	model.ID = types.StringValue(key.ID.String())
	model.Name = types.StringValue(key.Name)
	model.Description = types.StringValue(key.Description)
	model.PublicKey = types.StringValue(key.PublicKey)
	model.Fingerprint = types.StringValue(key.Fingerprint)
}
