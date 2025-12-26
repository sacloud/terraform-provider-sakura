// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_monitor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

var (
	_ datasource.DataSource              = &simpleMonitorDataSource{}
	_ datasource.DataSourceWithConfigure = &simpleMonitorDataSource{}
)

func NewSimpleMonitorDataSource() datasource.DataSource {
	return &simpleMonitorDataSource{}
}

type simpleMonitorDataSource struct {
	client *common.APIClient
}

type simpleMonitorDataSourceModel struct {
	simpleMonitorBaseModel
	Name types.String `tfsdk:"name"`
}

func (d *simpleMonitorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_monitor"
}

func (d *simpleMonitorDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

func (d *simpleMonitorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Simple Monitor"),
			"name":        common.SchemaDataSourceName("Simple Monitor"),
			"description": common.SchemaDataSourceDescription("Simple Monitor"),
			"tags":        common.SchemaDataSourceTags("Simple Monitor"),
			"icon_id":     common.SchemaDataSourceIconID("Simple Monitor"),
			"target": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The monitoring target of the simple monitor. This will be IP address or FQDN",
			},
			"delay_loop": schema.Int32Attribute{
				Computed:    true,
				Description: "The interval in seconds between checks",
			},
			"max_check_attempts": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of retry",
			},
			"retry_interval": schema.Int32Attribute{
				Computed:    true,
				Description: "The interval in seconds between retries",
			},
			"timeout": schema.Int32Attribute{
				Computed:    true,
				Description: "The timeout in seconds for monitoring",
			},
			"health_check": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Computed:    true,
						Description: desc.Sprintf("The protocol used for health checks. This will be one of [%s]", iaastypes.SimpleMonitorProtocolStrings),
					},
					"host_header": schema.StringAttribute{
						Computed:    true,
						Description: "The value of host header send when checking by HTTP/HTTPS",
					},
					"path": schema.StringAttribute{
						Computed:    true,
						Description: "The path used when checking by HTTP/HTTPS",
					},
					"status": schema.Int32Attribute{
						Computed:    true,
						Description: "The response-code to expect when checking by HTTP/HTTPS",
					},
					"contains_string": schema.StringAttribute{
						Computed:    true,
						Description: "The string that should be included in the response body when checking for HTTP/HTTPS",
					},
					"sni": schema.BoolAttribute{
						Computed:    true,
						Description: "The flag to enable SNI when checking by HTTP/HTTPS",
					},
					"username": schema.StringAttribute{
						Computed:    true,
						Description: "The user name for basic auth used when checking by HTTP/HTTPS",
					},
					"password": schema.StringAttribute{
						Computed:    true,
						Description: "The password for basic auth used when checking by HTTP/HTTPS",
					},
					"port": schema.Int32Attribute{
						Computed:    true,
						Description: "The port number used for monitoring",
					},
					"qname": schema.StringAttribute{
						Computed:    true,
						Description: "The FQDN used when checking by DNS",
					},
					"expected_data": schema.StringAttribute{
						Computed:    true,
						Description: "The expected value used when checking by DNS",
					},
					"community": schema.StringAttribute{
						Computed:    true,
						Description: "The SNMP community string used when checking by SNMP",
					},
					"snmp_version": schema.StringAttribute{
						Computed:    true,
						Description: "The SNMP version used when checking by SNMP",
					},
					"oid": schema.StringAttribute{
						Computed:    true,
						Description: "The SNMP OID used when checking by SNMP",
					},
					"remaining_days": schema.Int32Attribute{
						Computed:    true,
						Description: "The number of remaining days until certificate expiration used when checking SSL certificates",
					},
					"http2": schema.BoolAttribute{
						Computed:    true,
						Description: "The flag to enable HTTP/2 when checking by HTTPS",
					},
					"ftps": schema.StringAttribute{
						Computed:    true,
						Description: desc.Sprintf("The methods of invoking security for monitoring with FTPS. This will be one of [%s]", iaastypes.SimpleMonitorFTPSStrings),
					},
					"verify_sni": schema.BoolAttribute{
						Computed:    true,
						Description: "The flag to enable hostname verification for SNI",
					},
				},
			},
			"notify_email_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable notification by email",
			},
			"notify_email_html": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable HTML format instead of text format",
			},
			"notify_slack_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable notification by slack/discord",
			},
			"notify_slack_webhook": schema.StringAttribute{
				Computed:    true,
				Description: "The webhook URL for sending notification by slack/discord",
			},
			"notify_interval": schema.Int32Attribute{
				Computed:    true,
				Description: "The interval in hours between notification",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable monitoring by the simple monitor",
			},
			"monitoring_suite": common.SchemaDataSourceMonitoringSuite("Simple Monitor"),
		},
		MarkdownDescription: "Get information about an existing Simple Monitor.",
	}
}

func (d *simpleMonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data simpleMonitorDataSourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (data.Name.IsNull() && data.Name.IsUnknown()) && (data.Target.IsNull() && data.Target.IsUnknown()) {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'name' or 'target' must be specified.")
		return
	}
	name := data.Name
	if name.ValueString() == "" {
		name = data.Target
	}

	smOp := iaas.NewSimpleMonitorOp(d.client)
	res, err := smOp.Find(ctx, common.CreateFindCondition(data.ID, name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", "failed to find SimpleMonitor resource: "+err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.SimpleMonitors) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.SimpleMonitors[0])
	data.Name = types.StringValue(res.SimpleMonitors[0].Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
