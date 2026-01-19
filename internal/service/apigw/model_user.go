// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwUserBaseModel struct {
	ID            types.String             `tfsdk:"id"`
	Name          types.String             `tfsdk:"name"`
	Tags          types.Set                `tfsdk:"tags"`
	CreatedAt     types.String             `tfsdk:"created_at"`
	UpdatedAt     types.String             `tfsdk:"updated_at"`
	CustomID      types.String             `tfsdk:"custom_id"`
	IPRestriction *apigwIpRestrictionModel `tfsdk:"ip_restriction"`
	Groups        []apigwGroupModel        `tfsdk:"groups"`
}

type apigwIpRestrictionModel struct {
	Protocols    types.String `tfsdk:"protocols"`
	RestrictedBy types.String `tfsdk:"restricted_by"`
	Ips          types.Set    `tfsdk:"ips"`
}

type apigwGroupModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type apigwUserAuthenticationModel struct {
	BasicAuth *apigwUserAuthenticationBasicAuthModel `tfsdk:"basic_auth"`
	JWT       *apigwUserAuthenticationJWTModel       `tfsdk:"jwt"`
	HMACAuth  *apigwUserAuthenticationHMACAuthModel  `tfsdk:"hmac_auth"`
}

type apigwUserAuthenticationBasicAuthModel struct {
	Username types.String `tfsdk:"username"`
}

type apigwUserAuthenticationJWTModel struct {
	Key       types.String `tfsdk:"key"`
	Algorithm types.String `tfsdk:"algorithm"`
}

type apigwUserAuthenticationHMACAuthModel struct {
	Username types.String `tfsdk:"username"`
}

func (m *apigwUserBaseModel) updateState(user *v1.UserDetail) error {
	m.ID = types.StringValue(user.ID.Value.String())
	m.Name = types.StringValue(string(user.Name))
	m.Tags = common.StringsToTset(user.Tags)
	m.CreatedAt = types.StringValue(user.CreatedAt.Value.String())
	m.UpdatedAt = types.StringValue(user.UpdatedAt.Value.String())
	m.CustomID = types.StringValue(string(user.CustomID.Value))

	if user.IpRestrictionConfig.IsSet() {
		ipr := user.IpRestrictionConfig.Value
		m.IPRestriction = &apigwIpRestrictionModel{
			Protocols:    types.StringValue(string(ipr.Protocols)),
			RestrictedBy: types.StringValue(string(ipr.RestrictedBy)),
			Ips:          common.StringsToTset(ipr.Ips),
		}
	}

	var groups []apigwGroupModel
	for _, g := range user.Groups {
		group := apigwGroupModel{
			ID:   types.StringValue(g.ID.Value.String()),
			Name: types.StringValue(string(g.Name.Value)),
		}
		groups = append(groups, group)
	}
	if len(groups) > 0 {
		m.Groups = groups
	}

	return nil
}

func flattenAPIGWUserAuthentication(auth *v1.UserAuthentication) *apigwUserAuthenticationModel {
	if auth == nil {
		return nil
	}

	authentication := &apigwUserAuthenticationModel{}
	if auth.BasicAuth.IsSet() {
		basic := auth.BasicAuth.Value
		authentication.BasicAuth = &apigwUserAuthenticationBasicAuthModel{
			Username: types.StringValue(string(basic.UserName)),
		}
	}
	if auth.Jwt.IsSet() {
		jwt := auth.Jwt.Value
		authentication.JWT = &apigwUserAuthenticationJWTModel{
			Key:       types.StringValue(string(jwt.Key)),
			Algorithm: types.StringValue(string(jwt.Algorithm)),
		}
	}
	if auth.HmacAuth.IsSet() {
		hmac := auth.HmacAuth.Value
		authentication.HMACAuth = &apigwUserAuthenticationHMACAuthModel{
			Username: types.StringValue(string(hmac.UserName)),
		}
	}

	return authentication
}
