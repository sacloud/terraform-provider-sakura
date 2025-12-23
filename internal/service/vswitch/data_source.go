// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package vswitch

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type vSwitchDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &vSwitchDataSource{}
	_ datasource.DataSourceWithConfigure = &vSwitchDataSource{}
)

func NewvSwitchDataSource() datasource.DataSource {
	return &vSwitchDataSource{}
}

func (d *vSwitchDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vswitch"
}

func (d *vSwitchDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type vSwitchDataSourceModel struct {
	vSwitchBaseModel
}

func (d *vSwitchDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("vSwitch"),
			"name":        common.SchemaDataSourceName("vSwitch"),
			"description": common.SchemaDataSourceDescription("vSwitch"),
			"tags":        common.SchemaDataSourceTags("vSwitch"),
			"icon_id":     common.SchemaDataSourceIconID("vSwitch"),
			"zone":        common.SchemaDataSourceZone("vSwitch"),
			"bridge_id": schema.StringAttribute{
				Computed:    true,
				Description: "The bridge id attached to the vSwitch.",
			},
			"server_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of server id connected to the vSwitch",
			},
		},
		MarkdownDescription: "Get information about an existing vSwitch(Switch in v2).",
	}
}

func (d *vSwitchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vSwitchDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewSwitchOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read vSwitch: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.Switches) == 0 {
		resp.Diagnostics.AddError("Read: Search Error", "No vSwitch resource matched the filter.")
		return
	}

	sw := res.Switches[0]
	if err := data.updateState(ctx, d.client, sw, zone); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update state for vSwitch[%s]: %s", sw.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
