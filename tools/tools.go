// Copyright 2016-2025 The sacloud/terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

//go:build generate

package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)

// Generate documentation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir ..  --provider-name=sakura
