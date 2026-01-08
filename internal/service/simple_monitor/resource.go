// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_monitor

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type simpleMonitorResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &simpleMonitorResource{}
	_ resource.ResourceWithConfigure   = &simpleMonitorResource{}
	_ resource.ResourceWithImportState = &simpleMonitorResource{}
)

func NewSimpleMonitorResource() resource.Resource {
	return &simpleMonitorResource{}
}

func (r *simpleMonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_monitor"
}

func (r *simpleMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type simpleMonitorResourceModel struct {
	simpleMonitorBaseModel
	HealthCheck *simpleMonitorHealthCheckResourceModel `tfsdk:"health_check"`
	Timeouts    timeouts.Value                         `tfsdk:"timeouts"`
}

type simpleMonitorHealthCheckResourceModel struct {
	simpleMonitorHealthCheckModel
	Password          types.String `tfsdk:"password"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int32  `tfsdk:"password_wo_version"`
}

func (r *simpleMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Simple Monitor"),
			"description": common.SchemaResourceDescription("Simple Monitor"),
			"tags":        common.SchemaResourceTags("Simple Monitor"),
			"icon_id":     common.SchemaResourceIconID("Simple Monitor"),
			"target": schema.StringAttribute{
				Required:    true,
				Description: "The monitoring target of the simple monitor. This must be IP address or FQDN",
			},
			"max_check_attempts": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(3),
				Description: desc.Sprintf("The number of retry. %s", desc.Range(1, 10)),
				Validators: []validator.Int32{
					int32validator.Between(1, 10),
				},
			},
			"retry_interval": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(10),
				Description: desc.Sprintf("The interval in seconds between retries. %s", desc.Range(10, 3600)),
				Validators: []validator.Int32{
					int32validator.Between(10, 3600),
				},
			},
			"timeout": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(10),
				Description: desc.Sprintf("The timeout in seconds for monitoring. %s", desc.Range(10, 30)),
				Validators: []validator.Int32{
					int32validator.Between(10, 30),
				},
			},
			"delay_loop": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(60),
				Description: desc.Sprintf("The interval in seconds between checks. %s", desc.Range(60, 3600)),
				Validators: []validator.Int32{
					int32validator.Between(60, 3600),
				},
			},
			"health_check": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The health check configuration for the simple monitor.",
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Required:    true,
						Description: desc.Sprintf("The protocol used for health checks. This must be one of [%s]", iaastypes.SimpleMonitorProtocolStrings),
						Validators: []validator.String{
							stringvalidator.OneOf(iaastypes.SimpleMonitorProtocolStrings...),
						},
					},
					"port": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The target port number",
					},
					"host_header": schema.StringAttribute{
						Optional:    true,
						Description: "The value of host header send when checking by HTTP/HTTPS",
					},
					"path": schema.StringAttribute{
						Optional:    true,
						Description: "The path used when checking by HTTP/HTTPS",
					},
					"status": schema.Int32Attribute{
						Optional:    true,
						Description: "The response-code to expect when checking by HTTP/HTTPS",
					},
					"contains_string": schema.StringAttribute{
						Optional:    true,
						Description: "The string that should be included in the response body when checking for HTTP/HTTPS",
					},
					"sni": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "The flag to enable SNI when checking by HTTP/HTTPS",
					},
					"username": schema.StringAttribute{
						Optional:    true,
						Description: "The user name for basic auth used when checking by HTTP/HTTPS",
					},
					"password": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "The password for basic auth used when checking by HTTP/HTTPS. Use password_wo instead for newer deployments.",
						Validators: []validator.String{
							stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("health_check").AtName("password_wo")),
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_wo")),
						},
					},
					"password_wo": schema.StringAttribute{
						Optional:    true,
						WriteOnly:   true,
						Description: "The password for basic auth used when checking by HTTP/HTTPS",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password")),
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
						},
					},
					"password_wo_version": schema.Int32Attribute{
						Optional:    true,
						Description: "The version of the password_wo field. This value must be greater than 0 when set. Increment this when changing password.",
						Validators: []validator.Int32{
							int32validator.AtLeast(1),
							int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo")),
						},
					},
					"qname": schema.StringAttribute{
						Optional:    true,
						Description: "The FQDN used when checking by DNS",
					},
					"expected_data": schema.StringAttribute{
						Optional:    true,
						Description: "The expected value used when checking by DNS",
					},
					"community": schema.StringAttribute{
						Optional:    true,
						Description: "The SNMP community string used when checking by SNMP",
					},
					"snmp_version": schema.StringAttribute{
						Optional:    true,
						Description: "The SNMP version used when checking by SNMP. This must be one of [1, 2c]",
						Validators:  []validator.String{stringvalidator.OneOf("1", "2c")},
					},
					"oid": schema.StringAttribute{
						Optional:    true,
						Description: "The SNMP OID used when checking by SNMP",
					},
					"remaining_days": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The number of remaining days until certificate expiration used when checking SSL certificates.",
						Validators: []validator.Int32{
							int32validator.Between(1, 9999),
						},
					},
					"http2": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "The flag to enable HTTP/2 when checking by HTTPS",
					},
					"ftps": schema.StringAttribute{
						Optional:    true,
						Description: "The methods of invoking security for monitoring with FTPS.",
						Validators:  []validator.String{stringvalidator.OneOf(iaastypes.SimpleMonitorFTPSStrings...)},
					},
					"verify_sni": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "The flag to enable hostname verification for SNI",
					},
				},
			},
			"notify_email_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to enable notification by email",
			},
			"notify_email_html": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to enable HTML format instead of text format",
			},
			"notify_slack_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to enable notification by slack/discord",
			},
			"notify_slack_webhook": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The webhook URL for sending notification by slack/discord",
			},
			"notify_interval": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(2),
				Validators:  []validator.Int32{int32validator.Between(1, 72)},
				Description: "The interval in hours between notification.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "The flag to enable monitoring by the simple monitor",
			},
			"monitoring_suite": common.SchemaResourceMonitoringSuite("Simple Monitor"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Simple Monitor.",
	}
}

func (r *simpleMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *simpleMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config simpleMonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	smOp := iaas.NewSimpleMonitorOp(r.client)
	created, err := smOp.Create(ctx, expandSimpleMonitorCreateRequest(&plan, &config))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create SimpleMonitor: %s", err))
		return
	}

	updateModel(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *simpleMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state simpleMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sm := getSimpleMonitor(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sm == nil {
		return
	}

	updateModel(&state, sm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *simpleMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config simpleMonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	sm := getSimpleMonitor(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sm == nil {
		return
	}

	smOp := iaas.NewSimpleMonitorOp(r.client)
	if _, err := smOp.Update(ctx, sm.ID, expandSimpleMonitorUpdateRequest(&plan, &config)); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update SimpleMonitor: %s", err))
		return
	}

	updated := getSimpleMonitor(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if updated == nil {
		return
	}

	updateModel(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *simpleMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state simpleMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	sm := getSimpleMonitor(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sm == nil {
		return
	}

	if err := iaas.NewSimpleMonitorOp(r.client).Delete(ctx, sm.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete SimpleMonitor: %s", err))
		return
	}
}

func getSimpleMonitor(ctx context.Context, client *common.APIClient, id string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.SimpleMonitor {
	smOp := iaas.NewSimpleMonitorOp(client)
	sm, err := smOp.Read(ctx, common.SakuraCloudID(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read SimpleMonitor[%s]: %s", id, err))
		return nil
	}
	return sm
}

func expandSimpleMonitorNotifyInterval(d *simpleMonitorResourceModel) int {
	return int(d.NotifyInterval.ValueInt32()) * 60 * 60 // hours => seconds
}

func expandSimpleMonitorCreateRequest(model, config *simpleMonitorResourceModel) *iaas.SimpleMonitorCreateRequest {
	return &iaas.SimpleMonitorCreateRequest{
		Target:             model.Target.ValueString(),
		Enabled:            iaastypes.StringFlag(model.Enabled.ValueBool()),
		HealthCheck:        expandSimpleMonitorHealthCheck(model, config),
		DelayLoop:          int(model.DelayLoop.ValueInt32()),
		MaxCheckAttempts:   int(model.MaxCheckAttempts.ValueInt32()),
		RetryInterval:      int(model.RetryInterval.ValueInt32()),
		Timeout:            int(model.Timeout.ValueInt32()),
		NotifyEmailEnabled: iaastypes.StringFlag(model.NotifyEmailEnabled.ValueBool()),
		NotifyEmailHTML:    iaastypes.StringFlag(model.NotifyEmailHTML.ValueBool()),
		NotifySlackEnabled: iaastypes.StringFlag(model.NotifySlackEnabled.ValueBool()),
		SlackWebhooksURL:   model.NotifySlackWebhook.ValueString(),
		NotifyInterval:     expandSimpleMonitorNotifyInterval(model),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		MonitoringSuiteLog: common.ExpandMonitoringSuiteLog(model.MonitoringSuite),
	}
}

func expandSimpleMonitorUpdateRequest(model, config *simpleMonitorResourceModel) *iaas.SimpleMonitorUpdateRequest {
	return &iaas.SimpleMonitorUpdateRequest{
		Enabled:            iaastypes.StringFlag(model.Enabled.ValueBool()),
		HealthCheck:        expandSimpleMonitorHealthCheck(model, config),
		DelayLoop:          int(model.DelayLoop.ValueInt32()),
		MaxCheckAttempts:   int(model.MaxCheckAttempts.ValueInt32()),
		RetryInterval:      int(model.RetryInterval.ValueInt32()),
		Timeout:            int(model.Timeout.ValueInt32()),
		NotifyEmailEnabled: iaastypes.StringFlag(model.NotifyEmailEnabled.ValueBool()),
		NotifyEmailHTML:    iaastypes.StringFlag(model.NotifyEmailHTML.ValueBool()),
		NotifySlackEnabled: iaastypes.StringFlag(model.NotifySlackEnabled.ValueBool()),
		SlackWebhooksURL:   model.NotifySlackWebhook.ValueString(),
		NotifyInterval:     expandSimpleMonitorNotifyInterval(model),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		MonitoringSuiteLog: common.ExpandMonitoringSuiteLog(model.MonitoringSuite),
	}
}

func expandSimpleMonitorHealthCheck(model, config *simpleMonitorResourceModel) *iaas.SimpleMonitorHealthCheck {
	conf := model.HealthCheck
	protocol := conf.Protocol.ValueString()
	port := conf.Port.ValueInt32()
	password := config.HealthCheck.PasswordWO.ValueString()
	if password == "" {
		password = conf.Password.ValueString()
	}

	switch protocol {
	case "http":
		if port == 0 {
			port = 80
		}

		return &iaas.SimpleMonitorHealthCheck{
			Protocol:          iaastypes.SimpleMonitorProtocols.HTTP,
			Port:              iaastypes.StringNumber(port),
			Path:              conf.Path.ValueString(),
			Status:            iaastypes.StringNumber(conf.Status.ValueInt32()),
			ContainsString:    conf.ContainsString.ValueString(),
			Host:              conf.HostHeader.ValueString(),
			BasicAuthUsername: conf.Username.ValueString(),
			BasicAuthPassword: password,
		}
	case "https":
		if port == 0 {
			port = 443
		}

		return &iaas.SimpleMonitorHealthCheck{
			Protocol:          iaastypes.SimpleMonitorProtocols.HTTPS,
			Port:              iaastypes.StringNumber(port),
			Path:              conf.Path.ValueString(),
			Status:            iaastypes.StringNumber(conf.Status.ValueInt32()),
			ContainsString:    conf.ContainsString.ValueString(),
			SNI:               iaastypes.StringFlag(conf.SNI.ValueBool()),
			Host:              conf.HostHeader.ValueString(),
			BasicAuthUsername: conf.Username.ValueString(),
			BasicAuthPassword: password,
			HTTP2:             iaastypes.StringFlag(conf.Http2.ValueBool()),
		}
	case "dns":
		return &iaas.SimpleMonitorHealthCheck{
			Protocol:     iaastypes.SimpleMonitorProtocols.DNS,
			QName:        conf.QName.ValueString(),
			ExpectedData: conf.ExpectedData.ValueString(),
		}
	case "snmp":
		return &iaas.SimpleMonitorHealthCheck{
			Protocol:     iaastypes.SimpleMonitorProtocols.SNMP,
			Community:    conf.Community.ValueString(),
			SNMPVersion:  conf.SnmpVersion.ValueString(),
			OID:          conf.Oid.ValueString(),
			ExpectedData: conf.ExpectedData.ValueString(),
		}
	case "tcp":
		return &iaas.SimpleMonitorHealthCheck{
			Protocol: iaastypes.SimpleMonitorProtocols.TCP,
			Port:     iaastypes.StringNumber(port),
		}
	case "ssh":
		if port == 0 {
			port = 22
		}
		return &iaas.SimpleMonitorHealthCheck{
			Protocol: iaastypes.SimpleMonitorProtocols.SSH,
			Port:     iaastypes.StringNumber(port),
		}
	case "smtp":
		if port == 0 {
			port = 25
		}
		return &iaas.SimpleMonitorHealthCheck{
			Protocol: iaastypes.SimpleMonitorProtocols.SMTP,
			Port:     iaastypes.StringNumber(port),
		}
	case "pop3":
		if port == 0 {
			port = 110
		}
		return &iaas.SimpleMonitorHealthCheck{
			Protocol: iaastypes.SimpleMonitorProtocols.POP3,
			Port:     iaastypes.StringNumber(port),
		}
	case "ping":
		return &iaas.SimpleMonitorHealthCheck{
			Protocol: iaastypes.SimpleMonitorProtocols.Ping,
		}
	case "sslcertificate":
		days := 30
		if !conf.RemainingDays.IsNull() && !conf.RemainingDays.IsUnknown() {
			days = int(conf.RemainingDays.ValueInt32())
		}
		return &iaas.SimpleMonitorHealthCheck{
			Protocol:      iaastypes.SimpleMonitorProtocols.SSLCertificate,
			RemainingDays: days,
			VerifySNI:     iaastypes.StringFlag(conf.VerifySni.ValueBool()),
		}
	case "ftp":
		if port == 0 {
			port = 21
		}
		ftps := ""
		if !conf.Ftps.IsNull() && !conf.Ftps.IsUnknown() {
			ftps = conf.Ftps.ValueString()
		}
		return &iaas.SimpleMonitorHealthCheck{
			Protocol: iaastypes.SimpleMonitorProtocols.FTP,
			Port:     iaastypes.StringNumber(port),
			FTPS:     iaastypes.ESimpleMonitorFTPS(ftps),
		}
	}

	return nil
}

func updateModel(model *simpleMonitorResourceModel, sm *iaas.SimpleMonitor) {
	model.updateState(sm)
	model.HealthCheck = flattenSimpleMonitorHealthCheckResource(model, sm)
}

func flattenSimpleMonitorHealthCheckResource(model *simpleMonitorResourceModel, sm *iaas.SimpleMonitor) *simpleMonitorHealthCheckResourceModel {
	res := simpleMonitorHealthCheckResourceModel{}
	res.updateState(sm)

	hc := sm.HealthCheck
	hcModel := model.HealthCheck
	switch sm.HealthCheck.Protocol {
	case iaastypes.SimpleMonitorProtocols.HTTP, iaastypes.SimpleMonitorProtocols.HTTPS:
		if hcModel.Password.ValueString() != "" {
			res.Password = types.StringValue(hc.BasicAuthPassword)
		} else {
			res.Password = types.StringNull()
		}
		if hcModel.PasswordWOVersion.ValueInt32() > 0 {
			res.PasswordWOVersion = hcModel.PasswordWOVersion
		}
	}

	return &res
}
