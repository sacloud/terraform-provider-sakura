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

type logStorageDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &logStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &logStorageDataSource{}
)

func NewLogStorageDataSource() datasource.DataSource {
	return &logStorageDataSource{}
}

func (d *logStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_log_storage"
}

func (d *logStorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type logStorageDataSourceModel struct {
	logStorageBaseModel
}

func (d *logStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Monitoring Suite Log Storage"),
			"name":        common.SchemaDataSourceName("Monitoring Suite Log Storage"),
			"description": common.SchemaDataSourceDescription("Monitoring Suite Log Storage"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The project ID of the Log Storage.",
			},
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the Log Storage.",
			},
			"is_system": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to indicate whether this is a system Log Storage.",
			},
			"classification": schema.StringAttribute{
				Computed:    true,
				Description: "The bucket classification of the Log Storage.",
			},
			"expire_day": schema.Int64Attribute{
				Computed:    true,
				Description: "The expiration day of the Log Storage.",
			},
			"created_at": common.SchemaDataSourceCreatedAt("Monitoring Suite Log Storage"),
			"endpoints": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The endpoints of the Log Storage.",
				Attributes: map[string]schema.Attribute{
					"ingester": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "The ingester endpoint for the Log Storage.",
						Attributes: map[string]schema.Attribute{
							"address": schema.StringAttribute{
								Computed:    true,
								Description: "The ingester address for the Log Storage.",
							},
							"insecure": schema.BoolAttribute{
								Computed:    true,
								Description: "The flag to indicate whether the ingester uses insecure connection.",
							},
						},
					},
				},
			},
			"usage": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The usage of the Log Storage.",
				Attributes: map[string]schema.Attribute{
					"log_routings": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of Log Routings.",
					},
					"log_measure_rules": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of Log Measure Rules.",
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite Log Storage.",
	}
}

func (d *logStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data logStorageDataSourceModel
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

	op := monitoringsuite.NewLogsStorageOp(d.client)
	var storage *monitoringsuiteapi.LogStorage
	var err error
	if id != "" {
		storage, err = op.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Log Storage[%s]: %s", id, err))
			return
		}
	} else {
		storages, err := op.List(ctx, monitoringsuite.LogsStoragesListParams{})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Log Storage resources: %s", err))
			return
		}
		storage, err = filterLogStorageByName(storages, name)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	}

	data.updateState(storage)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterLogStorageByName(storages []monitoringsuiteapi.LogStorage, name string) (*monitoringsuiteapi.LogStorage, error) {
	match := slices.Collect(func(yield func(monitoringsuiteapi.LogStorage) bool) {
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
		return nil, fmt.Errorf("multiple Log Storages found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
