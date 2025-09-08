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
