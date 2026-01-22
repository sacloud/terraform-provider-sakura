// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/path"

	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type apigwServiceResource struct {
	client *v1.Client
}

func NewApigwServiceResource() resource.Resource {
	return &apigwServiceResource{}
}

var (
	_ resource.Resource                = &apigwServiceResource{}
	_ resource.ResourceWithConfigure   = &apigwServiceResource{}
	_ resource.ResourceWithImportState = &apigwServiceResource{}
)

func (r *apigwServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_service"
}

func (r *apigwServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwServiceResourceModel struct {
	apigwServiceBaseModel
	ObjectStorageConfig *apigwServiceObjectStorageResourceModel `tfsdk:"object_storage_config"`
	Timeouts            timeouts.Value                          `tfsdk:"timeouts"`
}

type apigwServiceObjectStorageResourceModel struct {
	apigwServiceObjectStorageModel
	AccessKeyWO          types.String `tfsdk:"access_key_wo"`
	SecretAccessKeyWO    types.String `tfsdk:"secret_access_key_wo"`
	CredentialsWOVersion types.Int32  `tfsdk:"credentials_wo_version"`
}

func (r *apigwServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("API Gateway Service"),
			"name":       common.SchemaResourceName("API Gateway Service"),
			"tags":       common.SchemaResourceTags("API Gateway Service"),
			"created_at": schemaResourceAPIGWCreatedAt("API Gateway Service"),
			"updated_at": schemaResourceAPIGWUpdatedAt("API Gateway Service"),
			"subscription_id": schema.StringAttribute{
				Required:    true,
				Description: "The subscription plan ID associated with the service",
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(uuid.Validate),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: "The protocol used by the backend (http or https)",
				Validators: []validator.String{
					stringvalidator.OneOf("http", "https"),
				},
			},
			"host": schema.StringAttribute{
				Required:    true,
				Description: "The host name of the backend.",
				Validators: []validator.String{
					sacloudvalidator.HostnameValidator(),
				},
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/"),
				Description: "The base path for the backend",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"port": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The port of the backend",
				Validators: []validator.Int32{
					int32validator.Between(0, 65535),
				},
			},
			"retries": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(5),
				Description: "The number of retries for backend requests",
				Validators: []validator.Int32{
					int32validator.Between(0, 32767),
				},
			},
			"connect_timeout": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Connect timeout in milliseconds for the backend",
			},
			"write_timeout": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Write timeout in milliseconds for the backend",
			},
			"read_timeout": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Read timeout in milliseconds for the backend",
			},
			"authentication": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("none"),
				Description: desc.Sprintf("Authentication method for the backend. This can be one of %s.", serviceAuthTypes),
				Validators: []validator.String{
					stringvalidator.OneOf(serviceAuthTypes...),
				},
			},
			"route_host": schema.StringAttribute{
				Computed:    true,
				Description: "The route host for the service",
			},
			"oidc": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "OIDC authentication configuration",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required:    true,
						Description: "The entity ID of OIDC authentication",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the OIDC authentication",
					},
				},
			},
			"cors_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "CORS configuration for the service",
				Attributes: map[string]schema.Attribute{
					"credentials": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether to allow credentials",
					},
					"access_control_exposed_headers": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Headers exposed to the client",
						Validators: []validator.String{
							stringvalidator.LengthAtMost(4000),
						},
					},
					"access_control_allow_headers": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Allowed request headers",
						Validators: []validator.String{
							stringvalidator.LengthAtMost(4000),
						},
					},
					"max_age": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(0),
						Description: "Max age for CORS",
					},
					"access_control_allow_methods": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(common.StringsToTset(common.MapTo(v1.HTTPMethodGET.AllValues(), common.ToString))),
						Description: "Allowed HTTP methods for CORS",
					},
					"access_control_allow_origins": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Allowed origins for CORS",
					},
					"preflight_continue": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to pass preflight result to next handler",
					},
					"private_network": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to restrict CORS to private network",
					},
				},
			},
			"object_storage_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Object Storage configuration used by the service",
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required:    true,
						Description: "The bucket name",
					},
					"folder": schema.StringAttribute{
						Optional:    true,
						Description: "The folder name within the bucket",
					},
					"endpoint": schema.StringAttribute{
						Required:    true,
						Description: "The object storage endpoint",
					},
					"region": schema.StringAttribute{
						Required:    true,
						Description: "The object storage region",
					},
					"access_key_wo": schema.StringAttribute{
						Required:    true,
						WriteOnly:   true,
						Description: "Access key for object storage",
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_access_key_wo")),
						},
					},
					"secret_access_key_wo": schema.StringAttribute{
						Required:    true,
						WriteOnly:   true,
						Description: "Secret access key for object storage",
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("access_key_wo")),
						},
					},
					"credentials_wo_version": schema.Int32Attribute{
						Optional:    true,
						Description: "The version of the credentials. This value must be greater than 0 when set. Increment this when changing credentials",
					},
					"use_document_index": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Whether to use document index for object storage",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manage an API Gateway Service.",
	}
}

