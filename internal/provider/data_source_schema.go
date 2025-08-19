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

package sakura

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

func schemaDataSourceId(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The ID of the %s.", name),
	}
}

func schemaDataSourceName(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The name of the %s.", name),
	}
}

func schemaDataSourceDescription(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The description of the %s.", name),
	}
}

func schemaDataSourceIconID(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The icon id attached to the %s", name),
	}
}

func schemaDataSourceTags(name string) schema.Attribute {
	return schema.SetAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The tags of the %s.", name),
	}
}

func schemaDataSourceZone(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The name of zone that the %s is in (e.g. `is1a`, `tk1a`)", name),
	}
}

func schemaDataSourceSize(name string) schema.Attribute {
	return schema.Int64Attribute{
		Computed:    true,
		Description: desc.Sprintf("The size of %s in GiB", name),
	}
}

func schemaDataSourcePlan(name string, plans []string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.DataSourcePlan(name, plans),
	}
}

func schemaDataSourceSwitchID(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The id of the switch connected from the %s", name),
	}
}

func schemaDataSourceIPAddress(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The IP address assigned to the %s", name),
	}
}

func schemaDataSourceNetMask(name string) schema.Attribute {
	return schema.Int32Attribute{
		Computed:    true,
		Description: desc.Sprintf("The bit length of the subnet assigned to the %s", name),
	}
}

func schemaDataSourceGateway(name string) schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The IP address of the gateway used by %s", name),
	}
}

func schemaDataSourceClass(name string, classes []string) schema.Attribute {
	return &schema.StringAttribute{
		Computed:    true,
		Description: desc.Sprintf("The class of the %s. This will be one of [%s]", name, classes),
	}
}
