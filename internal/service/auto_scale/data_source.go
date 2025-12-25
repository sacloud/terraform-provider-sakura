// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package auto_scale

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type autoScaleDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &autoScaleDataSource{}
	_ datasource.DataSourceWithConfigure = &autoScaleDataSource{}
)

func NewAutoScaleDataSource() datasource.DataSource {
	return &autoScaleDataSource{}
}

func (d *autoScaleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auto_scale"
}

func (d *autoScaleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type autoScaleDataSourceModel struct {
	autoScaleBaseModel
}

func (r *autoScaleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("AutoScale"),
			"name":        common.SchemaDataSourceName("AutoScale"),
			"description": common.SchemaDataSourceDescription("AutoScale"),
			"tags":        common.SchemaDataSourceTags("AutoScale"),
			"icon_id":     common.SchemaDataSourceIconID("AutoScale"),
			"zones": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of zone names where monitored resources are located",
			},
			"config": schema.StringAttribute{
				Computed:    true,
				Description: "The configuration file for sacloud/autoscaler",
			},
			"api_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the API key",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to enable AutoScale",
			},
			"trigger_type": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("This must be one of [%s]", []string{"cpu", "router", "schedule", "none"}),
			},
			"router_threshold_scaling": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"router_prefix": schema.StringAttribute{
						Computed:    true,
						Description: "Router name prefix to be monitored",
					},
					"direction": schema.StringAttribute{
						Computed:    true,
						Description: desc.Sprintf("This must be one of [%s]", []string{"in", "out"}),
					},
					"mbps": schema.Int32Attribute{
						Computed:    true,
						Description: "Mbps",
					},
				},
			},
			"cpu_threshold_scaling": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"server_prefix": schema.StringAttribute{
						Computed:    true,
						Description: "Server name prefix to be monitored",
					},
					"up": schema.Int32Attribute{
						Computed:    true,
						Description: "Threshold for average CPU utilization to scale up/out",
					},
					"down": schema.Int32Attribute{
						Computed:    true,
						Description: "Threshold for average CPU utilization to scale down/in",
					},
				},
			},
			"schedule_scaling": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("This must be one of [%s]", []string{"up", "down"}),
						},
						"hour": schema.Int32Attribute{
							Computed:    true,
							Description: "Hour to be triggered",
						},
						"minute": schema.Int32Attribute{
							Computed:    true,
							Description: desc.Sprintf("Minute to be triggered. This must be one of [%s]", []string{"0", "15", "30", "45"}),
						},
						"days_of_week": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: desc.Sprintf("A set of days of week to backed up. The values in the list must be in [%s]", iaastypes.DaysOfTheWeekStrings),
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing AutoScale.",
	}
}

func (d *autoScaleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data autoScaleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewAutoScaleOp(d.client)
	res, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", "failed to find AutoScale: "+err.Error())
	}
	if res == nil || res.Count == 0 || len(res.AutoScale) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.AutoScale[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
