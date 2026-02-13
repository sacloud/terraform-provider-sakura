// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type ssoBaseModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	IdpEntityID    types.String `tfsdk:"idp_entity_id"`
	IdpLoginURL    types.String `tfsdk:"idp_login_url"`
	IdpLogoutURL   types.String `tfsdk:"idp_logout_url"`
	IdpCertificate types.String `tfsdk:"idp_certificate"`
	SpEntityID     types.String `tfsdk:"sp_entity_id"`
	SpAcsURL       types.String `tfsdk:"sp_acs_url"`
	Assigned       types.Bool   `tfsdk:"assigned"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func (model *ssoBaseModel) updateState(sso *v1.SSOProfile) {
	model.ID = types.StringValue(strconv.Itoa(sso.ID))
	model.Name = types.StringValue(sso.Name)
	model.Description = types.StringValue(sso.Description)
	model.IdpEntityID = types.StringValue(sso.IdpEntityID)
	model.IdpLoginURL = types.StringValue(sso.IdpLoginURL)
	model.IdpLogoutURL = types.StringValue(sso.IdpLogoutURL)
	// 改行付きの証明書を設定した場合、APIレスポンスでは改行が削除されて返ってくるため、すでに値が設定されている場合はAPIレスポンスの値で上書きしないようにする
	// importでは値が不定なので、APIレスポンスの値を利用する
	if !utils.IsKnown(model.IdpCertificate) {
		model.IdpCertificate = types.StringValue(sso.IdpCertificate)
	}
	model.SpEntityID = types.StringValue(sso.SpEntityID)
	model.SpAcsURL = types.StringValue(sso.SpAcsURL)
	model.Assigned = types.BoolValue(sso.Assigned)
	model.CreatedAt = types.StringValue(sso.CreatedAt)
	model.UpdatedAt = types.StringValue(sso.UpdatedAt)
}
