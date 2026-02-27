// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/addon-api-go"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
)

type cdnDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &cdnDataSource{}
	_ datasource.DataSourceWithConfigure = &cdnDataSource{}
)

func NewCDNDataSource() datasource.DataSource {
	return &cdnDataSource{}
}

func (d *cdnDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_cdn"
}

func (d *cdnDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type cdnDataSourceModel struct {
	cdnBaseModel
}

func (d *cdnDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":            schemaDataSourceAddonID("Addon CDN"),
			"location":      schemaDataSourceAddonLocation("Addon CDN"),
			"pricing_level": schemaDataSourceAddonPricingLevel("Addon CDN"),
			"patterns": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The route patterns of the Addon CDN.",
			},
			"origin": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The origin settings of the Addon CDN.",
				Attributes: map[string]schema.Attribute{
					"hostname": schema.StringAttribute{
						Computed:    true,
						Description: "The origin host name.",
					},
					"host_header": schema.StringAttribute{
						Computed:    true,
						Description: "The origin host header.",
					},
				},
			},
			"deployment_name": schemaResourceAddonDeploymentName("Addon CDN"),
			"url":             schemaDataSourceAddonURL("Addon CDN"),
		},
		MarkdownDescription: "Get information about an existing Addon CDN.",
	}
}

func (d *cdnDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cdnDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewCDNOp(d.client)
	result := getAddon(ctx, "CDN", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	var body v1.NetworkRequestBody
	err := decodeCDNFamilyResponse(result, &body)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon CDN[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
