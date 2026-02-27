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

type dataLakeDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &dataLakeDataSource{}
	_ datasource.DataSourceWithConfigure = &dataLakeDataSource{}
)

func NewDataLakeDataSource() datasource.DataSource {
	return &dataLakeDataSource{}
}

func (d *dataLakeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_datalake"
}

func (d *dataLakeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type dataLakeDataSourceModel struct {
	dataLakeBaseModel
}

func (d *dataLakeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       schemaDataSourceAddonID("Addon DataLake"),
			"location": schemaDataSourceAddonLocation("Addon DataLake"),
			"performance": schema.Int32Attribute{
				Computed:    true,
				Description: "The performance setting of the Addon DataLake.",
			},
			"redundancy": schema.Int32Attribute{
				Computed:    true,
				Description: "The redundancy setting of the Addon DataLake.",
			},
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon DataLake"),
			"url":             schemaDataSourceAddonURL("Addon DataLake"),
		},
		MarkdownDescription: "Get information about an existing Addon DataLake.",
	}
}

func (d *dataLakeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataLakeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewDataLakeOp(d.client)
	result := getAddon(ctx, "DataLake", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeDataLakeResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon DataLake[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
