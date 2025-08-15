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
