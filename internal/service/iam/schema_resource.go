// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

func schemaResourceIAMCode(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: desc.Sprintf("The code of the %s", name),
	}
}
