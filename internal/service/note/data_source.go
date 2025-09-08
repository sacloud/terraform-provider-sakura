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

package note

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type noteDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &noteDataSource{}
	_ datasource.DataSourceWithConfigure = &noteDataSource{}
)

func NewNoteDataSource() datasource.DataSource {
	return &noteDataSource{}
}

func (d *noteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_note"
}

func (d *noteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type noteDataSourceModel struct {
	noteBaseModel
}

func (d *noteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Note"),
			"name":        common.SchemaDataSourceName("Note"),
			"description": common.SchemaDataSourceDescription("Note"),
			"icon_id":     common.SchemaDataSourceIconID("Note"),
			"tags":        common.SchemaDataSourceTags("Note"),
			"class":       common.SchemaDataSourceClass("Note", iaastypes.NoteClassStrings),
			"content": schema.StringAttribute{
				Computed:    true,
				Description: "The content of the Note",
			},
		},
	}
}

func (d *noteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data noteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewNoteOp(d.client)
	result, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud Note resource: %s", err))
		return
	}
	if result == nil || result.Count == 0 || len(result.Notes) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	note := result.Notes[0]
	data.updateState(note)
	data.IconID = types.StringValue(note.IconID.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
