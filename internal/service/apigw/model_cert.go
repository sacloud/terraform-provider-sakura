// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
)

type apigwCertBaseModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (m *apigwCertBaseModel) updateState(cert *v1.Certificate) {
	m.ID = types.StringValue(cert.ID.Value.String())
	m.Name = types.StringValue(string(cert.Name.Value))
	m.CreatedAt = types.StringValue(cert.CreatedAt.Value.String())
	m.UpdatedAt = types.StringValue(cert.UpdatedAt.Value.String())
}
