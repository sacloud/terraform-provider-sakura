// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
)

type policyBaseModel struct {
	Target   types.String         `tfsdk:"target"`
	TargetID types.String         `tfsdk:"target_id"`
	Bindings []policyBindingModel `tfsdk:"bindings"`
}

type policyBindingModel struct {
	Role       *policyRoleModel       `tfsdk:"role"`
	Principals []policyPrincipalModel `tfsdk:"principals"`
}

type policyRoleModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

type policyPrincipalModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

func (model *policyBaseModel) updateState(target string, bindings []v1.IamPolicy) {
	model.Target = types.StringValue(target)
	bindingModels := make([]policyBindingModel, 0, len(bindings))
	for _, b := range bindings {
		bModel := policyBindingModel{
			Role: &policyRoleModel{
				Type: types.StringValue(string(b.Role.Value.Type.Value)),
				ID:   types.StringValue(b.Role.Value.ID.Value),
			},
		}
		for _, p := range b.Principals {
			bModel.Principals = append(bModel.Principals, policyPrincipalModel{
				Type: types.StringValue(string(p.Type.Value)),
				ID:   types.StringValue(strconv.Itoa(p.ID.Value)),
			})
		}
		bindingModels = append(bindingModels, bModel)
	}
	model.Bindings = bindingModels
}
