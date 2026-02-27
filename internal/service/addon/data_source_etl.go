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

type etlDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &etlDataSource{}
	_ datasource.DataSourceWithConfigure = &etlDataSource{}
)

func NewETLDataSource() datasource.DataSource {
	return &etlDataSource{}
}

func (d *etlDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_etl"
}

func (d *etlDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type etlDataSourceModel struct {
	etlBaseModel
}

func (d *etlDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              schemaDataSourceAddonID("Addon ETL"),
			"location":        schemaDataSourceAddonLocation("Addon ETL"),
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon ETL"),
			"url":             schemaDataSourceAddonURL("Addon ETL"),
		},
		MarkdownDescription: "Get information about an existing Addon ETL.",
	}
}

func (d *etlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data etlDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewETLOp(d.client)
	result := getAddon(ctx, "ETL", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeETLResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon ETL[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
