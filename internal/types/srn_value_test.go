// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/sacloud/sacloud-sdk-go/srn"
	"github.com/sacloud/terraform-provider-sakura/internal/types"
)

func TestSRNValidateAttribute(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		srnValue      types.SRN
		expectedDiags diag.Diagnostics
	}{
		"empty-struct": {
			srnValue: types.SRN{},
		},
		"null": {
			srnValue: types.SRNNull(),
		},
		"unknown": {
			srnValue: types.SRNUnknown(),
		},
		"valid SRN": {
			srnValue: types.SRNValue("srnv1:is1c:sakura.networking-suite.subnet:112233445566"),
		},
		"invalid SRN - invalid version": {
			srnValue: types.SRNValue("srnv9:is1c:sakura.networking-suite.subnet:112233445566"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "Invalid SRN Value",
					"The string value provided is not a valid SRN format.\n"+
						"Value: srnv9:is1c:sakura.networking-suite.subnet:112233445566"),
			},
		},
		"invalid SRN - invalid location": {
			srnValue: types.SRNValue("srnv1::sakura.networking-suite.subnet:112233445566"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "Invalid SRN Value",
					"The string value provided is not a valid SRN format.\n"+
						"Value: srnv1::sakura.networking-suite.subnet:112233445566"),
			},
		},
		"invalid SRN - invalid resource": {
			srnValue: types.SRNValue("srnv1:is1c::112233445566"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "Invalid SRN Value",
					"The string value provided is not a valid SRN format.\n"+
						"Value: srnv1:is1c::112233445566"),
			},
		},
		"invalid SRN - invalid ID": {
			srnValue: types.SRNValue("srnv1:is1c:sakura.networking-suite.subnet:"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "Invalid SRN Value",
					"The string value provided is not a valid SRN format.\n"+
						"Value: srnv1:is1c:sakura.networking-suite.subnet:"),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := xattr.ValidateAttributeResponse{}
			testCase.srnValue.ValidateAttribute(
				context.Background(),
				xattr.ValidateAttributeRequest{Path: path.Root("test")},
				&resp,
			)

			if diff := cmp.Diff(resp.Diagnostics, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestSRNValidateParameter(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		srnValue        types.SRN
		expectedFuncErr *function.FuncError
	}{
		"empty-struct": {
			srnValue: types.SRN{},
		},
		"null": {
			srnValue: types.SRNNull(),
		},
		"unknown": {
			srnValue: types.SRNUnknown(),
		},
		"valid SRN": {
			srnValue: types.SRNValue("srnv1:is1c:sakura.networking-suite.subnet:112233445566"),
		},
		"invalid SRN - invalid version": {
			srnValue: types.SRNValue("srnv9:is1c:sakura.networking-suite.subnet:112233445566"),
			expectedFuncErr: function.NewArgumentFuncError(0, "Invalid SRN Value: "+
				"The string value provided is not a valid SRN format.\n"+
				"Value: srnv9:is1c:sakura.networking-suite.subnet:112233445566"),
		},
		"invalid SRN - invalid location": {
			srnValue: types.SRNValue("srnv1::sakura.networking-suite.subnet:112233445566"),
			expectedFuncErr: function.NewArgumentFuncError(0, "Invalid SRN Value: "+
				"The string value provided is not a valid SRN format.\n"+
				"Value: srnv1::sakura.networking-suite.subnet:112233445566",
			),
		},
		"invalid SRN - invalid resource": {
			srnValue: types.SRNValue("srnv1:is1c::112233445566"),
			expectedFuncErr: function.NewArgumentFuncError(0, "Invalid SRN Value: "+
				"The string value provided is not a valid SRN format.\n"+
				"Value: srnv1:is1c::112233445566",
			),
		},
		"invalid SRN - invalid ID": {
			srnValue: types.SRNValue("srnv1:is1c:sakura.networking-suite.subnet:"),
			expectedFuncErr: function.NewArgumentFuncError(0, "Invalid SRN Value: "+
				"The string value provided is not a valid SRN format.\n"+
				"Value: srnv1:is1c:sakura.networking-suite.subnet:",
			),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := function.ValidateParameterResponse{}
			testCase.srnValue.ValidateParameter(
				context.Background(),
				function.ValidateParameterRequest{
					Position: 0,
				},
				&resp,
			)

			if diff := cmp.Diff(resp.Error, testCase.expectedFuncErr); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestSRNValueSRN(t *testing.T) {
	t.Parallel()

	s, _ := srn.Parse("srnv1:is1c:sakura.networking-suite.subnet:112233445566")

	testCases := map[string]struct {
		srnValue    types.SRN
		expectedSRN srn.SRN
	}{
		"SRN value is null ": {
			srnValue:    types.SRNNull(),
			expectedSRN: srn.SRN{},
		},
		"SRN value is unknown ": {
			srnValue:    types.SRNUnknown(),
			expectedSRN: srn.SRN{},
		},
		"valid SRN ": {
			srnValue:    types.SRNValue("srnv1:is1c:sakura.networking-suite.subnet:112233445566"),
			expectedSRN: s,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := testCase.srnValue.ValueSRN()
			if v != testCase.expectedSRN {
				t.Errorf("Unexpected difference in SRN, got: %s, expected: %s", v, testCase.expectedSRN)
			}

			// SRNValue.ValueSRN() does not return diagnostics, so we skip such check
		})
	}
}
