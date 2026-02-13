// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
)

type idPolicyBaseModel struct {
	Bindings []idPolicyBindingModel `tfsdk:"bindings"`
}

type idPolicyBindingModel struct {
	Role       *idPolicyRoleModel       `tfsdk:"role"`
	Principals []idPolicyPrincipalModel `tfsdk:"principals"`
}

type idPolicyRoleModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

type idPolicyPrincipalModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

func (model *idPolicyBaseModel) updateState(bindings []v1.IdPolicy) {
	bindingModels := make([]idPolicyBindingModel, 0, len(bindings))
	for _, b := range bindings {
		bModel := idPolicyBindingModel{
			Role: &idPolicyRoleModel{
				Type: types.StringValue(string(b.Role.Value.Type.Value)),
				ID:   types.StringValue(b.Role.Value.ID.Value),
			},
		}
		for _, p := range b.Principals {
			bModel.Principals = append(bModel.Principals, idPolicyPrincipalModel{
				Type: types.StringValue(string(p.Type.Value)),
				ID:   types.StringValue(strconv.Itoa(p.ID.Value)),
			})
		}
		bindingModels = append(bindingModels, bModel)
	}
	model.Bindings = bindingModels
}
