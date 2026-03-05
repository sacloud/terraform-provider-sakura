// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"slices"
)

func IsTagsMatched(dataTags []string, destTags []string) bool {
	tagsMatched := true
	for _, tagToFind := range dataTags {
		if slices.Contains(destTags, tagToFind) {
			continue
		}
		tagsMatched = false
		break
	}
	return tagsMatched
}
