// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package enhanced_lb

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

type enhancedLBDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &enhancedLBDataSource{}
	_ datasource.DataSourceWithConfigure = &enhancedLBDataSource{}
)

func NewEnhancedLBDataSource() datasource.DataSource {
	return &enhancedLBDataSource{}
}

func (d *enhancedLBDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enhanced_lb"
}

func (d *enhancedLBDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type enhancedLBDataSourceModel struct {
	enhancedLBBaseModel
}

func (d *enhancedLBDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Enhanced LB"),
			"name":        common.SchemaDataSourceName("Enhanced LB"),
			"description": common.SchemaDataSourceDescription("Enhanced LB"),
			"tags":        common.SchemaDataSourceTags("Enhanced LB"),
			"icon_id":     common.SchemaDataSourceIconID("Enhanced LB"),
			"plan": schema.Int64Attribute{
				Computed:    true,
				Description: "The plan of the Enhanced LB",
			},
			"vip_failover": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable VIP fail-over",
			},
			"sticky_session": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable sticky session",
			},
			"gzip": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable gzip compression",
			},
			"backend_http_keep_alive": schema.StringAttribute{
				Computed:    true,
				Description: "Mode of http keep-alive with backend",
			},
			"proxy_protocol": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable proxy protocol v2",
			},
			"timeout": schema.Int64Attribute{
				Computed:    true,
				Description: "The timeout duration in seconds",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The name of region that the Enhanced LB is in. This will be one of [%s]", iaastypes.ProxyLBRegionStrings),
			},
			"syslog": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"server": schema.StringAttribute{
						Computed:    true,
						Description: "The address of syslog server",
					},
					"port": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of syslog port",
					},
				},
			},
			"bind_port": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"proxy_mode": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The proxy mode. This will be one of [%s]", iaastypes.ProxyLBProxyModeStrings),
						},
						"port": schema.Int32Attribute{
							Computed:    true,
							Description: "The number of listening port",
						},
						"redirect_to_https": schema.BoolAttribute{
							Computed:    true,
							Description: "The flag to enable redirection from http to https. This flag is used only when `proxy_mode` is `http`",
						},
						"support_http2": schema.BoolAttribute{
							Computed:    true,
							Description: "The flag to enable HTTP/2. This flag is used only when `proxy_mode` is `https`",
						},
						"ssl_policy": schema.StringAttribute{
							Computed:    true,
							Description: "The ssl policy",
						},
						"response_header": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"header": schema.StringAttribute{
										Computed:    true,
										Description: "The field name of HTTP header added to response by the Enhanced LB",
									},
									"value": schema.StringAttribute{
										Computed:    true,
										Description: "The field value of HTTP header added to response by the Enhanced LB",
									},
								},
							},
						},
					},
				},
			},
			"health_check": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Computed:    true,
						Description: desc.Sprintf("The protocol used for health checks. This will be one of [%s]", iaastypes.ProxyLBProtocolStrings),
					},
					"delay_loop": schema.Int64Attribute{
						Computed:    true,
						Description: "The interval in seconds between checks",
					},
					"host_header": schema.StringAttribute{
						Computed:    true,
						Description: "The value of host header send when checking by HTTP",
					},
					"path": schema.StringAttribute{
						Computed:    true,
						Description: "The path used when checking by HTTP",
					},
				},
			},
			"sorry_server": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"ip_address": schema.StringAttribute{
						Computed:    true,
						Description: "The IP address of the SorryServer. This will be used when all servers are down",
					},
					"port": schema.Int32Attribute{
						Computed:    true,
						Description: "The port number of the SorryServer. This will be used when all servers are down",
					},
				},
			},
			"certificate": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"server_cert": schema.StringAttribute{
						Computed:    true,
						Description: "The certificate for a server",
					},
					"intermediate_cert": schema.StringAttribute{
						Computed:    true,
						Description: "The intermediate certificate for a server",
					},
					"private_key": schema.StringAttribute{
						Computed:    true,
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
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"server_cert": schema.StringAttribute{
									Computed:    true,
									Description: "The certificate for a server",
								},
								"intermediate_cert": schema.StringAttribute{
									Computed:    true,
									Description: "The intermediate certificate for a server",
								},
								"private_key": schema.StringAttribute{
									Computed:    true,
									Sensitive:   true,
									Description: "The private key for a server",
								},
							},
						},
					},
				},
			},
			"server": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Computed:    true,
							Description: "The IP address of the destination server",
						},
						"port": schema.Int32Attribute{
							Computed:    true,
							Description: "The port number of the destination server",
						},
						"group": schema.StringAttribute{
							Computed:    true,
							Description: "The name of load balancing group. This is used when using rule-based load balancing",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "The flag to enable as destination of load balancing",
						},
					},
				},
			},
			"rule": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							Computed:    true,
							Description: "The value of HTTP host header that is used as condition of rule-based balancing",
						},
						"path": schema.StringAttribute{
							Computed:    true,
							Description: "The request path that is used as condition of rule-based balancing",
						},
						"source_ips": schema.StringAttribute{
							Computed:    true,
							Description: "IP address or CIDR block to which the rule will be applied",
						},
						"request_header_name": schema.StringAttribute{
							Computed:    true,
							Description: "The header name that the client will send when making a request",
						},
						"request_header_value": schema.StringAttribute{
							Computed:    true,
							Description: "The condition for the value of the request header specified by the request header name",
						},
						"request_header_value_ignore_case": schema.BoolAttribute{
							Computed:    true,
							Description: "Boolean value representing whether the request header value ignores case",
						},
						"request_header_value_not_match": schema.BoolAttribute{
							Computed:    true,
							Description: "Boolean value representing whether to apply the rules when the request header value conditions are met or when the conditions do not match",
						},
						"group": schema.StringAttribute{
							Computed:    true,
							Description: "The name of load balancing group. When Enhanced LB received request which matched to `host` and `path`, Enhanced LB forwards the request to servers that having same group name",
						},
						"action": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The type of action to be performed when requests matches the rule. This will be one of [%s]", iaastypes.ProxyLBRuleActionStrings()),
						},
						"redirect_location": schema.StringAttribute{
							Computed:    true,
							Description: "The URL to redirect to when the request matches the rule. see https://manual.sakura.ad.jp/cloud/appliance/enhanced-lb/#enhanced-lb-rule for details",
						},
						"redirect_status_code": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("HTTP status code for redirects sent when requests matches the rule. This will be one of [%s]", iaastypes.ProxyLBRedirectStatusCodeStrings()),
						},
						"fixed_status_code": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("HTTP status code for fixed response sent when requests matches the rule. This will be one of [%s]", iaastypes.ProxyLBFixedStatusCodeStrings()),
						},
						"fixed_content_type": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("Content-Type header value for fixed response sent when requests matches the rule. This will be one of [%s]", iaastypes.ProxyLBFixedContentTypeStrings()),
						},
						"fixed_message_body": schema.StringAttribute{
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
		},
	}
}

func (d *enhancedLBDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data enhancedLBDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewProxyLBOp(d.client)
	res, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Search Error", "could not find SakuraCloud Enhanced LB resource: "+err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.ProxyLBs) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	elb := res.ProxyLBs[0]
	if err := data.updateState(ctx, d.client, elb); err != nil {
		resp.Diagnostics.AddError("Read Error", "could not update SakuraCloud Enhanced LB state: "+err.Error())
		return
	}
	data.IconID = types.StringValue(elb.IconID.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
