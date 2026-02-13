// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
)

type userBaseModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Code                  types.String `tfsdk:"code"`
	Description           types.String `tfsdk:"description"`
	Status                types.String `tfsdk:"status"`
	Otp                   types.Object `tfsdk:"otp"`
	Member                types.Object `tfsdk:"member"`
	Email                 types.String `tfsdk:"email"`
	SecurityKeyRegistered types.Bool   `tfsdk:"security_key_registered"`
	Passwordless          types.Bool   `tfsdk:"passwordless"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

type userMemberModel struct {
	ID   types.String `tfsdk:"id"`
	Code types.String `tfsdk:"code"`
}

func (m userMemberModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"code": types.StringType,
	}
}

type userOtpModel struct {
	Status          types.String `tfsdk:"status"`
	HasRecoveryCode types.Bool   `tfsdk:"has_recovery_code"`
}

func (m userOtpModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":            types.StringType,
		"has_recovery_code": types.BoolType,
	}
}

func (model *userBaseModel) updateState(user *v1.User) {
	model.ID = types.StringValue(strconv.Itoa(user.ID))
	model.Name = types.StringValue(user.Name)
	model.Code = types.StringValue(user.Code)
	model.Description = types.StringValue(user.Description)
	model.Email = types.StringValue(user.Email)
	model.Status = types.StringValue(string(user.Status))
	otp := userOtpModel{
		Status:          types.StringValue(string(user.Otp.Status)),
		HasRecoveryCode: types.BoolValue(user.Otp.HasRecoveryCode),
	}
	otpM, diags := types.ObjectValueFrom(context.Background(), otp.AttributeTypes(), otp)
	if !diags.HasError() {
		model.Otp = otpM
	}
	member := userMemberModel{
		ID:   types.StringValue(strconv.Itoa(user.Member.ID)),
		Code: types.StringValue(user.Member.Code),
	}
	memberM, diags := types.ObjectValueFrom(context.Background(), member.AttributeTypes(), member)
	if !diags.HasError() {
		model.Member = memberM
	}
	model.SecurityKeyRegistered = types.BoolValue(user.IsSecurityKeyRegistered)
	model.Passwordless = types.BoolValue(user.IsPasswordless)
	model.CreatedAt = types.StringValue(user.CreatedAt)
	model.UpdatedAt = types.StringValue(user.UpdatedAt)
}
