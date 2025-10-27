// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package packet_filter

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
)

type packetFilterExpressionModel struct {
	Protocol        types.String `tfsdk:"protocol"`
	SourceNetwork   types.String `tfsdk:"source_network"`
	SourcePort      types.String `tfsdk:"source_port"`
	DestinationPort types.String `tfsdk:"destination_port"`
	Allow           types.Bool   `tfsdk:"allow"`
	Description     types.String `tfsdk:"description"`
}

type packetFilterBaseModel struct {
	ID          types.String                   `tfsdk:"id"`
	Name        types.String                   `tfsdk:"name"`
	Description types.String                   `tfsdk:"description"`
	Zone        types.String                   `tfsdk:"zone"`
	Expression  []*packetFilterExpressionModel `tfsdk:"expression"`
}

func (model *packetFilterBaseModel) updateState(pf *iaas.PacketFilter, zone string) {
	model.ID = types.StringValue(pf.ID.String())
	model.Name = types.StringValue(pf.Name)
	model.Description = types.StringValue(pf.Description)
	model.Zone = types.StringValue(zone)
	model.Expression = flattenPacketFilterExpressions(pf)
}

func flattenPacketFilterExpressions(pf *iaas.PacketFilter) []*packetFilterExpressionModel {
	var result []*packetFilterExpressionModel
	for _, e := range pf.Expression {
		result = append(result, flattenPacketFilterExpression(e))
	}
	return result
}

func flattenPacketFilterExpression(exp *iaas.PacketFilterExpression) *packetFilterExpressionModel {
	expression := &packetFilterExpressionModel{
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
