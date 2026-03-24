// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type traceStorageAccessKeyDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &traceStorageAccessKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &traceStorageAccessKeyDataSource{}
)

func NewTraceStorageAccessKeyDataSource() datasource.DataSource {
	return &traceStorageAccessKeyDataSource{}
}

func (d *traceStorageAccessKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_trace_storage_access_key"
}

func (d *traceStorageAccessKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type traceStorageAccessKeyDataSourceModel struct {
	accessKeyBaseModel
}

func (d *traceStorageAccessKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The UID of the trace storage access key.",
			},
			"storage_id": schema.StringAttribute{
				Required:    true,
				Description: "The trace storage ID for the access key.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the access key.",
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The token of the access key.",
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret of the access key.",
			},
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite trace storage access key.",
	}
}

func (d *traceStorageAccessKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data traceStorageAccessKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uid, err := parseUUID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: ID Error", fmt.Sprintf("invalid access key id: %s", err))
		return
	}

	op := monitoringsuite.NewTracesStorageOp(d.client)
	key, err := op.ReadKey(ctx, data.StorageID.ValueString(), uid)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read trace storage access key[%s]: %s", data.ID.ValueString(), err))
		return
	}

	updateAccessKeyState(&data.accessKeyBaseModel, data.StorageID.ValueString(), key.GetUID().String(), key.GetDescription().Value, key.GetToken(), key.GetSecret())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
