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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type apigwRouteDataSource struct {
	client *v1.Client
}

func NewApigwRouteDataSource() datasource.DataSource {
	return &apigwRouteDataSource{}
}

var (
	_ datasource.DataSource              = &apigwRouteDataSource{}
	_ datasource.DataSourceWithConfigure = &apigwRouteDataSource{}
)

func (r *apigwRouteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_route"
}

func (r *apigwRouteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwRouteDataSourceModel struct {
	apigwRouteBaseModel
}

func (r *apigwRouteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaDataSourceId("API Gateway Route"),
			"name":       common.SchemaDataSourceName("API Gateway Route"),
			"tags":       common.SchemaDataSourceComputedTags("API Gateway Route"),
			"created_at": schemaDataSourceAPIGWCreatedAt("API Gateway Route"),
			"updated_at": schemaDataSourceAPIGWUpdatedAt("API Gateway Route"),
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "The Service ID associated with the API Gateway Route",
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						return uuid.Validate(v)
					}),
				},
			},
			"protocols": schema.StringAttribute{
				Computed:    true,
				Description: "The protocols",
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "The request path",
			},
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "The request host",
			},
			"hosts": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The request hosts",
			},
			"methods": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The request methods",
			},
			"https_redirect_status_code": schema.Int32Attribute{
				Computed:    true,
				Description: "The HTTPS redirect status code",
			},
			"regex_priority": schema.Int32Attribute{
				Computed:    true,
				Description: "The regex priority",
			},
			"strip_path": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to strip the matching path prefix from the upstream request URL",
			},
			"preserve_host": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to preserve the original Host header",
			},
			"request_buffering": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to enable request buffering",
			},
			"response_buffering": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to enable response buffering",
			},
			"ip_restriction": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "IP restriction configuration for the route",
				Attributes: map[string]schema.Attribute{
					"protocols": schema.StringAttribute{
						Computed:    true,
						Description: "The protocols to restrict by.",
					},
					"restricted_by": schema.StringAttribute{
						Computed:    true,
						Description: "The category to restrict by.",
					},
					"ips": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "The IPv4 addresses to be restricted.",
					},
				},
			},
			"groups": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Groups associated with the user",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the API Gateway Group",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the API Gateway Group",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the group is enabled",
						},
					},
				},
			},
			"request_transformation": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Request transform configuration (headers, query params, body transformations).",
				Attributes: map[string]schema.Attribute{
					"http_method": schema.StringAttribute{
						Computed:    true,
						Description: "HTTP method (e.g. GET).",
					},
					"allow": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Allow list.",
						Attributes: map[string]schema.Attribute{
							"body": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "List of body fields to allow.",
							},
						},
					},
					"remove": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Fields to remove from the request.",
						Attributes: map[string]schema.Attribute{
							"header_keys": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "Header keys to remove.",
							},
							"query_params": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "Query parameter names to remove.",
							},
							"body": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "Body fields to remove.",
							},
						},
					},
					"rename": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Rename request fields (from -> to).",
						Attributes: map[string]schema.Attribute{
							"headers":      schemaDataSourceAPIGWListFromTo(),
							"query_params": schemaDataSourceAPIGWListFromTo(),
							"body":         schemaDataSourceAPIGWListFromTo(),
						},
					},
					"replace": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Replace values for keys.",
						Attributes: map[string]schema.Attribute{
							"headers":      schemaDataSourceAPIGWListKV(),
							"query_params": schemaDataSourceAPIGWListKV(),
							"body":         schemaDataSourceAPIGWListKV(),
						},
					},
					"add": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Add key/value pairs to request.",
						Attributes: map[string]schema.Attribute{
							"headers":      schemaDataSourceAPIGWListKV(),
							"query_params": schemaDataSourceAPIGWListKV(),
							"body":         schemaDataSourceAPIGWListKV(),
						},
					},
					"append": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Append values to existing keys.",
						Attributes: map[string]schema.Attribute{
							"headers":      schemaDataSourceAPIGWListKV(),
							"query_params": schemaDataSourceAPIGWListKV(),
							"body":         schemaDataSourceAPIGWListKV(),
						},
					},
				},
			},
			"response_transformation": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Response transform configuration (conditionals by status code, header/json/body transformations).",
				Attributes: map[string]schema.Attribute{
					"allow": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Allow list.",
						Attributes: map[string]schema.Attribute{
							"json_keys": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "List of JSON keys to allow.",
							},
						},
					},
					"remove": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Fields to remove from the response.",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaResourceAPIGWIfStatusCode(),
							"header_keys": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "Header keys to remove.",
							},
							"json_keys": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "JSON keys to remove from the response body.",
							},
						},
					},
					"rename": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Rename response fields (from -> to).",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaDataSourceAPIGWIfStatusCode(),
							"headers":        schemaDataSourceAPIGWListFromTo(),
						},
					},
					"replace": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Replace values for keys or body.",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaDataSourceAPIGWIfStatusCode(),
							"headers":        schemaDataSourceAPIGWListKV(),
							"json":           schemaDataSourceAPIGWListKV(),
							"body": schema.StringAttribute{
								Computed:    true,
								Description: "Replace whole response body with this string.",
							},
						},
					},
					"add": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Add key/value pairs to response.",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaDataSourceAPIGWIfStatusCode(),
							"headers":        schemaDataSourceAPIGWListKV(),
							"json":           schemaDataSourceAPIGWListKV(),
						},
					},
					"append": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Append values to existing keys in response.",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaDataSourceAPIGWIfStatusCode(),
							"headers":        schemaDataSourceAPIGWListKV(),
							"json":           schemaDataSourceAPIGWListKV(),
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing API Gateway Route.",
	}
}

func (d *apigwRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apigwRouteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id string
	if utils.IsKnown(data.Name) {
		routeOp := apigw.NewRouteOp(d.client, uuid.MustParse(data.ServiceID.ValueString()))
		routes, err := routeOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list API Gateway Routes: %s", err))
			return
		}
		route, err := filterAPIGWRouteByName(routes, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
		id = route.ID.Value.String()
	} else {
		id = data.ID.ValueString()
	}

	route := getAPIGWRoute(ctx, d.client, data.ServiceID.ValueString(), id, &resp.State, &resp.Diagnostics)
	if route == nil {
		return
	}
	authz, err := apigw.NewRouteExtraOp(d.client, uuid.MustParse(data.ServiceID.ValueString()), route.ID.Value).ReadAuthorization(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read API Gateway Route authorization: %s", err))
		return
	}

	data.updateState(ctx, d.client, route)
	data.Groups = flattenEnabledGroups(authz)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenEnabledGroups(authz *v1.RouteAuthorizationDetailResponse) []apigwGroupAuthModel {
	result := make([]apigwGroupAuthModel, 0, len(authz.Groups))
	for _, g := range authz.Groups {
		if g.Enabled.Value {
			model := apigwGroupAuthModel{
				ID:      types.StringValue(g.ID.Value.String()),
				Name:    types.StringValue(string(g.Name.Value)),
				Enabled: types.BoolValue(bool(g.Enabled.Value)),
			}
			result = append(result, model)
		}
	}
	return result
}

func filterAPIGWRouteByName(keys []v1.Route, name string) (*v1.Route, error) {
	match := slices.Collect(func(yield func(v1.Route) bool) {
		for _, v := range keys {
			if name != string(v.Name.Value) {
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
		return nil, fmt.Errorf("multiple API Gateway Routes found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
