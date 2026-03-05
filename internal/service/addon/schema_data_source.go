// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

func schemaDataSourceAddonID(name string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: desc.Sprintf("The ID of the %s", name),
	}
}

func schemaDataSourceAddonLocation(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The location of the %s", name),
	}
}

func schemaDataSourceAddonSKU(name string) schema.Attribute {
	return schema.Int32Attribute{
		Computed:    true,
		Description: desc.Sprintf("The SKU of the %s.", name),
	}
}

func schemaDataSourceAddonPricingLevel(name string) schema.Attribute {
	return schema.Int32Attribute{
		Computed:    true,
		Description: desc.Sprintf("The pricing level of the %s.", name),
	}
}

func schemaDataSourceAddonDeploymentName(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The deployment name of the %s", name),
	}
}

func schemaDataSourceAddonURL(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The URL of the %s", name),
	}
}
