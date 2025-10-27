// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package desc

import (
	"fmt"
	"strings"
)

func QuoteAndJoin(a []string, quote, sep string) string {
	if quote == "" {
		quote = `"`
	}
	var ret []string
	for _, w := range a {
		ret = append(ret, fmt.Sprintf("%s%s%s", quote, w, quote))
	}
	return strings.Join(ret, sep)
}

func QuoteAndJoinInt(a []int, quote, sep string) string {
	if quote == "" {
		quote = `"`
	}
	var ret []string
	for _, w := range a {
		ret = append(ret, fmt.Sprintf("%s%d%s", quote, w, quote))
	}
	return strings.Join(ret, sep)
}

func Sprintf(format string, a ...interface{}) string {
	args := make([]interface{}, len(a))
	for i, a := range a {
		var v interface{}
		switch a := a.(type) {
		case []string:
			v = QuoteAndJoin(a, "`", "/")
		case []int:
			v = QuoteAndJoinInt(a, "`", "/")
		default:
			v = a
		}
		args[i] = v
	}
	return fmt.Sprintf(format, args...)
}

func Range(min, max int) string {
	return fmt.Sprintf("This must be in the range [`%d`-`%d`]", min, max)
}

func Length(min, max int) string {
	return fmt.Sprintf("The length of this value must be in the range [`%d`-`%d`]", min, max)
}

func Conflicts(names ...string) string {
	return Sprintf("This conflicts with [%s]", names)
}

func ResourcePlan(resourceName string, plans interface{}) string {
	return Sprintf(
		"The plan name of the %s. This must be one of [%s]",
		resourceName, plans,
	)
}

func DataSourcePlan(resourceName string, plans interface{}) string {
	return Sprintf(
		"The plan name of the %s. This will be one of [%s]",
		resourceName, plans,
	)
}
