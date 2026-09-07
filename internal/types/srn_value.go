// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sacloud/sacloud-sdk-go/srn"
)

var (
	_ basetypes.StringValuable       = (*SRN)(nil)
	_ xattr.ValidateableAttribute    = (*SRN)(nil)
	_ function.ValidateableParameter = (*SRN)(nil)
)

type SRN struct {
	basetypes.StringValue
	value srn.SRN
}

func (SRN) Type(context.Context) attr.Type {
	return SRNType
}

func (v SRN) Equal(o attr.Value) bool {
	other, ok := o.(SRN)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v SRN) ValueSRN() srn.SRN {
	return v.value
}

func (v SRN) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if !srn.IsSRN(v.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid SRN Value", "The string value provided is not a valid SRN format.\n"+"Value: "+v.ValueString())
		return
	}
}

func (v SRN) ValidateParameter(ctx context.Context, req function.ValidateParameterRequest, resp *function.ValidateParameterResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if !srn.IsSRN(v.ValueString()) {
		resp.Error = function.NewArgumentFuncError(req.Position, "Invalid SRN Value: "+
			"The string value provided is not a valid SRN format.\n"+"Value: "+v.ValueString())
		return
	}
}

func SRNNull() SRN {
	return SRN{StringValue: basetypes.NewStringNull()}
}

func SRNUnknown() SRN {
	return SRN{StringValue: basetypes.NewStringUnknown()}
}

func SRNValue(value string) SRN {
	// Note: srn.Parse() will return an empty SRN if the value is invalid,
	// but we don't need to handle that here because validation is done in ValidateAttribute and ValidateParameter.
	s, _ := srn.Parse(value)
	return SRN{
		StringValue: basetypes.NewStringValue(value),
		value:       s,
	}
}

func SRNPointerValue(value *string) SRN {
	if value == nil {
		return SRNNull()
	}

	s, _ := srn.Parse(*value)
	return SRN{
		StringValue: basetypes.NewStringPointerValue(value),
		value:       s,
	}
}
