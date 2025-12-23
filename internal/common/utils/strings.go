// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"strconv"
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

func MustAtoI(target string) int {
	v, _ := strconv.Atoi(target)
	return v
}

func MustAtoInt64(target string) int64 {
	v, _ := strconv.ParseInt(target, 10, 64)
	return v
}
