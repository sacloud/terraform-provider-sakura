// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

func listed[T, U any](yield func(*U) ([]T, *U, error)) (ret []T, err error) {
	var cursor *U

	for {
		var items []T

		items, cursor, err = yield(cursor)
		ret = append(ret, items...)

		if err != nil {
			return
		}

		if cursor == nil {
			break
		}
	}
	return
}
