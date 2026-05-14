// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/webaccel-api-go"
)

type webAccelDataSource struct {
	client *webaccel.Client
}

var (
	_ datasource.DataSource              = &webAccelDataSource{}
	_ datasource.DataSourceWithConfigure = &webAccelDataSource{}
)

func NewWebAccelDataSource() datasource.DataSource {
	return &webAccelDataSource{}
}

func (d *webAccelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webaccel"
}

func (d *webAccelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.WebaccelClient
}

type webAccelDataSourceModel struct {
	webAccelBaseModel

	HasCertificate   types.Bool                `tfsdk:"has_certificate"`
	HostHeader       types.String              `tfsdk:"host_header"`
	Status           types.String              `tfsdk:"status"`
	Logging          *webAccelLoggingModel     `tfsdk:"logging"`
	OriginParameters *webAccelOriginParamModel `tfsdk:"origin_parameters"`
}

func (d *webAccelDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resourceName := "Web Accelerator"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaDataSourceId(resourceName),
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the site",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("domain")),
				},
			},
			"domain": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Domain name of the site. Required when domain_type is own_domain",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"origin": schema.StringAttribute{
				Computed:    true,
				Description: "Origin hostname or IP address (deprecated; use origin_parameters.origin)",
			},
			"request_protocol": schema.StringAttribute{
				Computed:    true,
				Description: "Request protocol of the site. This must be one of [http+https, https, https-redirect]",
			},
			"origin_parameters": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Origin parameters of the site",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:    true,
						Description: "Origin type of the site. This must be one of [web, bucket]",
					},
					"origin": schema.StringAttribute{
						Computed:    true,
						Description: "Origin hostname or IP address. Required for type = web",
					},
					"protocol": schema.StringAttribute{
						Computed:    true,
						Description: "Request protocol for the origin host. Required for type = web",
					},
					"host_header": schema.StringAttribute{
						Computed:    true,
						Description: "Host header to the origin. Optional for type = web",
					},
					"endpoint": schema.StringAttribute{
						Computed:    true,
						Description: "Object Storage's S3 endpoint without protocol scheme. Required for type = bucket",
					},
					"region": schema.StringAttribute{
						Computed:    true,
						Description: "Object Storage's S3 region. Required for type = bucket",
					},
					"bucket_name": schema.StringAttribute{
						Computed:    true,
						Description: "Object Storage's bucket name. Required for type = bucket",
					},
					"use_document_index": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the document indexing for the bucket is enabled or not. Optional for type = bucket",
					},
				},
			},
			"subdomain":       schema.StringAttribute{Computed: true},
			"domain_type":     schema.StringAttribute{Computed: true},
			"has_certificate": schema.BoolAttribute{Computed: true},
			"host_header":     schema.StringAttribute{Computed: true},
			"status":          schema.StringAttribute{Computed: true},
			"cname_record_value": schema.StringAttribute{
				Computed:    true,
				Description: "CNAME record value for the site",
			},
			"txt_record_value": schema.StringAttribute{
				Computed:    true,
				Description: "TXT record value for the site",
			},
			"logging": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Logging configuration of the site",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the site logging is enabled or not",
					},
					"endpoint": schema.StringAttribute{
						Computed:    true,
						Description: "Logging Object Storage's S3 endpoint",
					},
					"region": schema.StringAttribute{
						Computed:    true,
						Description: "Logging Object Storage's S3 region",
					},
					"bucket_name": schema.StringAttribute{
						Computed:    true,
						Description: "Logging Object Storage's bucket name",
					},
				},
			},
			"default_cache_ttl": schema.Int32Attribute{
				Computed:    true,
				Description: "The default cache TTL of the site",
			},
			"vary_support": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the site recognizes the Vary header or not",
			},
			"cors_rules": schema.SetNestedAttribute{
				Computed:    true,
				Description: "CORS rules of the site",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"allow_all": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the site permits cross origin requests for all or not",
						},
						"allowed_origins": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of allowed origins for CORS",
						},
					},
				},
			},
			"normalize_ae": schema.StringAttribute{
				Computed:    true,
				Description: "Accept-encoding normalization. This must be one of [gzip, br+gzip]",
			},
		},
		MarkdownDescription: "Get information about an existing Web Accelerator site.",
	}
}

