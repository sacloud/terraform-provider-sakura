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

type searchDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &searchDataSource{}
	_ datasource.DataSourceWithConfigure = &searchDataSource{}
)

func NewSearchDataSource() datasource.DataSource {
	return &searchDataSource{}
}

func (d *searchDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_search"
}

func (d *searchDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type searchDataSourceModel struct {
	searchBaseModel
}

func (d *searchDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       schemaDataSourceAddonID("Addon Search"),
			"location": schemaDataSourceAddonLocation("Addon Search"),
			"partition_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The partition count of the Addon Search.",
			},
			"replica_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The replica count of the Addon Search.",
			},
			"sku":             schemaDataSourceAddonSKU("Addon Search"),
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon Search"),
			"url":             schemaDataSourceAddonURL("Addon Search"),
		},
		MarkdownDescription: "Get information about an existing Addon Search.",
	}
}

func (d *searchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data searchDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewSearchOp(d.client)
	result := getAddon(ctx, "Addon Search", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeSearchResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon Search[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
