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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
)

type iconDataSourceModel struct {
	ID     types.String      `tfsdk:"id"`
	Name   types.String      `tfsdk:"name"`
	Tags   types.Set         `tfsdk:"tags"`
	URL    types.String      `tfsdk:"url"`
	Filter *filterBlockModel `tfsdk:"filter"`
}

func NewIconDataSource() datasource.DataSource {
	return &iconDataSource{}
}

type iconDataSource struct {
	client *APIClient
}

var (
	_ datasource.DataSource              = &iconDataSource{}
	_ datasource.DataSourceWithConfigure = &iconDataSource{}
)

func (d *iconDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

func (d *iconDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_icon"
}

func (d *iconDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   schemaDataSourceId("Icon"),
			"tags": schemaDataSourceTags("Icon"),
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the Icon.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL for getting the icon's raw data",
			},
		},
		Blocks: filterSchema(&filterSchemaOption{}),
	}
}

func (d *iconDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state iconDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewIconOp(d.client)
	findCondition := &iaas.FindCondition{}
	if state.Filter != nil {
		findCondition.Filter = expandSearchFilter(state.Filter)
	}

	res, err := searcher.Find(ctx, findCondition)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", "could not find SakuraCloud ContainerRegistry")
		return
	}
	if res == nil || res.Count == 0 || len(res.Icons) == 0 {
		filterNoResultErr(&resp.Diagnostics)
		return
	}

	icon := res.Icons[0]
	state.ID = types.StringValue(icon.ID.String())
	state.Name = types.StringValue(icon.Name)
	state.Tags = stringsToTset(ctx, icon.Tags)
	state.URL = types.StringValue(icon.URL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