func (d *webAccelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data webAccelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	domain := data.Domain.ValueString()
	if name == "" && domain == "" {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'name' or 'domain' must be specified")
		return
	}

	op := webaccel.NewOp(d.client)
	res, err := op.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list WebAccel sites: %s", err))
		return
	}
	if res == nil || len(res.Sites) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	var site *webaccel.Site
	for _, s := range res.Sites {
		if (name != "" && s.Name == name) || (domain != "" && s.Domain == domain) {
			site = s
			break
		}
	}
	if site == nil {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	if err := data.updateState(site); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update state from API response: %s", err))
		return
	}
	data.HasCertificate = types.BoolValue(site.HasCertificate)
	data.HostHeader = types.StringValue(site.HostHeader)
	data.Status = types.StringValue(site.Status)

	originParams, err := flattenWebAccelOriginParametersDataSource(site)
	if err != nil {
		resp.Diagnostics.AddError("Read: Invalid origin_parameters", err.Error())
		return
	}
	data.OriginParameters = originParams

	logCfg, err := op.ReadLogUploadConfig(ctx, site.ID)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read logging config for WebAccel site[%s]: %s", site.ID, err))
		return
	}
	if logCfg != nil && logCfg.Bucket == "" {
		logCfg = nil
	}
	data.Logging = flattenWebAccelLogUploadConfigDataSource(logCfg)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenWebAccelOriginParametersDataSource(site *webaccel.Site) (*webAccelOriginParamModel, error) {
	model := webAccelOriginParamModel{
		Type:             types.StringNull(),
		Origin:           types.StringNull(),
		Protocol:         types.StringNull(),
		HostHeader:       types.StringNull(),
		Endpoint:         types.StringNull(),
		Region:           types.StringNull(),
		BucketName:       types.StringNull(),
		UseDocumentIndex: types.BoolNull(),
	}

	switch site.OriginType {
	case webaccel.OriginTypesWebServer:
		model.Type = types.StringValue("web")
		model.Origin = types.StringValue(site.Origin)
		switch site.OriginProtocol {
		case webaccel.OriginProtocolsHttp:
			model.Protocol = types.StringValue("http")
		case webaccel.OriginProtocolsHttps:
			model.Protocol = types.StringValue("https")
		default:
			return nil, fmt.Errorf("invalid origin protocol: %s", site.OriginProtocol)
		}
		if site.HostHeader != "" {
			model.HostHeader = types.StringValue(site.HostHeader)
		}
	case webaccel.OriginTypesObjectStorage:
		model.Type = types.StringValue("bucket")
		model.Endpoint = types.StringValue(site.S3Endpoint)
		model.Region = types.StringValue(site.S3Region)
		model.BucketName = types.StringValue(site.BucketName)
		model.UseDocumentIndex = types.BoolValue(site.DocIndex == webaccel.DocIndexEnabled)
	default:
		return nil, fmt.Errorf("unknown origin type: %s", site.OriginType)
	}

	return &model, nil
}

func flattenWebAccelLogUploadConfigDataSource(cfg *webaccel.LogUploadConfig) *webAccelLoggingModel {
	if cfg == nil {
		return nil
	}

	ep, _ := strings.CutPrefix(cfg.Endpoint, "https://") // WebAccel API returns logging's endpoint with https:// prefix
	return &webAccelLoggingModel{
		Enabled:    types.BoolValue(cfg.Status == "enabled"),
		Endpoint:   types.StringValue(ep),
		Region:     types.StringValue(cfg.Region),
		BucketName: types.StringValue(cfg.Bucket),
	}
}
