// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"strings"
)

func StringInSlice(validList []string, k string, v string, ignoreCase bool) error {
	for _, valid := range validList {
		if v == valid {
			return nil
		}
		if ignoreCase && strings.EqualFold(v, valid) {
			return nil
		}
	}

	return fmt.Errorf("invalid %s value: %s. valid values are %s", k, v, validList)
}
