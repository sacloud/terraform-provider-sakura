// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

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

// This `~[16]byte` is ugly!
// But we have no other possible way to do this.
func intoUUID[T ~[16]byte](v types.String) (t T, err error) {
	u, err := uuid.Parse(v.ValueString())

	if err == nil {
		t = T(u)
	}

	return
}

// This `~[16]byte` is ugly!
// But we have no other possible way to do this.
func uuid2StringValue[T ~[16]byte](t T) types.String {
	return types.StringValue(uuid.UUID(t).String())
}

func intoRFC2822[T ~int | ~int64](t T) types.String {
	return types.StringValue(time.Unix(common.ToInt64(t), 0).Format(time.RFC822))
}
