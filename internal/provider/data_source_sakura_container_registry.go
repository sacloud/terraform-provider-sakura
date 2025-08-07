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
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iaas-api-go"
)

type containerRegistryDataSourceModel struct {
	containerRegistryResourceModel
	Filter *filterBlockModel `tfsdk:"filter"`
}

type containerRegistryDataSource struct {
	client *APIClient
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
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

func (d *containerRegistryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("ContainerRegistry"),
			"name":        schemaDataSourceName("ContainerRegistry"),
			"description": schemaDataSourceDescription("ContainerRegistry"),
			"tags":        schemaDataSourceTags("ContainerRegistry"),
			"icon_id":     schemaDataSourceIconID("ContainerRegistry"),
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
							//Sensitive:   true,
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
		Blocks: filterSchema(&filterSchemaOption{}),
	}
}

func (d *containerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state containerRegistryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewContainerRegistryOp(d.client)

	findCondition := &iaas.FindCondition{}
	if state.Filter != nil {
		findCondition.Filter = expandSearchFilter(state.Filter)
	}

	res, err := searcher.Find(ctx, findCondition)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", "could not find SakuraCloud ContainerRegistry")
		return
	}
	if res.Count == 0 || len(res.ContainerRegistries) == 0 {
		filterNoResultErr(&resp.Diagnostics)
		return
	}

	target := res.ContainerRegistries[0]
	state.updateState(ctx, d.client, target, true, &resp.Diagnostics)
	/*
		users := getContainerRegistryUsers(ctx, d.client, target)
		if users == nil {
			resp.Diagnostics.AddError("Read Error", "could not get users of SakuraCloud ContainerRegistry")
			return
		}


		state.ID = types.StringValue(target.ID.String())
		state.Name = types.StringValue(target.Name)
		state.AccessLevel = types.StringValue(string(target.AccessLevel))
		state.VirtualDomain = types.StringValue(target.VirtualDomain)
		state.SubDomainLabel = types.StringValue(target.SubDomainLabel)
		state.FQDN = types.StringValue(target.FQDN)
		state.IconID = types.StringValue(target.IconID.String())
		state.Description = types.StringValue(target.Description)
		state.Tags = stringsToTset(ctx, target.Tags)
		state.User = flattenContainerRegistryUsers(state.User, users, false)
	*/

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