func (r *apigwServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *apigwServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config apigwServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	serviceOp := apigw.NewServiceOp(r.client)
	created, err := serviceOp.Create(ctx, expandAPIGWServiceCreateRequest(&plan, &config))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create API Gateway service: %s", err))
		return
	}

	service := getAPIGWService(ctx, r.client, created.ID.Value.String(), &resp.State, &resp.Diagnostics)
	if service == nil {
		return
	}

	plan.updateState(service)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apigwServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := getAPIGWService(ctx, r.client, data.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if service == nil {
		return
	}

	data.updateState(service)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apigwServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config apigwServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	service := getAPIGWService(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if service == nil {
		return
	}

	serviceOp := apigw.NewServiceOp(r.client)
	err := serviceOp.Update(ctx, expandAPIGWServiceUpdateRequest(&plan, &config), service.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update API Gateway service: %s", err))
		return
	}

	service = getAPIGWService(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if service == nil {
		return
	}

	plan.updateState(service)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apigwServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	service := getAPIGWService(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if service == nil {
		return
	}

	serviceOp := apigw.NewServiceOp(r.client)
	err := serviceOp.Delete(ctx, service.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete API Gateway service[%s]: %s", service.ID.Value.String(), err))
		return
	}
}

func getAPIGWService(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.ServiceDetailResponse {
	serviceOp := apigw.NewServiceOp(client)
	service, err := serviceOp.Read(ctx, uuid.MustParse(id))
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read APIGW service[%s]: %s", id, err))
		return nil
	}

	return service
}

func expandAPIGWServiceCreateRequest(plan, config *apigwServiceResourceModel) *v1.ServiceDetailRequest {
	res := &v1.ServiceDetailRequest{
		Name:                v1.Name(plan.Name.ValueString()),
		Tags:                common.TsetToStrings(plan.Tags),
		Host:                plan.Host.ValueString(),
		Path:                v1.NewOptString(plan.Path.ValueString()),
		Protocol:            v1.ServiceDetailRequestProtocol(plan.Protocol.ValueString()),
		Retries:             v1.NewOptInt(int(plan.Retries.ValueInt32())),
		Authentication:      v1.NewOptServiceDetailRequestAuthentication(v1.ServiceDetailRequestAuthentication(plan.Authentication.ValueString())),
		CorsConfig:          expandCorsConfig(plan.CORSConfig),
		ObjectStorageConfig: expandObjectStorageConfig(plan.ObjectStorageConfig, config.ObjectStorageConfig),
		Subscription: v1.ServiceSubscriptionRequest{
			ID: uuid.MustParse(plan.SubscriptionID.ValueString()),
		},
	}
	if utils.IsKnown(plan.Port) {
		res.Port = v1.NewOptInt(int(plan.Port.ValueInt32()))
	}
	if utils.IsKnown(plan.ConnectTimeout) {
		res.ConnectTimeout = v1.NewOptInt(int(plan.ConnectTimeout.ValueInt32()))
	}
	if utils.IsKnown(plan.WriteTimeout) {
		res.WriteTimeout = v1.NewOptInt(int(plan.WriteTimeout.ValueInt32()))
	}
	if utils.IsKnown(plan.ReadTimeout) {
		res.ReadTimeout = v1.NewOptInt(int(plan.ReadTimeout.ValueInt32()))
	}
	if plan.Authentication.ValueString() == "oidc" {
		res.Oidc = v1.NewOptOidcSummary(v1.OidcSummary{
			ID: v1.NewOptUUID(uuid.MustParse(plan.OIDC.ID.ValueString())),
		})
	}
	return res
}

// This method is used by resource_route.go as well
func ToHTTPMethod[S ~string](s S) v1.HTTPMethod {
	return v1.HTTPMethod(s)
}

func expandCorsConfig(model *apigwServiceCORSModel) v1.OptCorsConfig {
	res := v1.OptCorsConfig{}
	if model == nil {
		return res
	}
	conf := v1.CorsConfig{
		Credentials:               v1.NewOptBool(model.Credentials.ValueBool()),
		MaxAge:                    v1.NewOptInt32(model.MaxAge.ValueInt32()),
		AccessControlAllowMethods: common.MapTo(common.TsetToStrings(model.AccessControlAllowMethods), ToHTTPMethod),
	}
	if utils.IsKnown(model.AccessControlAllowHeaders) {
		conf.AccessControlAllowHeaders = v1.NewOptString(model.AccessControlAllowHeaders.ValueString())
	}
	if utils.IsKnown(model.AccessControlAllowOrigins) {
		conf.AccessControlAllowOrigins = v1.NewOptString(model.AccessControlAllowOrigins.ValueString())
	}
	if utils.IsKnown(model.AccessControlExposedHeaders) {
		conf.AccessControlExposedHeaders = v1.NewOptString(model.AccessControlExposedHeaders.ValueString())
	}
	if utils.IsKnown(model.PreflightContinue) {
		conf.PreflightContinue = v1.NewOptBool(model.PreflightContinue.ValueBool())
	}
	if utils.IsKnown(model.PrivateNetwork) {
		conf.PrivateNetwork = v1.NewOptBool(model.PrivateNetwork.ValueBool())
	}
	res.SetTo(conf)
	return res
}

func expandObjectStorageConfig(plan, config *apigwServiceObjectStorageResourceModel) v1.OptObjectStorageConfig {
	res := v1.OptObjectStorageConfig{}
	if plan == nil {
		return res
	}
	conf := v1.ObjectStorageConfig{
		BucketName:       plan.Bucket.ValueString(),
		Endpoint:         plan.Endpoint.ValueString(),
		Region:           plan.Region.ValueString(),
		AccessKeyID:      config.AccessKeyWO.ValueString(),
		SecretAccessKey:  config.SecretAccessKeyWO.ValueString(),
		UseDocumentIndex: plan.UseDocumentIndex.ValueBool(),
	}
	if utils.IsKnown(plan.Folder) {
		conf.FolderName = v1.NewOptString(plan.Folder.ValueString())
	}

	res.SetTo(conf)
	return res
}

func expandAPIGWServiceUpdateRequest(plan, config *apigwServiceResourceModel) *v1.ServiceDetail {
	res := &v1.ServiceDetail{
		Name:                v1.Name(plan.Name.ValueString()),
		Tags:                common.TsetToStrings(plan.Tags),
		Host:                plan.Host.ValueString(),
		Path:                v1.NewOptString(plan.Path.ValueString()),
		Port:                v1.NewOptInt(int(plan.Port.ValueInt32())),
		Protocol:            v1.ServiceDetailProtocol(plan.Protocol.ValueString()),
		Retries:             v1.NewOptInt(int(plan.Retries.ValueInt32())),
		CorsConfig:          expandCorsConfig(plan.CORSConfig),
		ObjectStorageConfig: expandObjectStorageConfig(plan.ObjectStorageConfig, config.ObjectStorageConfig),
	}
	if utils.IsKnown(plan.Port) {
		res.Port = v1.NewOptInt(int(plan.Port.ValueInt32()))
	}
	if utils.IsKnown(plan.ConnectTimeout) {
		res.ConnectTimeout = v1.NewOptInt(int(plan.ConnectTimeout.ValueInt32()))
	}
	if utils.IsKnown(plan.WriteTimeout) {
		res.WriteTimeout = v1.NewOptInt(int(plan.WriteTimeout.ValueInt32()))
	}
	if utils.IsKnown(plan.ReadTimeout) {
		res.ReadTimeout = v1.NewOptInt(int(plan.ReadTimeout.ValueInt32()))
	}
	if plan.Authentication.ValueString() == "oidc" {
		res.Oidc = v1.NewOptOidcSummary(v1.OidcSummary{
			ID: v1.NewOptUUID(uuid.MustParse(plan.OIDC.ID.ValueString())),
		})
	}
	return res
}
