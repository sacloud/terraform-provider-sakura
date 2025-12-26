// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package script

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type scriptDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &scriptDataSource{}
	_ datasource.DataSourceWithConfigure = &scriptDataSource{}
)

func NewScriptDataSource() datasource.DataSource {
	return &scriptDataSource{}
}

func (d *scriptDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_script"
}

func (d *scriptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type scriptDataSourceModel struct {
	scriptBaseModel
}

func (d *scriptDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Script"),
			"name":        common.SchemaDataSourceName("Script"),
			"description": common.SchemaDataSourceDescription("Script"),
			"icon_id":     common.SchemaDataSourceIconID("Script"),
			"tags":        common.SchemaDataSourceTags("Script"),
			"class":       common.SchemaDataSourceClass("Script", iaastypes.NoteClassStrings),
			"content": schema.StringAttribute{
				Computed:    true,
				Description: "The content of the Script",
			},
		},
		MarkdownDescription: "Get information about an existing Script(note in v2).",
	}
}

func (d *scriptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data scriptDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewNoteOp(d.client)
	result, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find SakuraCloud Script resource: %s", err))
		return
	}
	if result == nil || result.Count == 0 || len(result.Notes) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	script := result.Notes[0]
	data.updateState(script)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
