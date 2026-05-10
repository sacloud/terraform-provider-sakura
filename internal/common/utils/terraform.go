// Copyright 2016-2025 The sacloud/terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func IsKnown(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

// SDK v2のHasChangeの代替。複雑なGoの値の比較に使う。Terraformの値の比較にはEqualを使う
func HasChange(x, y any, opts ...cmp.Option) bool {
	return !cmp.Equal(x, y, opts...)
}
