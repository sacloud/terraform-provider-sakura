// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package enhanced_lb

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type enhancedLBResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &enhancedLBResource{}
	_ resource.ResourceWithConfigure   = &enhancedLBResource{}
	_ resource.ResourceWithImportState = &enhancedLBResource{}
)

func NewEnhancedLBResource() resource.Resource {
	return &enhancedLBResource{}
}

func (r *enhancedLBResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enhanced_lb"
}

func (r *enhancedLBResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type enhancedLBResourceModel struct {
	enhancedLBBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *enhancedLBResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Enhanced LB"),
			"name":        common.SchemaResourceName("Enhanced LB"),
			"description": common.SchemaResourceDescription("Enhanced LB"),
			"tags":        common.SchemaResourceTags("Enhanced LB"),
			"icon_id":     common.SchemaResourceIconID("Enhanced LB"),
			"plan": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: desc.ResourcePlan("Enhanced LB", iaastypes.ProxyLBPlanValues),
				Default:     int64default.StaticInt64(int64(iaastypes.ProxyLBPlans.CPS100.Int())),
				Validators: []validator.Int64{
					int64validator.OneOf(common.MapTo(iaastypes.ProxyLBPlanValues, common.IntToInt64)...),
				},
			},
			"vip_failover": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The flag to enable VIP fail-over",
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"sticky_session": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The flag to enable sticky session",
				Default:     booldefault.StaticBool(false),
			},
			"gzip": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The flag to enable gzip compression",
				Default:     booldefault.StaticBool(false),
			},
			"backend_http_keep_alive": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("Mode of http keep-alive with backend. This must be one of [%s]", iaastypes.ProxyLBBackendHttpKeepAliveStrings),
				Default:     stringdefault.StaticString(iaastypes.ProxyLBBackendHttpKeepAlive.Safe.String()),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.ProxyLBBackendHttpKeepAliveStrings...),
				},
			},
			"proxy_protocol": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The flag to enable proxy protocol v2",
				Default:     booldefault.StaticBool(false),
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The timeout duration in seconds",
				Default:     int64default.StaticInt64(10),
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The name of region that the Enhanced LB is in. This must be one of [%s]", iaastypes.ProxyLBRegionStrings),
				Default:     stringdefault.StaticString(iaastypes.ProxyLBRegions.IS1.String()),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.ProxyLBRegionStrings...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"syslog": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"server": schema.StringAttribute{
						Required:    true,
						Description: "The address of syslog server",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
					"port": schema.Int32Attribute{
						Required:    true,
						Description: "The number of syslog port",
						Validators: []validator.Int32{
							int32validator.Between(1, 65535),
						},
					},
				},
			},
			"bind_port": schema.ListNestedAttribute{
				Required: true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(2),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"proxy_mode": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The proxy mode. This must be one of [%s]", iaastypes.ProxyLBProxyModeStrings),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.ProxyLBProxyModeStrings...),
							},
						},
						"port": schema.Int32Attribute{
							Required:    true,
							Description: "The number of listening port",
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
							},
						},
						"redirect_to_https": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The flag to enable redirection from http to https. This flag is used only when `proxy_mode` is `http`",
							Default:     booldefault.StaticBool(false),
						},
						"support_http2": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The flag to enable HTTP/2. This flag is used only when `proxy_mode` is `https`",
							Default:     booldefault.StaticBool(false),
						},
						"ssl_policy": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: desc.Sprintf("The ssl policy. This must be one of [%s]", iaastypes.ProxyLBSSLPolicies),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.ProxyLBSSLPolicies...),
							},
						},
						"response_header": schema.ListNestedAttribute{
							Optional: true,
							Validators: []validator.List{
								listvalidator.SizeAtMost(10),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"header": schema.StringAttribute{
										Required:    true,
										Description: "The field name of HTTP header added to response by the Enhanced LB",
									},
									"value": schema.StringAttribute{
										Required:    true,
										Description: "The field value of HTTP header added to response by the Enhanced LB",
									},
								},
							},
						},
					},
				},
			},
			"health_check": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Required:    true,
						Description: desc.Sprintf("The protocol used for health checks. This must be one of [%s]", iaastypes.ProxyLBProtocolStrings),
						Validators: []validator.String{
							stringvalidator.OneOf(iaastypes.ProxyLBProtocolStrings...),
						},
					},
					"delay_loop": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: desc.Sprintf("The interval in seconds between checks. %s", desc.Range(10, 60)),
						Default:     int64default.StaticInt64(10),
						Validators: []validator.Int64{
							int64validator.Between(10, 60),
						},
					},
					"host_header": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The value of host header send when checking by HTTP",
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("path")),
						},
					},
					"path": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The path used when checking by HTTP",
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("host_header")),
						},
					},
				},
			},
			"sorry_server": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"ip_address": schema.StringAttribute{
						Required:    true,
						Description: "The IP address of the SorryServer. This will be used when all servers are down",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.Both),
						},
					},
					"port": schema.Int32Attribute{
						Required:    true,
						Description: "The port number of the SorryServer. This will be used when all servers are down",
						Validators: []validator.Int32{
							int32validator.Between(1, 65535),
						},
					},
				},
			},
			"certificate": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"server_cert": schema.StringAttribute{
						Required:    true,
						Description: "The certificate for a server",
					},
					"intermediate_cert": schema.StringAttribute{
						Required:    true,
						Description: "The intermediate certificate for a server",
					},
					"private_key": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "The private key for a server",
					},
					"common_name": schema.StringAttribute{
						Computed:    true,
						Description: "The common name of the certificate",
					},
					"subject_alt_names": schema.StringAttribute{
						Computed:    true,
						Description: "The subject alternative names of the certificate",
					},
					"additional_certificate": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"server_cert": schema.StringAttribute{
									Required:    true,
									Description: "The certificate for a server",
								},
								"intermediate_cert": schema.StringAttribute{
									Optional:    true,
									Description: "The intermediate certificate for a server",
								},
								"private_key": schema.StringAttribute{
									Required:    true,
									Sensitive:   true,
									Description: "The private key for a server",
								},
							},
						},
					},
				},
			},
			"server": schema.ListNestedAttribute{
				Optional: true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(40),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Required:    true,
							Description: "The IP address of the destination server",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.Both),
							},
						},
						"port": schema.Int32Attribute{
							Required:    true,
							Description: desc.Sprintf("The port number of the destination server. %s", desc.Range(1, 65535)),
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
							},
						},
						"group": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: desc.Sprintf("The name of load balancing group. This is used when using rule-based load balancing. %s", desc.Length(1, 10)),
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 10),
							},
						},
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The flag to enable as destination of load balancing",
							Default:     booldefault.StaticBool(true),
						},
					},
				},
			},
			"rule": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The value of HTTP host header that is used as condition of rule-based balancing",
						},
						"path": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The request path that is used as condition of rule-based balancing",
						},
						"source_ips": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "IP address or CIDR block to which the rule will be applied. Multiple values can be specified by separating them with a space or comma",
						},
						"request_header_name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The header name that the client will send when making a request",
						},
						"request_header_value": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The condition for the value of the request header specified by the request header name",
						},
						"request_header_value_ignore_case": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Boolean value representing whether the request header value ignores case",
							Default:     booldefault.StaticBool(false),
						},
						"request_header_value_not_match": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Boolean value representing whether to apply the rules when the request header value conditions are met or when the conditions do not match",
							Default:     booldefault.StaticBool(false),
						},
						"group": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: desc.Sprintf("The name of load balancing group. When enhanced LB received request which matched to `host` and `path`, enhanced LB forwards the request to servers that having same group name. %s", desc.Length(1, 10)),
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 10),
							},
						},
						"action": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: desc.Sprintf("The type of action to be performed when requests matches the rule. This must be one of [%s]", iaastypes.ProxyLBRuleActionStrings()),
							Default:     stringdefault.StaticString(iaastypes.ProxyLBRuleActions.Forward.String()),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.ProxyLBRuleActionStrings()...),
							},
						},
						"redirect_location": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The URL to redirect to when the request matches the rule. see https://manual.sakura.ad.jp/cloud/appliance/enhanced-lb/#enhanced-lb-rule for details",
						},
						"redirect_status_code": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: desc.Sprintf("HTTP status code for redirects sent when requests matches the rule. This must be one of [%s]", iaastypes.ProxyLBRedirectStatusCodeStrings()),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.ProxyLBRedirectStatusCodeStrings()...),
							},
						},
						"fixed_status_code": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: desc.Sprintf("HTTP status code for fixed response sent when requests matches the rule. This must be one of [%s]", iaastypes.ProxyLBFixedStatusCodeStrings()),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.ProxyLBFixedStatusCodeStrings()...),
							},
						},
						"fixed_content_type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: desc.Sprintf("Content-Type header value for fixed response sent when requests matches the rule. This must be one of [%s]", iaastypes.ProxyLBFixedContentTypeStrings()),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.ProxyLBFixedContentTypeStrings()...),
							},
						},
						"fixed_message_body": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Content body for fixed response sent when requests matches the rule",
						},
					},
				},
			},
			"letsencrypt": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "The flag to accept the current Let's Encrypt terms of service(see: https://letsencrypt.org/repository/). This must be set `true` explicitly",
					},
					"common_name": schema.StringAttribute{
						Computed:    true,
						Description: "The common name of the certificate",
					},
					"subject_alt_names": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "The subject alternative names of the certificate",
					},
				},
			},
			"fqdn": schema.StringAttribute{
				Computed:    true,
				Description: "The FQDN for accessing to the Enhanced LB. This is typically used as value of CNAME record",
			},
			"vip": schema.StringAttribute{
				Computed:    true,
				Description: "The virtual IP address assigned to the Enhanced LB",
			},
			"proxy_networks": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of CIDR block used by the Enhanced LB to access the server",
			},
			"monitoring_suite": common.SchemaResourceMonitoringSuite("Enhanced LB"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Enhanced Load Balancer(proxylb in v2).",
	}
}

