// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package packet_filter

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

func expandPacketFilterExpressions(exprs []*packetFilterExpressionModel) []*iaas.PacketFilterExpression {
	var expressions []*iaas.PacketFilterExpression
	for _, e := range exprs {
		expressions = append(expressions, expandPacketFilterExpression(e))
	}
	return expressions
}

func expandPacketFilterExpression(expr *packetFilterExpressionModel) *iaas.PacketFilterExpression {
	action := "deny"
	if expr.Allow.ValueBool() {
		action = "allow"
	}

	exp := &iaas.PacketFilterExpression{
		Protocol:      types.Protocol(expr.Protocol.ValueString()),
		SourceNetwork: types.PacketFilterNetwork(expr.SourceNetwork.ValueString()),
		Action:        types.Action(action),
		Description:   expr.Description.ValueString(),
	}
	if !expr.SourcePort.IsNull() && !expr.SourcePort.IsUnknown() {
		exp.SourcePort = types.PacketFilterPort(expr.SourcePort.ValueString())
	}
	if !expr.DestinationPort.IsNull() && !expr.DestinationPort.IsUnknown() {
		exp.DestinationPort = types.PacketFilterPort(expr.DestinationPort.ValueString())
	}

	return exp
}
