// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
)

type apigwDomainBaseModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	CertificateId   types.String `tfsdk:"certificate_id"`
	CertificateName types.String `tfsdk:"certificate_name"`
}

func (m *apigwDomainBaseModel) updateState(domain *v1.Domain) {
	m.ID = types.StringValue(domain.ID.Value.String())
	m.Name = types.StringValue(domain.DomainName)
	m.CreatedAt = types.StringValue(domain.CreatedAt.Value.String())
	m.UpdatedAt = types.StringValue(domain.UpdatedAt.Value.String())
	m.CertificateName = types.StringValue(domain.CertificateName.Value)
	if domain.CertificateId.IsSet() {
		m.CertificateId = types.StringValue(domain.CertificateId.Value.String())
	}
}
