// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
	"github.com/sacloud/webaccel-api-go"
)

type webAccelResource struct {
	client *webaccel.Client
}

var (
	_ resource.Resource                = &webAccelResource{}
	_ resource.ResourceWithConfigure   = &webAccelResource{}
	_ resource.ResourceWithImportState = &webAccelResource{}
)

func NewWebAccelResource() resource.Resource {
	return &webAccelResource{}
}

func (r *webAccelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webaccel"
}

func (r *webAccelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.WebaccelClient
}

type webAccelResourceModel struct {
	webAccelBaseModel
	OnetimeURLSecretsWO        types.List                        `tfsdk:"onetime_url_secrets_wo"`
	OnetimeURLSecretsWOVersion types.Int32                       `tfsdk:"onetime_url_secrets_wo_version"`
	Logging                    *webAccelLoggingWithKeysModel     `tfsdk:"logging"`
	OriginParameters           *webAccelOriginParamWithKeysModel `tfsdk:"origin_parameters"`
}

type webAccelOriginParamWithKeysModel struct {
	// go-cmp doesn't support embed struct, so redefine the fields here with the same tags.
	Type                 types.String `tfsdk:"type"`
	Origin               types.String `tfsdk:"origin"`
	Protocol             types.String `tfsdk:"protocol"`
	HostHeader           types.String `tfsdk:"host_header"`
	Endpoint             types.String `tfsdk:"endpoint"`
	Region               types.String `tfsdk:"region"`
	BucketName           types.String `tfsdk:"bucket_name"`
	UseDocumentIndex     types.Bool   `tfsdk:"use_document_index"`
	AccessKeyWO          types.String `tfsdk:"access_key_wo"`
	SecretAccessKeyWO    types.String `tfsdk:"secret_access_key_wo"`
	CredentialsWOVersion types.Int32  `tfsdk:"credentials_wo_version"`
}

type webAccelLoggingWithKeysModel struct {
	Enabled              types.Bool   `tfsdk:"enabled"`
	Endpoint             types.String `tfsdk:"endpoint"`
	Region               types.String `tfsdk:"region"`
	BucketName           types.String `tfsdk:"bucket_name"`
	AccessKeyWO          types.String `tfsdk:"access_key_wo"`
	SecretAccessKeyWO    types.String `tfsdk:"secret_access_key_wo"`
	CredentialsWOVersion types.Int32  `tfsdk:"credentials_wo_version"`
}

