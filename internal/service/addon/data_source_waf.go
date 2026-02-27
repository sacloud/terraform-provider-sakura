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

type wafDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &wafDataSource{}
	_ datasource.DataSourceWithConfigure = &wafDataSource{}
)

func NewWAFDataSource() datasource.DataSource {
	return &wafDataSource{}
}

func (d *wafDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_waf"
}

func (d *wafDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureAddonDataSourceClient(req, resp)
}

type wafDataSourceModel struct {
	wafBaseModel
}

func (d *wafDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":            schemaDataSourceAddonID("Addon WAF"),
			"location":      schemaDataSourceAddonLocation("Addon WAF"),
			"pricing_level": schemaDataSourceAddonPricingLevel("Addon WAF"),
			"patterns": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The route patterns of the Addon WAF.",
			},
			"origin": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The origin settings of the Addon WAF.",
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
			"deployment_name": schemaDataSourceAddonDeploymentName("Addon WAF"),
			"url":             schemaDataSourceAddonURL("Addon WAF"),
		},
		MarkdownDescription: "Get information about an existing Addon WAF.",
	}
}

func (d *wafDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data wafDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	op := addon.NewWAFOp(d.client)
	result := getAddon(ctx, "WAF", id, op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	var body v1.WafRequestBody
	err := decodeCDNFamilyResponse(result, &body)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon WAF[%s] response: %s", id, err))
		return
	}

	data.updateState(id, "", result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
