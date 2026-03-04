// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	simplenotification "github.com/sacloud/simple-notification-api-go"
	v1 "github.com/sacloud/simple-notification-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type routingDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &routingDataSource{}
	_ datasource.DataSourceWithConfigure = &routingDataSource{}
)

func NewRoutingDataSource() datasource.DataSource {
	return &routingDataSource{}
}

func (d *routingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_notification_routing"
}

func (d *routingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.SimpleNotificationClient
}

type routingDataSourceModel struct {
	routingBaseModel
}

func (d *routingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	const resourceName = "SimpleNotification Routing"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId(resourceName),
			"name":        common.SchemaDataSourceName(resourceName),
			"description": common.SchemaDataSourceDescription(resourceName),
			"tags":        common.SchemaDataSourceTags(resourceName),
			"icon_id":     common.SchemaDataSourceIconID(resourceName),
			"match_labels": schema.ListNestedAttribute{
				Computed:    true,
				Description: desc.Sprintf("The type of the %s.", resourceName),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The key of the match label.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the match label.",
						},
					},
				},
			},
			"source_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The value of the %s.", resourceName),
			},
			"target_group_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The value of the %s.", resourceName),
			},
		},
		MarkdownDescription: "Get information about an existing SimpleNotification Routing.",
	}
}

func (d *routingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data routingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tags := common.TsetToStrings(data.Tags)
	if name == "" && len(tags) == 0 {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'name' or 'tags' must be specified.")
		return
	}

	routingOp := simplenotification.NewRoutingOp(d.client)
	destListRes, err := routingOp.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list SimpleNotification ProcessConfiguration resources: %s", err))
		return
	}

	for _, dest := range destListRes.CommonServiceItems {
		if name != "" && dest.Name != name {
			continue
		}

		tagsMatched := utils.IsTagsMatched(tags, dest.Tags)
		if !tagsMatched {
			continue
		}

		if err := data.updateState(&dest); err != nil {
			resp.Diagnostics.AddError("Read: State Error", fmt.Sprintf("failed to update state from API response: %s", err))
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError("Read: Search Error", fmt.Sprintf("failed to find any SimpleNotification routing resources with name=%q and tags=%v", name, tags))
}
