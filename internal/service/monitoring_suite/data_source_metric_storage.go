// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type metricStorageDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &metricStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &metricStorageDataSource{}
)

func NewMetricStorageDataSource() datasource.DataSource {
	return &metricStorageDataSource{}
}

func (d *metricStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_metric_storage"
}

func (d *metricStorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type metricStorageDataSourceModel struct {
	metricStorageBaseModel
}

func (d *metricStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Monitoring Suite metric storage"),
			"name":        common.SchemaDataSourceName("Monitoring Suite metric storage"),
			"description": common.SchemaDataSourceDescription("Monitoring Suite metric storage"),
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The account ID of the metric storage.",
			},
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the metric storage.",
			},
			"is_system": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to indicate whether this is a system metric storage.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp of the metric storage.",
			},
			"endpoints": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The endpoints of the metric storage.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Computed:    true,
						Description: "The address of the metric storage endpoint.",
					},
				},
			},
			"usage": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The usage of the metric storage.",
				Attributes: map[string]schema.Attribute{
					"metric_routings": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of metric routings.",
					},
					"alert_rules": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of alert rules.",
					},
					"log_measure_rules": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of log measure rules.",
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite metric storage.",
	}
}

func (d *metricStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data metricStorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'id' or 'name' must be specified.")
		return
	}

	op := monitoringsuite.NewMetricsStorageOp(d.client)
	var storage *monitoringsuiteapi.MetricsStorage
	var err error
	if id != "" {
		storage, err = op.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read metric storage[%s]: %s", id, err))
			return
		}
	} else {
		storages, err := op.List(ctx, monitoringsuite.MetricsStorageListParams{})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list metric storage resources: %s", err))
			return
		}
		storage, err = filterMetricsStorageByName(storages, name)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	}

	data.updateState(storage)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterMetricsStorageByName(storages []monitoringsuiteapi.MetricsStorage, name string) (*monitoringsuiteapi.MetricsStorage, error) {
	match := slices.Collect(func(yield func(monitoringsuiteapi.MetricsStorage) bool) {
		for _, storage := range storages {
			if storage.Name.Value != name {
				continue
			}
			if !yield(storage) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple metric storages found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
