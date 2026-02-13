// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type authBaseModel struct {
	PasswordPolicy types.Object `tfsdk:"password_policy"`
	Conditions     types.Object `tfsdk:"conditions"`
}

type authPasswordModel struct {
	MinLength        types.Int32 `tfsdk:"min_length"`
	RequireUppercase types.Bool  `tfsdk:"require_uppercase"`
	RequireLowercase types.Bool  `tfsdk:"require_lowercase"`
	RequireSymbols   types.Bool  `tfsdk:"require_symbols"`
}

func (m authPasswordModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"min_length":        types.Int32Type,
		"require_uppercase": types.BoolType,
		"require_lowercase": types.BoolType,
		"require_symbols":   types.BoolType,
	}
}

type authConditionsModel struct {
	//IPRestriction        *authIPRestrictionModel       `tfsdk:"ip_restriction"`
	//DatetimeRestriction  *authDatetimeRestrictionModel `tfsdk:"datetime_restriction"`
	IPRestriction        types.Object `tfsdk:"ip_restriction"`
	DatetimeRestriction  types.Object `tfsdk:"datetime_restriction"`
	RequireTwoFactorAuth types.Bool   `tfsdk:"require_two_factor_auth"`
}

func (m authConditionsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_restriction":          types.ObjectType{AttrTypes: authIPRestrictionModel{}.AttributeTypes()},
		"datetime_restriction":    types.ObjectType{AttrTypes: authDatetimeRestrictionModel{}.AttributeTypes()},
		"require_two_factor_auth": types.BoolType,
	}
}

type authIPRestrictionModel struct {
	Mode          types.String `tfsdk:"mode"`
	SourceNetwork types.List   `tfsdk:"source_network"`
}

func (m authIPRestrictionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":           types.StringType,
		"source_network": types.ListType{ElemType: types.StringType},
	}
}

type authDatetimeRestrictionModel struct {
	After  types.String `tfsdk:"after"`
	Before types.String `tfsdk:"before"`
}

func (m authDatetimeRestrictionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"after":  types.StringType,
		"before": types.StringType,
	}
}

func (model *authBaseModel) updateState(pp *v1.PasswordPolicy, ac *v1.AuthConditions) {
	if pp != nil {
		m := &authPasswordModel{
			MinLength:        types.Int32Value(int32(pp.MinLength)),
			RequireUppercase: types.BoolValue(pp.RequireUppercase),
			RequireLowercase: types.BoolValue(pp.RequireLowercase),
			RequireSymbols:   types.BoolValue(pp.RequireSymbols),
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if !diags.HasError() {
			model.PasswordPolicy = value
		}
	} else {
		model.PasswordPolicy = types.ObjectNull(authPasswordModel{}.AttributeTypes())
	}

	if ac != nil {
		m := &authConditionsModel{
			RequireTwoFactorAuth: types.BoolValue(ac.RequireTwoFactorAuth.Enabled),
		}
		if ac.IPRestriction.OneOf.Type == v1.AuthConditionsIPRestrictionSum0AuthConditionsIPRestrictionSum {
			ipr := authIPRestrictionModel{
				Mode:          types.StringValue(string(ac.IPRestriction.OneOf.AuthConditionsIPRestrictionSum0.Mode.Value)),
				SourceNetwork: types.ListNull(types.StringType),
			}
			value, diags := types.ObjectValueFrom(context.Background(), ipr.AttributeTypes(), ipr)
			if !diags.HasError() {
				m.IPRestriction = value
			}
		} else {
			ipr := authIPRestrictionModel{
				Mode:          types.StringValue(string(ac.IPRestriction.OneOf.AuthConditionsIPRestrictionSum1.Mode.Value)),
				SourceNetwork: common.StringsToTlist(ac.IPRestriction.OneOf.AuthConditionsIPRestrictionSum1.SourceNetwork),
			}
			value, diags := types.ObjectValueFrom(context.Background(), ipr.AttributeTypes(), ipr)
			if !diags.HasError() {
				m.IPRestriction = value
			}
		}
		dtr := authDatetimeRestrictionModel{}
		if ac.DatetimeRestriction.After.IsNull() {
			dtr.After = types.StringNull()
		} else {
			dtr.After = types.StringValue(ac.DatetimeRestriction.After.Value.Format(time.RFC3339Nano))
		}
		if ac.DatetimeRestriction.Before.IsNull() {
			dtr.Before = types.StringNull()
		} else {
			dtr.Before = types.StringValue(ac.DatetimeRestriction.Before.Value.Format(time.RFC3339Nano))
		}
		v, diags := types.ObjectValueFrom(context.Background(), dtr.AttributeTypes(), dtr)
		if !diags.HasError() {
			m.DatetimeRestriction = v
		}

		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if !diags.HasError() {
			model.Conditions = value
		}
	} else {
		model.Conditions = types.ObjectNull(authConditionsModel{}.AttributeTypes())
	}
}

func flattenDatetimeRestrictionField(dt v1.NilDateTime) types.String {
	if dt.IsNull() {
		return types.StringNull()
	}
	return types.StringValue(dt.Value.String())
}
