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

type streamingDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &streamingDataSource{}
	_ datasource.DataSourceWithConfigure = &streamingDataSource{}
)

func NewStreamingDataSource() datasource.DataSource {
	return &streamingDataSource{}
}

func (d *streamingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_streaming"
}

func (d *streamingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type streamingDataSourceModel struct {
	streamingBaseModel
}

func (d *streamingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       schemaDataSourceAddonID("Addon Streaming"),
			"location": schemaDataSourceAddonLocation("Addon Streaming"),
			"unit_count": schema.StringAttribute{
				Computed:    true,
				Description: "The unit count of the Addon Streaming.",
			},
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon Streaming"),
			"url":             schemaDataSourceAddonURL("Addon Streaming"),
		},
		MarkdownDescription: "Get information about an existing Addon Streaming.",
	}
}

func (d *streamingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data streamingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewStreamingOp(d.client)
	result := getAddon(ctx, "Streaming", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	body, err := decodeStreamingResponse(result)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon Streaming[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
