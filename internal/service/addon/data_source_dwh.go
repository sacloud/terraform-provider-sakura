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

type dwhDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &dwhDataSource{}
	_ datasource.DataSourceWithConfigure = &dwhDataSource{}
)

func NewDWHDataSource() datasource.DataSource {
	return &dwhDataSource{}
}

func (d *dwhDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_dwh"
}

func (d *dwhDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type dwhDataSourceModel struct {
	dwhBaseModel
}

func (d *dwhDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              schemaDataSourceAddonID("Addon DWH"),
			"location":        schemaDataSourceAddonLocation("Addon DWH"),
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon DWH"),
			"url":             schemaDataSourceAddonURL("Addon DWH"),
		},
		MarkdownDescription: "Get information about an existing Addon DWH.",
	}
}

func (d *dwhDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dwhDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewDWHOp(d.client)
	result := getAddon(ctx, "DWH", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeDWHResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon DWH[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
