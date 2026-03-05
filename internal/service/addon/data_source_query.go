// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/addon-api-go"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
)

type queryDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &queryDataSource{}
	_ datasource.DataSourceWithConfigure = &queryDataSource{}
)

func NewQueryDataSource() datasource.DataSource {
	return &queryDataSource{}
}

func (d *queryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_query"
}

func (d *queryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type queryDataSourceModel struct {
	queryBaseModel
}

func (d *queryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              schemaDataSourceAddonID("Addon Query"),
			"location":        schemaDataSourceAddonLocation("Addon Query"),
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon Query"),
			"url":             schemaDataSourceAddonURL("Addon Query"),
		},
		MarkdownDescription: "Get information about an existing Addon Query.",
	}
}

func (d *queryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data queryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewQueryOp(d.client)
	result := getAddon(ctx, "Query", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeQueryResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon Query[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
