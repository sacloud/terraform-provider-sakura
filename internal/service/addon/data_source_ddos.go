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

type ddosDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &ddosDataSource{}
	_ datasource.DataSourceWithConfigure = &ddosDataSource{}
)

func NewDDoSDataSource() datasource.DataSource {
	return &ddosDataSource{}
}

func (d *ddosDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_ddos"
}

func (d *ddosDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type ddosDataSourceModel struct {
	ddosBaseModel
}

func (d *ddosDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":            schemaDataSourceAddonID("Addon DDoS"),
			"location":      schemaDataSourceAddonLocation("Addon DDoS"),
			"pricing_level": schemaDataSourceAddonPricingLevel("Addon DDoS"),
			"patterns": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The route patterns of the Addon DDoS.",
			},
			"origin": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The origin settings of the Addon DDoS.",
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
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon DDoS"),
			"url":             schemaDataSourceAddonURL("Addon DDoS"),
		},
		MarkdownDescription: "Get information about an existing Addon DDoS.",
	}
}

func (d *ddosDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ddosDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewDDoSOp(d.client)
	result := getAddon(ctx, "DDoS", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	var body v1.DdosRequestBody
	err := decodeCDNFamilyResponse(result, &body)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon DDoS[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
