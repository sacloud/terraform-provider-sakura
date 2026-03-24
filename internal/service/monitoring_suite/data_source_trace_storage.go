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
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type traceStorageDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &traceStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &traceStorageDataSource{}
)

func NewTraceStorageDataSource() datasource.DataSource {
	return &traceStorageDataSource{}
}

func (d *traceStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_trace_storage"
}

func (d *traceStorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type traceStorageDataSourceModel struct {
	traceStorageBaseModel
}

func (d *traceStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Monitoring Suite trace storage"),
			"name":        common.SchemaDataSourceName("Monitoring Suite trace storage"),
			"description": common.SchemaDataSourceDescription("Monitoring Suite trace storage"),
			"tags":        common.SchemaDataSourceTags("Monitoring Suite trace storage"),
			"icon_id": schema.StringAttribute{
				Computed:    true,
				Description: "The icon ID of the trace storage.",
			},
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The account ID of the trace storage.",
			},
			"resource_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The resource ID of the trace storage.",
			},
			"retention_period_days": schema.Int64Attribute{
				Computed:    true,
				Description: "The retention period days of the trace storage.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp of the trace storage.",
			},
			"endpoints": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The endpoints of the trace storage.",
				Attributes: map[string]schema.Attribute{
					"ingester": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "The ingester endpoint for the trace storage.",
						Attributes: map[string]schema.Attribute{
							"address": schema.StringAttribute{
								Computed:    true,
								Description: "The ingester address for the trace storage.",
							},
							"insecure": schema.BoolAttribute{
								Computed:    true,
								Description: "The flag to indicate whether the ingester uses insecure connection.",
							},
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite trace storage.",
	}
}

func (d *traceStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data traceStorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()
	tags := common.TsetToStrings(data.Tags)
	if id == "" && name == "" && len(tags) == 0 {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'id', 'name', or 'tags' must be specified.")
		return
	}

	op := monitoringsuite.NewTracesStorageOp(d.client)
	var storage *monitoringsuiteapi.TraceStorage
	var err error
	if id != "" {
		storage, err = op.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read trace storage[%s]: %s", id, err))
			return
		}
	} else {
		storages, err := op.List(ctx, monitoringsuite.TracesStorageListParams{})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list trace storage resources: %s", err))
			return
		}
		storage, err = filterTraceStorageByNameAndTags(storages, name, tags)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	}

	updateTraceStorageState(&data.traceStorageBaseModel, storage)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterTraceStorageByNameAndTags(storages []monitoringsuiteapi.TraceStorage, name string, tags []string) (*monitoringsuiteapi.TraceStorage, error) {
	match := slices.Collect(func(yield func(monitoringsuiteapi.TraceStorage) bool) {
		for _, storage := range storages {
			if name != "" && storage.Name.Or("") != name {
				continue
			}
			if len(tags) > 0 && !utils.IsTagsMatched(tags, storage.Tags) {
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
		return nil, fmt.Errorf("multiple trace storages found with the same condition. name=%q tags=%v", name, tags)
	}
	return &match[0], nil
}