func (r *webAccelResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceName := "WebAccel"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaResourceId(resourceName),
			"name": common.SchemaResourceName(resourceName),
			"domain_type": schema.StringAttribute{
				Required:    true,
				Description: "Domain type of the site. This must be one of [subdomain, own_domain]",
				Validators: []validator.String{
					stringvalidator.OneOf("subdomain", "own_domain"),
				},
			},
			"origin": schema.StringAttribute{
				Computed:    true,
				Description: "Origin hostname or IP address (deprecated; use origin_parameters.origin)",
			},
			"subdomain": schema.StringAttribute{
				Computed:    true,
				Description: "Subdomain of the site",
			},
			"cname_record_value": schema.StringAttribute{
				Computed:    true,
				Description: "CNAME record value for the site",
			},
			"txt_record_value": schema.StringAttribute{
				Computed:    true,
				Description: "TXT record value for the site",
			},
			"domain": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Domain name of the site. Required when domain_type is own_domain",
			},
			"request_protocol": schema.StringAttribute{
				Required:    true,
				Description: "Request protocol of the site. This must be one of [http+https, https, https-redirect]",
				Validators: []validator.String{
					stringvalidator.OneOf("http+https", "https", "https-redirect"),
				},
			},
			"origin_parameters": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Origin parameters of the site",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:    true,
						Description: "Origin type of the site. This must be one of [web, bucket]",
						Validators: []validator.String{
							stringvalidator.OneOf("web", "bucket"),
						},
					},
					"origin": schema.StringAttribute{
						Optional:    true,
						Description: "Origin hostname or IP address. Required for type = web",
					},
					"protocol": schema.StringAttribute{
						Optional:    true,
						Description: "Request protocol for the origin host. Required for type = web",
						Validators: []validator.String{
							stringvalidator.OneOf("http", "https"),
						},
					},
					"host_header": schema.StringAttribute{
						Optional:    true,
						Description: "Host header to the origin. Optional for type = web",
					},
					"endpoint": schema.StringAttribute{
						Optional:    true,
						Description: "Object Storage's S3 endpoint without protocol scheme. Required for type = bucket",
						Validators: []validator.String{
							sacloudvalidator.HostnameValidator(),
						},
					},
					"region": schema.StringAttribute{
						Optional:    true,
						Description: "Object Storage's S3 region. Required for type = bucket",
					},
					"bucket_name": schema.StringAttribute{
						Optional:    true,
						Description: "Object Storage's bucket name. Required for type = bucket",
					},
					"access_key_wo": schema.StringAttribute{
						Optional:    true,
						WriteOnly:   true,
						Description: "Object Storage's access key. Required for type = bucket",
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("credentials_wo_version")),
						},
					},
					"secret_access_key_wo": schema.StringAttribute{
						Optional:    true,
						WriteOnly:   true,
						Description: "Object Storage's secret access key. Required for type = bucket",
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("credentials_wo_version")),
						},
					},
					"credentials_wo_version": schema.Int32Attribute{
						Optional:    true,
						Description: "The version of the credential fields. This value must be greater than 0 when set. Increment this when changing credentials.",
						Validators: []validator.Int32{
							int32validator.AtLeast(1),
							int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("access_key_wo")),
							int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_access_key_wo")),
						},
					},
					"use_document_index": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether the document indexing for the bucket is enabled or not. Optional for type = bucket",
					},
				},
			},
			"cors_rules": schema.SetNestedAttribute{
				Optional:    true,
				Description: "CORS rules of the site",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"allow_all": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Whether the site permits cross origin requests for all or not",
						},
						"allowed_origins": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "List of allowed origins for CORS",
						},
					},
				},
			},
			"logging": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Logging configuration of the site",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Required:    true,
						Description: "Whether the site logging is enabled or not",
					},
					"endpoint": schema.StringAttribute{
						Required:    true,
						Description: "Object Storage's S3 endpoint without protocol scheme",
					},
					"region": schema.StringAttribute{
						Required:    true,
						Description: "Object Storage's S3 region",
					},
					"bucket_name": schema.StringAttribute{
						Required:    true,
						Description: "Object Storage's bucket name",
					},
					"access_key_wo": schema.StringAttribute{
						Required:    true,
						WriteOnly:   true,
						Description: "Object Storage's access key",
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("credentials_wo_version")),
						},
					},
					"secret_access_key_wo": schema.StringAttribute{
						Required:    true,
						WriteOnly:   true,
						Description: "Object Storage's secret access key",
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("credentials_wo_version")),
						},
					},
					"credentials_wo_version": schema.Int32Attribute{
						Optional:    true,
						Description: "The version of the credentials fields. This value must be greater than 0 when set. Increment this when changing credentials.",
						Validators: []validator.Int32{
							int32validator.AtLeast(1),
							int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("access_key_wo")),
							int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_access_key_wo")),
						},
					},
				},
			},
			"onetime_url_secrets_wo": schema.ListAttribute{
				Optional:    true,
				WriteOnly:   true,
				ElementType: types.StringType,
				Description: "The site-wide onetime url secrets",
				Validators: []validator.List{
					listvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("onetime_url_secrets_wo_version")),
				},
			},
			"onetime_url_secrets_wo_version": schema.Int32Attribute{
				Optional:    true,
				Description: "The version of the onetime_url_secrets field. This value must be greater than 0 when set. Increment this when changing secrets.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
					int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("onetime_url_secrets_wo")),
				},
			},
			"vary_support": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the site recognizes the Vary header or not",
			},
			"default_cache_ttl": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(-1),
				Description: "The default cache TTL of the site",
				Validators: []validator.Int32{
					int32validator.Between(-1, 604800),
				},
			},
			"normalize_ae": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Accept-encoding normalization. This must be one of [gzip, br+gzip]",
				Validators: []validator.String{
					stringvalidator.OneOf("gzip", "br+gzip"),
				},
			},
		},
		MarkdownDescription: "Manages a WebAccel site.",
	}
}

