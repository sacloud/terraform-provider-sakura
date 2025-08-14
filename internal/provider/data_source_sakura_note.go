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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
)

type noteDataSource struct {
	client *APIClient
}

func NewNoteDataSource() datasource.DataSource {
	return &noteDataSource{}
}

var (
	_ datasource.DataSource              = &noteDataSource{}
	_ datasource.DataSourceWithConfigure = &noteDataSource{}
)

func (d *noteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_note"
}

func (d *noteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type noteDataSourceModel struct {
	sakuraNoteBaseModel
	Filter *filterBlockModel `tfsdk:"filter"`
}

func (d *noteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("Note"),
			"name":        schemaDataSourceName("Note"),
			"description": schemaDataSourceDescription("Note"),
			"icon_id":     schemaDataSourceIconID("Note"),
			"tags":        schemaDataSourceTags("Note"),
			"class":       schemaDataSourceClass("Note", iaastypes.NoteClassStrings),
			"content": schema.StringAttribute{
				Computed:    true,
				Description: "The content of the Note",
			},
		},
		Blocks: filterSchema(&filterSchemaOption{}),
	}
}

func (d *noteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data noteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewNoteOp(d.client)
	findCondition := &iaas.FindCondition{}
	if data.Filter != nil {
		findCondition.Filter = expandSearchFilter(data.Filter)
	}

	result, err := searcher.Find(ctx, findCondition)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud Note resource: %s", err))
		return
	}
	if result == nil || result.Count == 0 || len(result.Notes) == 0 {
		filterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(result.Notes[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
