// Copyright 2016-2025 The sacloud/terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func IsKnown(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}