func (r *webAccelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *webAccelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config webAccelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqCreate, diags := expandWebAccelOriginParamsForCreation(&plan, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqCreate.Name = plan.Name.ValueString()
	reqCreate.DomainType = plan.DomainType.ValueString()
	if utils.IsKnown(plan.Domain) {
		reqCreate.Domain = plan.Domain.ValueString()
	}

	reqCreate.RequestProtocol = expandWebAccelRequestProtocol(plan.RequestProtocol)

	if utils.IsKnown(plan.VarySupport) {
		reqCreate.VarySupport = expandWebAccelVarySupportParameter(plan.VarySupport)
	}
	if utils.IsKnown(plan.DefaultCacheTTL) {
		ttl := int(plan.DefaultCacheTTL.ValueInt32())
		reqCreate.DefaultCacheTTL = &ttl
	}
	if utils.IsKnown(plan.NormalizeAE) {
		reqCreate.NormalizeAE = expandWebAccelNormalizeAEParameter(plan.NormalizeAE)
	}

	// NOTE: WebAccel site creation API does not accept CORS, onetime secrets, or logging.
	// Apply them after create when configured.
	var (
		hasCorsRule         = len(plan.CorsRules) > 0
		hasOnetimeURLSecret = utils.IsKnown(config.OnetimeURLSecretsWO)
		hasLoggingConfig    = plan.Logging != nil
	)

	if hasCorsRule {
		// CORSルールのチェックを事前に行う。Create後にチェックするとサイトは作られるがエラーとなりリソースが残るため。
		_, err := expandWebAccelCORSParameters(&plan)
		if err != nil {
			resp.Diagnostics.AddError("Create: Invalid cors_rules", err.Error())
			return
		}
	}

	op := webaccel.NewOp(r.client)
	created, err := op.Create(ctx, reqCreate)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create WebAccel site: %s", err))
		return
	}

	if hasCorsRule || hasOnetimeURLSecret {
		updateReq := new(webaccel.UpdateSiteRequest)
		if hasCorsRule {
			corsRule, _ := expandWebAccelCORSParameters(&plan)
			updateReq.CORSRules = &[]*webaccel.CORSRule{corsRule}
		} else {
			updateReq.CORSRules = &[]*webaccel.CORSRule{}
		}
		if hasOnetimeURLSecret {
			updateReq.OnetimeURLSecrets = expandWebAccelOnetimeURLSecrets(config.OnetimeURLSecretsWO)
		} else {
			updateReq.OnetimeURLSecrets = &[]string{}
		}
		if _, err := op.Update(ctx, created.ID, updateReq); err != nil {
			resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update WebAccel site after create: %s", err))
			return
		}
	}

	if hasLoggingConfig {
		logReq := expandLoggingParameters(&plan, &config)
		if _, err := op.ApplyLogUploadConfig(ctx, created.ID, logReq); err != nil {
			resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to apply logging config for WebAccel site: %s", err))
			return
		}
	}

	state, err := op.Read(ctx, created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read created WebAccel site: %s", err))
		return
	}
	logUploadConfig, err := op.ReadLogUploadConfig(ctx, created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read logging config for WebAccel site: %s", err))
		return
	}
	if logUploadConfig != nil && logUploadConfig.Bucket == "" {
		logUploadConfig = nil
	}

	if err := plan.updateModel(state, logUploadConfig, &config); err != nil {
		resp.Diagnostics.AddError("Create: API Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webAccelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webAccelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := webaccel.NewOp(r.client)
	site, err := op.Read(ctx, state.ID.ValueString())
	if err != nil {
		if webaccel.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read WebAccel site[%s]: %s", state.ID.ValueString(), err))
		return
	}

	logUploadConfig, err := op.ReadLogUploadConfig(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read logging config for WebAccel site[%s]: %s", state.ID.ValueString(), err))
		return
	}
	if logUploadConfig != nil && logUploadConfig.Bucket == "" {
		logUploadConfig = nil
	}

	if err := state.updateModel(site, logUploadConfig, &state); err != nil {
		resp.Diagnostics.AddError("Read: API Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webAccelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config webAccelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := webaccel.NewOp(r.client)

	if isWebAccelSiteUpdateRequired(&plan, &state) {
		updateReq, diags := expandWebAccelOriginParametersForUpdate(&plan, &config, &state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.Name = plan.Name.ValueString()
		updateReq.RequestProtocol = expandWebAccelRequestProtocol(plan.RequestProtocol)

		if isVersionIncremented(plan.OnetimeURLSecretsWOVersion, state.OnetimeURLSecretsWOVersion) {
			updateReq.OnetimeURLSecrets = expandWebAccelOnetimeURLSecrets(config.OnetimeURLSecretsWO)
		}
		if utils.IsKnown(plan.VarySupport) {
			updateReq.VarySupport = expandWebAccelVarySupportParameter(plan.VarySupport)
		}
		if utils.IsKnown(plan.DefaultCacheTTL) {
			ttl := int(plan.DefaultCacheTTL.ValueInt32())
			updateReq.DefaultCacheTTL = &ttl
		}
		if utils.IsKnown(plan.NormalizeAE) {
			updateReq.NormalizeAE = expandWebAccelNormalizeAEParameter(plan.NormalizeAE)
		}

		if len(plan.CorsRules) > 0 {
			corsRule, err := expandWebAccelCORSParameters(&plan)
			if err != nil {
				resp.Diagnostics.AddError("Update: Invalid cors_rules", err.Error())
				return
			}
			updateReq.CORSRules = &[]*webaccel.CORSRule{corsRule}
		} else {
			updateReq.CORSRules = &[]*webaccel.CORSRule{}
		}

		if _, err := op.Update(ctx, state.ID.ValueString(), updateReq); err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update WebAccel site[%s]: %s", state.ID.ValueString(), err))
			return
		}
	}

	if utils.HasChange(plan.Logging, state.Logging) {
		if plan.Logging == nil {
			if err := op.DeleteLogUploadConfig(ctx, state.ID.ValueString()); err != nil {
				resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to delete logging config for WebAccel site[%s]: %s", state.ID.ValueString(), err))
				return
			}
		} else {
			logReq := expandLoggingParameters(&plan, &config)
			if _, err := op.ApplyLogUploadConfig(ctx, state.ID.ValueString(), logReq); err != nil {
				resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to apply logging config for WebAccel site[%s]: %s", state.ID.ValueString(), err))
				return
			}
		}
	}

	current, err := op.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read WebAccel site[%s]: %s", state.ID.ValueString(), err))
		return
	}
	logUploadConfig, err := op.ReadLogUploadConfig(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read logging config for WebAccel site[%s]: %s", state.ID.ValueString(), err))
		return
	}
	if logUploadConfig != nil && logUploadConfig.Bucket == "" {
		logUploadConfig = nil
	}

	if err := plan.updateModel(current, logUploadConfig, &config); err != nil {
		resp.Diagnostics.AddError("Update: API Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webAccelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webAccelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := webaccel.NewOp(r.client)
	if _, err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete WebAccel site[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func (m *webAccelResourceModel) updateModel(site *webaccel.Site, logUploadConfig *webaccel.LogUploadConfig, config *webAccelResourceModel) error {
	if err := m.updateState(site); err != nil {
		return fmt.Errorf("failed to update WebAccel site[%s]: %w", site.ID, err)
	}

	originParams, err := flattenWebAccelOriginParameters(config, site)
	if err != nil {
		return err
	}
	m.OriginParameters = originParams

	if logUploadConfig != nil {
		ver := m.Logging.CredentialsWOVersion // preserve the version in state when log upload config exists
		m.Logging = flattenWebAccelLogUploadConfig(logUploadConfig)
		m.Logging.CredentialsWOVersion = ver
	} else {
		m.Logging = nil
	}

	return nil
}

func isWebAccelSiteUpdateRequired(plan, state *webAccelResourceModel) bool {
	return !plan.Name.Equal(state.Name) ||
		!plan.RequestProtocol.Equal(state.RequestProtocol) ||
		utils.HasChange(plan.OriginParameters, state.OriginParameters, cmpopts.EquateComparable(webAccelOriginParamModel{})) ||
		utils.HasChange(plan.CorsRules, state.CorsRules) ||
		isOriginCredentialsUpdateRequired(plan, state) ||
		isVersionIncremented(plan.OnetimeURLSecretsWOVersion, state.OnetimeURLSecretsWOVersion) ||
		!plan.VarySupport.Equal(state.VarySupport) ||
		!plan.DefaultCacheTTL.Equal(state.DefaultCacheTTL) ||
		!plan.NormalizeAE.Equal(state.NormalizeAE)
}

func isOriginCredentialsUpdateRequired(plan, state *webAccelResourceModel) bool {
	if plan == nil || state == nil || plan.OriginParameters == nil || state.OriginParameters == nil {
		return false
	}
	if plan.OriginParameters.Type.ValueString() != "bucket" {
		return false
	}
	return isVersionIncremented(plan.OriginParameters.CredentialsWOVersion, state.OriginParameters.CredentialsWOVersion)
}

func isVersionIncremented(planVal, stateVal types.Int32) bool {
	if !utils.IsKnown(planVal) {
		return false
	}
	if !utils.IsKnown(stateVal) {
		return planVal.ValueInt32() > 0
	}
	return planVal.ValueInt32() > stateVal.ValueInt32()
}

func expandWebAccelOriginParamsForCreation(plan, config *webAccelResourceModel) (*webaccel.CreateSiteRequest, diag.Diagnostics) {
	req := new(webaccel.CreateSiteRequest)
	upd, diags := expandWebAccelOriginParametersForUpdate(plan, config, nil)
	if diags.HasError() {
		return nil, diags
	}
	req.OriginType = upd.OriginType
	req.Origin = upd.Origin
	req.OriginProtocol = upd.OriginProtocol
	req.HostHeader = upd.HostHeader
	req.S3Endpoint = upd.S3Endpoint
	req.S3Region = upd.S3Region
	req.BucketName = upd.BucketName
	req.AccessKeyID = upd.AccessKeyID
	req.SecretAccessKey = upd.SecretAccessKey
	req.DocIndex = upd.DocIndex
	if plan.OriginParameters.Type.ValueString() == "bucket" && config != nil && config.OriginParameters != nil {
		req.AccessKeyID = config.OriginParameters.AccessKeyWO.ValueString()
		req.SecretAccessKey = config.OriginParameters.SecretAccessKeyWO.ValueString()
	}

	return req, diags
}

func expandWebAccelOriginParametersForUpdate(plan, config, state *webAccelResourceModel) (*webaccel.UpdateSiteRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := new(webaccel.UpdateSiteRequest)
	originParam := plan.OriginParameters

	switch originParam.Type.ValueString() {
	case "web":
		req.OriginType = webaccel.OriginTypesWebServer
		req.Origin = originParam.Origin.ValueString()
		switch originParam.Protocol.ValueString() {
		case "http":
			req.OriginProtocol = webaccel.OriginProtocolsHttp
		case "https":
			req.OriginProtocol = webaccel.OriginProtocolsHttps
		default:
			diags.AddError("Invalid origin protocol", "origin_parameters.protocol must be one of http or https")
			return nil, diags
		}
		if !originParam.HostHeader.IsNull() && !originParam.HostHeader.IsUnknown() {
			req.HostHeader = originParam.HostHeader.ValueString()
		}
	case "bucket":
		req.OriginType = webaccel.OriginTypesObjectStorage
		req.S3Endpoint = originParam.Endpoint.ValueString()
		req.S3Region = originParam.Region.ValueString()
		req.BucketName = originParam.BucketName.ValueString()
		if config != nil && config.OriginParameters != nil && state != nil && state.OriginParameters != nil {
			if isVersionIncremented(plan.OriginParameters.CredentialsWOVersion, state.OriginParameters.CredentialsWOVersion) {
				req.AccessKeyID = config.OriginParameters.AccessKeyWO.ValueString()
				req.SecretAccessKey = config.OriginParameters.SecretAccessKeyWO.ValueString()
			}
		}
		if utils.IsKnown(originParam.UseDocumentIndex) && originParam.UseDocumentIndex.ValueBool() {
			req.DocIndex = webaccel.DocIndexEnabled
		} else {
			req.DocIndex = webaccel.DocIndexDisabled
		}
	default:
		diags.AddError("Invalid origin type", "origin_parameters.type must be one of web or bucket")
		return nil, diags
	}

	return req, diags
}

func expandWebAccelRequestProtocol(v types.String) string {
	switch v.ValueString() {
	case "http+https":
		return webaccel.RequestProtocolsHttpAndHttps
	case "https":
		return webaccel.RequestProtocolsHttpsOnly
	case "https-redirect":
		return webaccel.RequestProtocolsRedirectToHttps
	default:
		return ""
	}
}

func expandWebAccelCORSParameters(plan *webAccelResourceModel) (*webaccel.CORSRule, error) {
	rule := &webaccel.CORSRule{}

	corsRule := plan.CorsRules[0] // WebAccel site supports only one CORS rule, so take the first one if exists
	allowAll := !corsRule.AllowAll.IsNull() && !corsRule.AllowAll.IsUnknown() && corsRule.AllowAll.ValueBool()
	allowedOrigins := common.TlistToStrings(corsRule.AllowedOrigins)

	if allowAll && len(allowedOrigins) > 0 {
		return nil, fmt.Errorf("invalid cors_rules: allow_all and allowed_origins are mutually exclusive")
	}
	if !allowAll && len(allowedOrigins) == 0 {
		return nil, fmt.Errorf("invalid cors_rules: either allow_all or allowed_origins must be specified")
	}

	if allowAll {
		rule.AllowsAnyOrigin = true
	} else {
		rule.AllowedOrigins = allowedOrigins
	}
	return rule, nil
}

func expandLoggingParameters(plan, config *webAccelResourceModel) *webaccel.LogUploadConfig {
	logCfg := plan.Logging
	logCfgKeys := plan.Logging
	if config != nil && config.Logging != nil {
		logCfgKeys = config.Logging
	}

	req := new(webaccel.LogUploadConfig)
	if logCfg.Enabled.ValueBool() {
		req.Status = "enabled"
	} else {
		req.Status = "disabled"
	}
	req.Bucket = logCfg.BucketName.ValueString()
	req.AccessKeyID = logCfgKeys.AccessKeyWO.ValueString()
	req.SecretAccessKey = logCfgKeys.SecretAccessKeyWO.ValueString()
	req.Endpoint = "https://" + logCfg.Endpoint.ValueString()
	req.Region = logCfg.Region.ValueString()

	return req
}

func expandWebAccelOnetimeURLSecrets(list types.List) *[]string {
	secrets := common.TlistToStrings(list)
	return &secrets
}

func expandWebAccelVarySupportParameter(v types.Bool) string {
	if v.ValueBool() {
		return webaccel.VarySupportEnabled
	}
	return webaccel.VarySupportDisabled
}

func expandWebAccelNormalizeAEParameter(v types.String) string {
	switch v.ValueString() {
	case "gzip":
		return webaccel.NormalizeAEGz
	case "br+gzip":
		return webaccel.NormalizeAEBrGz
	default:
		return ""
	}
}

func flattenWebAccelOriginParameters(config *webAccelResourceModel, site *webaccel.Site) (*webAccelOriginParamWithKeysModel, error) {
	originParam := &webAccelOriginParamWithKeysModel{
		Type:                 types.StringNull(),
		Origin:               types.StringNull(),
		Protocol:             types.StringNull(),
		HostHeader:           types.StringNull(),
		Endpoint:             types.StringNull(),
		Region:               types.StringNull(),
		BucketName:           types.StringNull(),
		UseDocumentIndex:     types.BoolNull(),
		CredentialsWOVersion: types.Int32Null(),
	}

	switch site.OriginType {
	case webaccel.OriginTypesWebServer:
		originParam.Type = types.StringValue("web")
		originParam.Origin = types.StringValue(site.Origin)
		switch site.OriginProtocol {
		case webaccel.OriginProtocolsHttp:
			originParam.Protocol = types.StringValue("http")
		case webaccel.OriginProtocolsHttps:
			originParam.Protocol = types.StringValue("https")
		default:
			return nil, fmt.Errorf("invalid web origin protocol: %s", site.OriginProtocol)
		}
		if site.HostHeader != "" {
			originParam.HostHeader = types.StringValue(site.HostHeader)
		}
	case webaccel.OriginTypesObjectStorage:
		originParam.Type = types.StringValue("bucket")
		originParam.Endpoint = types.StringValue(site.S3Endpoint)
		originParam.Region = types.StringValue(site.S3Region)
		originParam.BucketName = types.StringValue(site.BucketName)

		if config == nil || config.OriginParameters == nil {
			return nil, fmt.Errorf("origin_parameters must be provided to keep bucket credentials")
		}
		confParam := config.OriginParameters
		if utils.IsKnown(confParam.UseDocumentIndex) {
			originParam.UseDocumentIndex = types.BoolValue(confParam.UseDocumentIndex.ValueBool())
		}
		if utils.IsKnown(confParam.CredentialsWOVersion) {
			originParam.CredentialsWOVersion = types.Int32Value(confParam.CredentialsWOVersion.ValueInt32())
		}
	default:
		return nil, fmt.Errorf("unknown origin type: %s", site.OriginType)
	}

	return originParam, nil
}

func flattenWebAccelLogUploadConfig(cfg *webaccel.LogUploadConfig) *webAccelLoggingWithKeysModel {
	if cfg == nil {
		return nil
	}

	ep, _ := strings.CutPrefix(cfg.Endpoint, "https://") // WebAccel API returns logging's endpoint with https:// prefix
	return &webAccelLoggingWithKeysModel{
		Enabled:              types.BoolValue(cfg.Status == "enabled"),
		Endpoint:             types.StringValue(ep),
		Region:               types.StringValue(cfg.Region),
		BucketName:           types.StringValue(cfg.Bucket),
		CredentialsWOVersion: types.Int32Null(),
	}
}
