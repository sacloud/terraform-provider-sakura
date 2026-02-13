// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

func SchemaDataSourceId(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The ID of the %s.", name),
	}
}

func SchemaDataSourceName(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The name of the %s.", name),
	}
}

func SchemaDataSourceDescription(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The description of the %s.", name),
	}
}

func SchemaDataSourceIconID(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The icon id attached to the %s", name),
	}
}

func SchemaDataSourceTags(name string) schema.Attribute {
	return schema.SetAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The tags of the %s.", name),
	}
}

func SchemaDataSourceComputedTags(name string) schema.Attribute {
	return schema.SetAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: desc.Sprintf("The tags of the %s.", name),
	}
}

func SchemaDataSourceZone(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The name of zone that the %s is in (e.g. `is1a`, `tk1a`)", name),
	}
}

func SchemaDataSourceSize(name string) schema.Attribute {
	return schema.Int64Attribute{
		Computed:    true,
		Description: desc.Sprintf("The size of %s in GiB", name),
	}
}

func SchemaDataSourcePlan(name string, plans []string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.DataSourcePlan(name, plans),
	}
}

func SchemaDataSourceServerID(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The id of the server connected to the %s", name),
	}
}

func SchemaDataSourceVSwitchID(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The id of the vSwitch connected from the %s", name),
	}
}

func SchemaDataSourceIPAddress(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The IP address assigned to the %s", name),
	}
}

func SchemaDataSourceIPAddresses(name string) schema.Attribute {
	return schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: desc.Sprintf("The IP addresses assigned to the %s", name),
	}
}

func SchemaDataSourceNetMask(name string) schema.Attribute {
	return schema.Int32Attribute{
		Computed:    true,
		Description: desc.Sprintf("The bit length of the subnet assigned to the %s", name),
	}
}

func SchemaDataSourceGateway(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The IP address of the gateway used by %s", name),
	}
}

func SchemaDataSourceClass(name string, classes []string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The class of the %s. This will be one of [%s]", name, classes),
	}
}

func SchemaDataSourceMonitoringSuite(name string) schema.Attribute {
	return schema.SingleNestedAttribute{
		Computed:    true,
		Description: desc.Sprintf("The monitoring suite settings of the %s.", name),
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Enable sending signals to Monitoring Suite",
			},
		},
	}
}

func SchemaDataSourceEncryptionDisk(name string) schema.Attribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"encryption_algorithm": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The disk encryption algorithm. This must be one of [%s]", iaastypes.DiskEncryptionAlgorithmStrings),
			},
			"kms_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the KMS key for encryption",
			},
		},
	}
}

func SchemaDataSourceCreatedAt(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The creation timestamp of the %s", name),
	}
}

func SchemaDataSourceUpdatedAt(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The last update timestamp of the %s", name),
	}
}
