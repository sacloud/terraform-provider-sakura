// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*srnType)(nil)
)

var (
	SRNType = srnType{}
)

type srnType struct {
	basetypes.StringType
}

func (t srnType) Equal(o attr.Type) bool {
	other, ok := o.(srnType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t srnType) ValueType(context.Context) attr.Value {
	return SRN{}
}

func (t srnType) String() string {
	return "SRNType"
}

func (t srnType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return SRNNull(), diags
	}
	if in.IsUnknown() {
		return SRNUnknown(), diags
	}

	return SRNValue(in.ValueString()), diags
}

func (t srnType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
