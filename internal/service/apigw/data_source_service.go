// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type apigwServiceDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &apigwServiceDataSource{}
	_ datasource.DataSourceWithConfigure = &apigwServiceDataSource{}
)

func NewApigwServiceDataSource() datasource.DataSource {
	return &apigwServiceDataSource{}
}

func (r *apigwServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_service"
}

func (r *apigwServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwServiceDataSourceModel struct {
	apigwServiceBaseModel
	ObjectStorageConfig *apigwServiceObjectStorageModel `tfsdk:"object_storage_config"`
}

func (r *apigwServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaDataSourceId("API Gateway Service"),
			"name":       common.SchemaDataSourceName("API Gateway Service"),
			"tags":       common.SchemaDataSourceComputedTags("API Gateway Service"),
			"created_at": schemaDataSourceAPIGWCreatedAt("API Gateway Service"),
			"updated_at": schemaDataSourceAPIGWUpdatedAt("API Gateway Service"),
			"subscription_id": schema.StringAttribute{
				Computed:    true,
				Description: "The subscription ID associated with the service",
			},
			"protocol": schema.StringAttribute{
				Computed:    true,
				Description: "The protocol used by the backend (http or https)",
			},
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "The host name of the backend",
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "The base path for the backend",
			},
			"port": schema.Int32Attribute{
				Computed:    true,
				Description: "The port of the backend",
			},
			"retries": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of retries for backend requests",
			},
			"connect_timeout": schema.Int32Attribute{
				Computed:    true,
				Description: "Connect timeout in milliseconds for backend requests",
			},
			"write_timeout": schema.Int32Attribute{
				Computed:    true,
				Description: "Write timeout in milliseconds for backend requests",
			},
			"read_timeout": schema.Int32Attribute{
				Computed:    true,
				Description: "Read timeout in milliseconds for backend requests",
			},
			"authentication": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("Authentication method for the backend. This can be one of %s.", serviceAuthTypes),
			},
			"route_host": schema.StringAttribute{
				Computed:    true,
				Description: "The route host for the service",
			},
			"oidc": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "OIDC authentication configuration for the service",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The entity ID of OIDC authentication",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the OIDC authentication",
					},
				},
			},
			"cors_config": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "CORS configuration for the service",
				Attributes: map[string]schema.Attribute{
					"credentials": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to allow credentials",
					},
					"access_control_exposed_headers": schema.StringAttribute{
						Computed:    true,
						Description: "Headers exposed to the client",
					},
					"access_control_allow_headers": schema.StringAttribute{
						Computed:    true,
						Description: "Allowed request headers",
					},
					"max_age": schema.Int32Attribute{
						Computed:    true,
						Description: "Max age for CORS",
					},
					"access_control_allow_methods": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "Allowed HTTP methods for CORS",
					},
					"access_control_allow_origins": schema.StringAttribute{
						Computed:    true,
						Description: "Allowed origins for CORS",
					},
					"preflight_continue": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to pass preflight result to next handler",
					},
					"private_network": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to restrict CORS to private network",
					},
				},
			},
			"object_storage_config": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Object Storage configuration used by the service",
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Computed:    true,
						Description: "The bucket name",
					},
					"folder": schema.StringAttribute{
						Computed:    true,
						Description: "The folder name within the bucket",
					},
					"endpoint": schema.StringAttribute{
						Computed:    true,
						Description: "The object storage endpoint",
					},
					"region": schema.StringAttribute{
						Computed:    true,
						Description: "The object storage region",
					},
					"use_document_index": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to use document index for object storage",
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing API Gateway Service.",
	}
}

func (d *apigwServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apigwServiceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.IsNull() && data.ID.IsNull() {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'id' or 'name' must be specified.")
		return
	}

	serviceOp := apigw.NewServiceOp(d.client)
	var id uuid.UUID
	if utils.IsKnown(data.Name) {
		services, err := serviceOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list API Gateway services: %s", err))
			return
		}
		service, err := filterAPIGWServiceByName(services, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
		id = service.ID.Value
	} else {
		id = uuid.MustParse(data.ID.ValueString())
	}

	// Listの結果にはcors_configやobject_storage_configが含まれないためReadで取得する
	service, err := serviceOp.Read(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read API Gateway service[%s]: %s", id.String(), err.Error()))
		return
	}

	data.updateState(service)
	data.ObjectStorageConfig = flattenAPIGWServiceObjectStorageConfig(service.ObjectStorageConfig)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterAPIGWServiceByName(keys []v1.ServiceDetailResponse, name string) (*v1.ServiceDetailResponse, error) {
	match := slices.Collect(func(yield func(v1.ServiceDetailResponse) bool) {
		for _, v := range keys {
			if name != string(v.Name) {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple API Gateway services found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
