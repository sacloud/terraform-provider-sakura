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
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	registryBuilder "github.com/sacloud/iaas-service-go/containerregistry/builder"
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

type sakuraContainerRegistryBaseModel struct {
	sakuraBaseModel
	AccessLevel    types.String                        `tfsdk:"access_level"`
	VirtualDomain  types.String                        `tfsdk:"virtual_domain"`
	SubDomainLabel types.String                        `tfsdk:"subdomain_label"`
	FQDN           types.String                        `tfsdk:"fqdn"`
	IconID         types.String                        `tfsdk:"icon_id"`
	User           []*sakuraContainerRegistryUserModel `tfsdk:"user"`
}

type sakuraContainerRegistryUserModel struct {
	Name       types.String `tfsdk:"name"`
	Password   types.String `tfsdk:"password"`
	Permission types.String `tfsdk:"permission"`
}

func (model *sakuraContainerRegistryBaseModel) updateState(ctx context.Context, c *APIClient, reg *iaas.ContainerRegistry, includePassword bool, diags *diag.Diagnostics) {
	users := getContainerRegistryUsers(ctx, c, reg)
	if users == nil {
		diags.AddError("Get Users Error", "could not get users for SakuraCloud ContainerRegistry")
		return
	}

	model.updateBaseState(reg.ID.String(), reg.Name, reg.Description, reg.Tags)
	model.AccessLevel = types.StringValue(string(reg.AccessLevel))
	model.VirtualDomain = types.StringValue(reg.VirtualDomain)
	model.SubDomainLabel = types.StringValue(reg.SubDomainLabel)
	model.FQDN = types.StringValue(reg.FQDN)
	model.IconID = types.StringValue(reg.IconID.String())
	model.User = flattenContainerRegistryUsers(model.User, users, includePassword)
}

func getContainerRegistryUsers(ctx context.Context, client *APIClient, user *iaas.ContainerRegistry) []*iaas.ContainerRegistryUser {
	regOp := iaas.NewContainerRegistryOp(client)
	users, err := regOp.ListUsers(ctx, user.ID)
	if err != nil {
		return nil
	}
	return users.Users
}

func expandContainerRegistryUsers(users []*sakuraContainerRegistryUserModel) []*registryBuilder.User {
	if len(users) == 0 {
		return nil
	}

	var results []*registryBuilder.User
	for _, u := range users {
		results = append(results, &registryBuilder.User{
			UserName:   u.Name.ValueString(),
			Password:   u.Password.ValueString(),
			Permission: iaastypes.EContainerRegistryPermission(u.Permission.ValueString()),
		})
	}
	return results
}

func flattenContainerRegistryUsers(conf []*sakuraContainerRegistryUserModel, users []*iaas.ContainerRegistryUser, includePassword bool) []*sakuraContainerRegistryUserModel {
	inputs := expandContainerRegistryUsers(conf)

	var results []*sakuraContainerRegistryUserModel
	for _, user := range users {
		v := &sakuraContainerRegistryUserModel{
			Name:       types.StringValue(user.UserName),
			Permission: types.StringValue(string(user.Permission)),
		}
		if includePassword {
			password := ""
			for _, i := range inputs {
				if i.UserName == user.UserName {
					password = i.Password
					break
				}
			}
			v.Password = types.StringValue(password)
		}
		results = append(results, v)
	}
	return results
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

type sakuraInternetBaseModel struct {
	sakuraBaseModel
	Zone               types.String `tfsdk:"zone"`
	IconID             types.String `tfsdk:"icon_id"`
	Netmask            types.Int32  `tfsdk:"netmask"`
	BandWidth          types.Int32  `tfsdk:"band_width"`
	EnableIPv6         types.Bool   `tfsdk:"enable_ipv6"`
	SwitchID           types.String `tfsdk:"switch_id"`
	ServerIDs          types.Set    `tfsdk:"server_ids"`
	NetworkAddress     types.String `tfsdk:"network_address"`
	Gateway            types.String `tfsdk:"gateway"`
	MinIPAddress       types.String `tfsdk:"min_ip_address"`
	MaxIPAddress       types.String `tfsdk:"max_ip_address"`
	IPAddresses        types.Set    `tfsdk:"ip_addresses"`
	IPv6Prefix         types.String `tfsdk:"ipv6_prefix"`
	IPv6PrefixLen      types.Int32  `tfsdk:"ipv6_prefix_len"`
	IPv6NetworkAddress types.String `tfsdk:"ipv6_network_address"`
	AssignedTags       types.Set    `tfsdk:"assigned_tags"`
}

func (model *sakuraInternetBaseModel) updateState(ctx context.Context, client *APIClient, zone string, data *iaas.Internet) error {
	swOp := iaas.NewSwitchOp(client)
	sw, err := swOp.Read(ctx, zone, data.Switch.ID)
	if err != nil {
		return fmt.Errorf("could not read SakuraCloud Switch[%s]: %s", data.Switch.ID, err)
	}

	var serverIDs []string
	if sw.ServerCount > 0 {
		servers, err := swOp.GetServers(ctx, zone, sw.ID)
		if err != nil {
			return fmt.Errorf("could not find SakuraCloud Servers of Switch[%s]: %s", sw.ID.String(), err)
		}
		for _, s := range servers.Servers {
			serverIDs = append(serverIDs, s.ID.String())
		}
	}

	var enableIPv6 bool
	var ipv6Prefix, ipv6NetworkAddress string
	var ipv6PrefixLen int
	if len(data.Switch.IPv6Nets) > 0 {
		enableIPv6 = true
		ipv6Prefix = data.Switch.IPv6Nets[0].IPv6Prefix
		ipv6PrefixLen = data.Switch.IPv6Nets[0].IPv6PrefixLen
		ipv6NetworkAddress = fmt.Sprintf("%s/%d", ipv6Prefix, ipv6PrefixLen)
	}
	assigned, unassigned := partitionInternetTags(data.Tags)

	model.updateBaseState(data.ID.String(), data.Name, data.Description, unassigned)
	model.IconID = types.StringValue(data.IconID.String())
	model.Netmask = types.Int32Value(int32(data.NetworkMaskLen))
	model.BandWidth = types.Int32Value(int32(data.BandWidthMbps))
	model.SwitchID = types.StringValue(sw.ID.String())
	model.ServerIDs = stringsToTset(serverIDs)
	model.NetworkAddress = types.StringValue(sw.Subnets[0].NetworkAddress)
	model.Gateway = types.StringValue(sw.Subnets[0].DefaultRoute)
	model.MinIPAddress = types.StringValue(sw.Subnets[0].AssignedIPAddressMin)
	model.MaxIPAddress = types.StringValue(sw.Subnets[0].AssignedIPAddressMax)
	model.EnableIPv6 = types.BoolValue(enableIPv6)
	model.IPv6Prefix = types.StringValue(ipv6Prefix)
	model.IPv6PrefixLen = types.Int32Value(int32(ipv6PrefixLen))
	model.IPv6NetworkAddress = types.StringValue(ipv6NetworkAddress)
	model.Zone = types.StringValue(zone)
	model.IPAddresses = stringsToTset(sw.Subnets[0].GetAssignedIPAddresses())
	model.AssignedTags = stringsToTset(assigned)

	return nil
}

func partitionInternetTags(tags []string) (assigned, unassigned []string) {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@previous-id") {
			assigned = append(assigned, tag)
		} else {
			unassigned = append(unassigned, tag)
		}
	}
	return
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
