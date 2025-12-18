// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type stringBackupTimeValidator struct{}

var _ validator.String = stringBackupTimeValidator{}

func (v stringBackupTimeValidator) Description(_ context.Context) string {
	return "string must be a valid backup time in HH:MM format"
}

func (v stringBackupTimeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringBackupTimeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var timeStrings []string
	minutes := []int{0, 15, 30, 45}

	// create list [00:00 ,,, 23:45]
	for hour := 0; hour <= 23; hour++ {
		for _, minute := range minutes {
			timeStrings = append(timeStrings, fmt.Sprintf("%02d:%02d", hour, minute))
		}
	}

	value := req.ConfigValue.ValueString()
	if err := utils.StringInSlice(timeStrings, "backup time", value, false); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, v.Description(ctx), err.Error())
		return
	}
}

func BackupTimeValidator() stringBackupTimeValidator {
	return stringBackupTimeValidator{}
}
