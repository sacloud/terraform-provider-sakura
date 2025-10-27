// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/minio/minio-go/v7/pkg/cors"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageBucketCorsResource struct{}

var (
	_ resource.Resource                = &objectStorageBucketCorsResource{}
	_ resource.ResourceWithImportState = &objectStorageBucketCorsResource{}
)

func NewObjectStorageBucketCorsResource() resource.Resource {
	return &objectStorageBucketCorsResource{}
}

func (r *objectStorageBucketCorsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_bucket_cors"
}

type objectStorageBucketCorsResourceModel struct {
	objectStorageS3CompatModel
	CorsRules []*objectStorageBucketCorsRuleModel `tfsdk:"cors_rules"`
	Timeouts  timeouts.Value                      `tfsdk:"timeouts"`
}

type objectStorageBucketCorsRuleModel struct {
	AllowedMethods types.Set    `tfsdk:"allowed_methods"`
	AllowedOrigins types.Set    `tfsdk:"allowed_origins"`
	AllowedHeaders types.Set    `tfsdk:"allowed_headers"`
	ExposeHeaders  types.Set    `tfsdk:"expose_headers"`
	MaxAgeSeconds  types.Int32  `tfsdk:"max_age_seconds"`
	ID             types.String `tfsdk:"id"`
}

func (r *objectStorageBucketCorsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("Object Storage Bucket CORS"),
			"endpoint":   SchemaResourceEndpoint("Object Storage Bucket CORS"),
			"access_key": SchemaResourceAccessKey("Object Storage Bucket CORS"),
			"secret_key": SchemaResourceSecretKey("Object Storage Bucket CORS"),
			"bucket":     SchemaResourceBucket("Object Storage Bucket CORS"),
			"cors_rules": schema.ListNestedAttribute{
				Required:    true,
				Description: "The CORS rules for the Object Storage Bucket.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"allowed_methods": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "The set of HTTP methods that are allowed to access the origin.",
						},
						"allowed_origins": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "The set of origins that are allowed to access to the bucket.",
						},
						"allowed_headers": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "The set of headers used in `Access-Control-Request-Headers`",
						},
						"expose_headers": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "The set of headers in the response that users can access from the application.",
						},
						"max_age_seconds": schema.Int32Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The time in seconds that the browser is to cache the preflight response.",
						},
						"id": schema.StringAttribute{
							Optional:    true,
							Description: "The ID of the CORS rule.",
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *objectStorageBucketCorsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *objectStorageBucketCorsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStorageBucketCorsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	corsConf, err := setBucketCorsConfiguration(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	plan.updateS3State()
	plan.CorsRules = flattenCorsConfiguration(corsConf)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketCorsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStorageBucketCorsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := state.getMinIOClient()
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Errorf("failed to create MinIO client: %w", err).Error())
		return
	}

	corsConf, err := client.GetBucketCors(ctx, state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to get bucket CORS configuration: %s", err.Error()))
		return
	}

	state.updateS3State()
	state.CorsRules = flattenCorsConfiguration(corsConf)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStorageBucketCorsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan objectStorageBucketCorsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	corsConf, err := setBucketCorsConfiguration(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}

	plan.updateS3State()
	plan.CorsRules = flattenCorsConfiguration(corsConf)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageBucketCorsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectStorageBucketCorsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := state.getMinIOClient()
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Errorf("failed to create MinIO client: %w", err).Error())
		return
	}

	err = client.SetBucketCors(ctx, state.Bucket.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete bucket CORS configuration: %s", err.Error()))
		return
	}
}

func setBucketCorsConfiguration(ctx context.Context, model *objectStorageBucketCorsResourceModel) (*cors.Config, error) {
	client, err := model.getMinIOClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	err = client.SetBucketCors(ctx, model.Bucket.ValueString(), expandCorsConfiguration(model))
	if err != nil {
		return nil, fmt.Errorf("failed to set bucket CORS configuration: %w", err)
	}

	corsConf, err := client.GetBucketCors(ctx, model.Bucket.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket CORS configuration: %w", err)
	}

	return corsConf, nil
}

func expandCorsConfiguration(model *objectStorageBucketCorsResourceModel) *cors.Config {
	rules := make([]cors.Rule, 0, len(model.CorsRules))
	for _, rule := range model.CorsRules {
		r := cors.Rule{
			AllowedMethod: common.TsetToStrings(rule.AllowedMethods),
			AllowedOrigin: common.TsetToStrings(rule.AllowedOrigins),
			ExposeHeader:  common.TsetToStrings(rule.ExposeHeaders),
		}
		if !rule.AllowedHeaders.IsNull() && !rule.AllowedHeaders.IsUnknown() {
			r.AllowedHeader = common.TsetToStrings(rule.AllowedHeaders)
		}
		if !rule.MaxAgeSeconds.IsNull() && !rule.MaxAgeSeconds.IsUnknown() {
			r.MaxAgeSeconds = int(rule.MaxAgeSeconds.ValueInt32())
		}
		if !rule.ID.IsNull() && !rule.ID.IsUnknown() {
			r.ID = rule.ID.ValueString()
		}
		rules = append(rules, r)
	}
	return cors.NewConfig(rules)
}

func flattenCorsConfiguration(conf *cors.Config) []*objectStorageBucketCorsRuleModel {
	rules := make([]*objectStorageBucketCorsRuleModel, 0, len(conf.CORSRules))
	for _, rule := range conf.CORSRules {
		r := &objectStorageBucketCorsRuleModel{
			AllowedMethods: common.StringsToTset(rule.AllowedMethod),
			AllowedOrigins: common.StringsToTset(rule.AllowedOrigin),
			AllowedHeaders: types.SetNull(types.StringType), // これを入れないと"Value Conversion Error"が起きる
			ExposeHeaders:  types.SetNull(types.StringType), // ditto
			MaxAgeSeconds:  types.Int32Value(int32(rule.MaxAgeSeconds)),
		}
		if len(rule.AllowedHeader) > 0 {
			r.AllowedHeaders = common.StringsToTset(rule.AllowedHeader)
		}
		if len(rule.ExposeHeader) > 0 {
			r.ExposeHeaders = common.StringsToTset(rule.ExposeHeader)
		}
		if len(rule.ID) > 0 {
			r.ID = types.StringValue(rule.ID)
		}
		rules = append(rules, r)
	}
	return rules
}
