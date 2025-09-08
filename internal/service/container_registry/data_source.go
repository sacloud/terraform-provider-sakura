// Copyright 2016-2025 terraform-provider-sakura authors
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

package container_registry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type containerRegistryDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &containerRegistryDataSource{}
	_ datasource.DataSourceWithConfigure = &containerRegistryDataSource{}
)

func NewContainerRegistryDataSource() datasource.DataSource {
	return &containerRegistryDataSource{}
}

func (d *containerRegistryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_registry"
}

func (d *containerRegistryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type containerRegistryDataSourceModel struct {
	containerRegistryBaseModel
}

func (d *containerRegistryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("ContainerRegistry"),
			"name":        common.SchemaDataSourceName("ContainerRegistry"),
			"description": common.SchemaDataSourceDescription("ContainerRegistry"),
			"tags":        common.SchemaDataSourceTags("ContainerRegistry"),
			"icon_id":     common.SchemaDataSourceIconID("ContainerRegistry"),
			"access_level": schema.StringAttribute{
				Computed:    true,
				Description: "The level of access that allow to users. This will be one of [read, write, admin]",
			},
			"virtual_domain": schema.StringAttribute{
				Computed:    true,
				Description: "The alias for accessing the container registry",
			},
			"subdomain_label": schema.StringAttribute{
				Computed:    true,
				Description: "The label at the lowest of the FQDN used when be accessed from users",
			},
			"fqdn": schema.StringAttribute{
				Computed:    true,
				Description: "The FQDN for accessing the container registry. FQDN is built from `subdomain_label` + `.sakuracr.jp`",
			},
			"user": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The user name used to authenticate remote access",
						},
						"password": schema.StringAttribute{
							Computed: true,
							//Sensitive:   true, // password is not sensitive because this attribute is always empty in data source
							Description: "The password used to authenticate remote access",
						},
						"permission": schema.StringAttribute{
							Computed:    true,
							Description: "The level of access that allow to the user. This will be one of [read, write, admin]",
						},
					},
				},
			},
		},
	}
}

func (d *containerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data containerRegistryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewContainerRegistryOp(d.client)
	res, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", "could not find SakuraCloud ContainerRegistry")
		return
	}
	if res.Count == 0 || len(res.ContainerRegistries) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	cr := res.ContainerRegistries[0]
	data.updateState(ctx, d.client, cr, true, &resp.Diagnostics)
	data.IconID = types.StringValue(cr.IconID.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
