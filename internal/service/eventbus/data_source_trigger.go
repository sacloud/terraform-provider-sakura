// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type triggerDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &triggerDataSource{}
	_ datasource.DataSourceWithConfigure = &triggerDataSource{}
)

func NewEventBusTriggerDataSource() datasource.DataSource {
	return &triggerDataSource{}
}

func (d *triggerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eventbus_trigger"
}

func (d *triggerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.EventBusClient
}

type triggerDataSourceModel struct {
	triggerBaseModel
}

func (d *triggerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	const resourceName = "EventBus Trigger"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId(resourceName),
			"name":        common.SchemaDataSourceName(resourceName),
			"description": common.SchemaDataSourceDescription(resourceName),
			"tags":        common.SchemaDataSourceTags(resourceName),
			"icon_id":     common.SchemaDataSourceIconID(resourceName),

			"process_configuration_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The ProcessConfiguration ID of the %s.", resourceName),
			},
			"source": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The source of the %s.", resourceName),
			},
			"types": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: desc.Sprintf("The types of the %s.", resourceName),
			},
			"conditions": schema.ListNestedAttribute{
				Computed:    true,
				Description: desc.Sprintf("The conditions of the %s.", resourceName),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The key of the condition for %s.", resourceName),
						},
						"op": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The operator of the condition for %s.", resourceName),
						},
						"values": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: desc.Sprintf("The values of the condition for %s.", resourceName),
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing EventBus Trigger.",
	}
}

func (d *triggerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data triggerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tags := common.TsetToStrings(data.Tags)
	if name == "" && len(tags) == 0 {
		resp.Diagnostics.AddError("Invalid Attribute", "Either name or tags must be specified.")
		return
	}

	triggerOp := eventbus.NewTriggerOp(d.client)
	triggers, err := triggerOp.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("could not find SakuraCloud EventBus Trigger resource: %s", err))
		return
	}

	for _, t := range triggers {
		if name != "" && t.Name != name {
			continue
		}

		tagsMatched := true
		for _, tagToFind := range tags {
			if slices.Contains(t.Tags, tagToFind) {
				continue
			}
			tagsMatched = false
			break
		}
		if !tagsMatched {
			continue
		}

		if err := data.updateState(&t); err != nil {
			resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update EventBus Trigger[%s] state: %s", data.ID.String(), err))
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError("API Error", fmt.Sprintf("could not find any SakuraCloud EventBus Trigger resources with name=%q and tags=%v", name, tags))
}