func (r *enhancedLBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *enhancedLBResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, state *enhancedLBResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if state == nil || plan == nil {
		return
	}

	if plan.Plan.ValueInt64() != state.Plan.ValueInt64() {
		resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())
	}
}

func (r *enhancedLBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan enhancedLBResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	enhancedLBOp := iaas.NewProxyLBOp(r.client)
	elb, err := enhancedLBOp.Create(ctx, expandEnhancedLBCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Enhanced LB: %s", err))
		return
	}

	certs := expandEnhancedLBCerts(&plan)
	if certs != nil {
		_, err := enhancedLBOp.SetCertificates(ctx, elb.ID, &iaas.ProxyLBSetCertificatesRequest{
			PrimaryCerts:    certs.PrimaryCert,
			AdditionalCerts: certs.AdditionalCerts,
		})
		if err != nil {
			resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to set Certificates to Enhanced LB[%s]: %s", elb.ID, err))
			return
		}
	}

	if err := plan.updateState(ctx, r.client, elb); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", "failed to update Enhanced LB state: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *enhancedLBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state enhancedLBResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	elb := getELB(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if elb == nil {
		return
	}

	if err := state.updateState(ctx, r.client, elb); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", "failed to update Enhanced LB state: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *enhancedLBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state enhancedLBResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	enhancedLBOp := iaas.NewProxyLBOp(r.client)
	id := state.ID.ValueString()

	common.SakuraMutexKV.Lock(id)
	defer common.SakuraMutexKV.Unlock(id)

	elb := getELB(ctx, r.client, id, &resp.State, &resp.Diagnostics)
	if elb == nil {
		return
	}

	elb, err := enhancedLBOp.Update(ctx, common.SakuraCloudID(id), expandEnhancedLBUpdateRequest(&plan, &state))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Enhanced LB[%s]: %s", id, err))
		return
	}

	if !plan.Plan.Equal(state.Plan) {
		newPlan := iaastypes.EProxyLBPlan(plan.Plan.ValueInt64())
		serviceClass := iaastypes.ProxyLBServiceClass(newPlan, elb.Region)
		upd, err := enhancedLBOp.ChangePlan(ctx, elb.ID, &iaas.ProxyLBChangePlanRequest{ServiceClass: serviceClass})
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to change plan of Enhanced LB[%s]: %s", elb.ID, err))
			return
		}
		elb = upd
		plan.ID = types.StringValue(elb.ID.String())
	}

	if elb.LetsEncrypt == nil && !plan.Certificate.Equal(state.Certificate) {
		certs := expandEnhancedLBCerts(&plan)
		if certs == nil {
			if err := enhancedLBOp.DeleteCertificates(ctx, elb.ID); err != nil {
				resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to delete Certificates of Enhanced LB[%s]: %s", elb.ID, err))
				return
			}
		} else {
			_, err := enhancedLBOp.SetCertificates(ctx, elb.ID, &iaas.ProxyLBSetCertificatesRequest{
				PrimaryCerts:    certs.PrimaryCert,
				AdditionalCerts: certs.AdditionalCerts,
			})
			if err != nil {
				resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to set Certificates to Enhanced LB[%s]: %s", elb.ID, err))
				return
			}
		}
	}

	if err := plan.updateState(ctx, r.client, elb); err != nil {
		resp.Diagnostics.AddError("Update: Terraform Error", "failed to update Enhanced LB state: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *enhancedLBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state enhancedLBResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	enhancedLBOp := iaas.NewProxyLBOp(r.client)
	id := state.ID.ValueString()

	common.SakuraMutexKV.Lock(id)
	defer common.SakuraMutexKV.Unlock(id)

	elb := getELB(ctx, r.client, id, &resp.State, &resp.Diagnostics)
	if elb == nil {
		return
	}

	if err := enhancedLBOp.Delete(ctx, elb.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Enhanced LB[%s]: %s", elb.ID, err))
		return
	}
}

func getELB(ctx context.Context, client *common.APIClient, id string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.ProxyLB {
	enhancedLBOp := iaas.NewProxyLBOp(client)
	elb, err := enhancedLBOp.Read(ctx, common.SakuraCloudID(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Enhanced LB[%s]: %s", id, err))
		return nil
	}

	return elb
}
