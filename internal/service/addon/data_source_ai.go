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

type aiDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &aiDataSource{}
	_ datasource.DataSourceWithConfigure = &aiDataSource{}
)

func NewAIDataSource() datasource.DataSource {
	return &aiDataSource{}
}

func (d *aiDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_ai"
}

func (d *aiDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type aiDataSourceModel struct {
	aiBaseModel
}

func (d *aiDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              schemaDataSourceAddonID("Addon AI"),
			"location":        schemaDataSourceAddonLocation("Addon AI"),
			"sku":             schemaDataSourceAddonSKU("Addon AI"),
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon AI"),
			"url":             schemaDataSourceAddonURL("Addon AI"),
		},
		MarkdownDescription: "Get information about an existing Addon AI.",
	}
}

func (d *aiDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data aiDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewAIOp(d.client)
	result := getAddon(ctx, "AI", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeAIResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon AI[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
