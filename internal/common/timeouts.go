// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	Timeout5min   = 5 * time.Minute
	Timeout20min  = 20 * time.Minute
	Timeout60min  = 60 * time.Minute
	Timeout24hour = 24 * time.Hour
)

func SetupTimeoutCreate(ctx context.Context, tov timeouts.Value, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	createTimeout, diags := tov.Create(ctx, defaultTimeout)

	if diags.HasError() {
		tflog.Info(ctx, fmt.Sprintf("Failed to get create timeout. Use default timeout: %s", createTimeout))
	}

	return context.WithTimeout(ctx, createTimeout)
}

func SetupTimeoutUpdate(ctx context.Context, tov timeouts.Value, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	updateTimeout, diags := tov.Update(ctx, defaultTimeout)

	if diags.HasError() {
		tflog.Info(ctx, fmt.Sprintf("Failed to get update timeout. Use default timeout: %s", updateTimeout))
	}

	return context.WithTimeout(ctx, updateTimeout)
}

func SetupTimeoutDelete(ctx context.Context, tov timeouts.Value, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	deleteTimeout, diags := tov.Delete(ctx, defaultTimeout)

	if diags.HasError() {
		tflog.Info(ctx, fmt.Sprintf("Failed to get delete timeout. Use default timeout: %s", deleteTimeout))
	}

	return context.WithTimeout(ctx, deleteTimeout)
}
