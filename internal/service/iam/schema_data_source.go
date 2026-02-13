// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func schemaDataSourceIAMCode(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: fmt.Sprintf("The code of %s", name),
	}
}
