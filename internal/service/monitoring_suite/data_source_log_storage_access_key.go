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

type logStorageAccessKeyDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &logStorageAccessKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &logStorageAccessKeyDataSource{}
)

func NewLogStorageAccessKeyDataSource() datasource.DataSource {
	return &logStorageAccessKeyDataSource{}
}

func (d *logStorageAccessKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_log_storage_access_key"
}

func (d *logStorageAccessKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type logStorageAccessKeyDataSourceModel struct {
	accessKeyBaseModel
}

func (d *logStorageAccessKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The UID of the log storage access key.",
			},
			"storage_id": schema.StringAttribute{
				Required:    true,
				Description: "The log storage ID for the access key.",
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
		MarkdownDescription: "Get information about an existing Monitoring Suite log storage access key.",
	}
}

func (d *logStorageAccessKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data logStorageAccessKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uid, err := parseUUID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: ID Error", fmt.Sprintf("invalid access key id: %s", err))
		return
	}

	op := monitoringsuite.NewLogsStorageOp(d.client)
	key, err := op.ReadKey(ctx, data.StorageID.ValueString(), uid)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read log storage access key[%s]: %s", data.ID.ValueString(), err))
		return
	}

	data.updateState(data.StorageID.ValueString(), key.GetUID().String(), key.GetDescription().Value, key.GetToken(), key.GetSecret().String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
