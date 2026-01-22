// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type apigwRouteResource struct {
	client *v1.Client
}

func NewApigwRouteResource() resource.Resource {
	return &apigwRouteResource{}
}

var (
	_ resource.Resource                = &apigwRouteResource{}
	_ resource.ResourceWithConfigure   = &apigwRouteResource{}
	_ resource.ResourceWithImportState = &apigwRouteResource{}
)

func (r *apigwRouteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_route"
}

func (r *apigwRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwRouteResourceModel struct {
	apigwRouteBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *apigwRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("API Gateway Route"),
			"name":       common.SchemaResourceName("API Gateway Route"),
			"tags":       common.SchemaResourceTags("API Gateway Route"),
			"created_at": schemaResourceAPIGWCreatedAt("API Gateway Route"),
			"updated_at": schemaResourceAPIGWUpdatedAt("API Gateway Route"),
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "The Service ID associated with the API Gateway Route",
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(uuid.Validate),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocols": schema.StringAttribute{
				Required:    true,
				Description: "The protocols supported by the Route",
				Validators: []validator.String{
					stringvalidator.OneOf(common.MapTo(v1.RouteDetailProtocolsHTTP.AllValues(), common.ToString)...),
				},
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/"),
				Description: "The path to access the Route. '/' or '~' prefix is required",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					sacloudvalidator.StringFuncValidator(func(v string) error {
						if v[0] != '/' && v[0] != '~' {
							return fmt.Errorf("the path must start with '/' or '~'")
						}
						return nil
					}),
				},
			},
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "The auto-issued host when hosts is not specified",
			},
			"hosts": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The list of hosts. Auto-issued host or API Gateway Domain can be used.",
			},
			"methods": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(common.StringsToTset(common.MapTo(v1.HTTPMethodGET.AllValues(), common.ToString))),
				Description: "HTTP methods to access the Route",
			},
			"https_redirect_status_code": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(426),
				Description: "The HTTPS redirect status code",
			},
			"regex_priority": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(0),
				Description: "The regex priority",
				Validators: []validator.Int32{
					int32validator.Between(0, 255),
				},
			},
			"strip_path": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to strip the matching route path from the upstream request URL",
			},
			"preserve_host": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to preserve the original Host header",
			},
			"request_buffering": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to enable request buffering",
			},
			"response_buffering": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to enable response buffering",
			},
			"ip_restriction": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "IP restriction configuration for the user",
				Attributes: map[string]schema.Attribute{
					"protocols": schema.StringAttribute{
						Required:    true,
						Description: "The protocols to restrict",
					},
					"restricted_by": schema.StringAttribute{
						Required:    true,
						Description: "The category to restrict by",
					},
					"ips": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The IPv4 addresses to be restricted",
					},
				},
			},
			"groups": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Groups associated with the user",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "ID of the API Gateway Group",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the API Gateway Group",
						},
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "Whether the authorization of this group is enabled",
						},
					},
				},
			},
			"request_transformation": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Request transform configuration (headers, query params, body transformations)",
				Attributes: map[string]schema.Attribute{
					"http_method": schema.StringAttribute{
						Required:    true,
						Description: "HTTP method (e.g. GET)",
						Validators: []validator.String{
							stringvalidator.OneOf(common.MapTo(v1.HTTPMethodGET.AllValues(), common.ToString)...),
						},
					},
					"allow": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Allow list",
						Attributes: map[string]schema.Attribute{
							"body": schema.SetAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "List of body fields to allow",
							},
						},
					},
					"remove": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Fields to remove from the request",
						Attributes: map[string]schema.Attribute{
							"header_keys": schema.SetAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Header keys to remove",
							},
							"query_params": schema.SetAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Query parameter names to remove",
							},
							"body": schema.SetAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Body fields to remove",
							},
						},
					},
					"rename": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Rename request fields (from -> to)",
						Attributes: map[string]schema.Attribute{
							"headers":      schemaResourceAPIGWListFromTo(),
							"query_params": schemaResourceAPIGWListFromTo(),
							"body":         schemaResourceAPIGWListFromTo(),
						},
					},
					"replace": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Replace values for keys",
						Attributes: map[string]schema.Attribute{
							"headers":      schemaResourceAPIGWListKV(),
							"query_params": schemaResourceAPIGWListKV(),
							"body":         schemaResourceAPIGWListKV(),
						},
					},
					"add": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Add key/value pairs to request",
						Attributes: map[string]schema.Attribute{
							"headers":      schemaResourceAPIGWListKV(),
							"query_params": schemaResourceAPIGWListKV(),
							"body":         schemaResourceAPIGWListKV(),
						},
					},
					"append": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Append values to existing keys",
						Attributes: map[string]schema.Attribute{
							"headers":      schemaResourceAPIGWListKV(),
							"query_params": schemaResourceAPIGWListKV(),
							"body":         schemaResourceAPIGWListKV(),
						},
					},
				},
			},
			"response_transformation": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Response transform configuration (conditionals by status code, header/json/body transformations)",
				Attributes: map[string]schema.Attribute{
					"allow": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Allow list",
						Attributes: map[string]schema.Attribute{
							"json_keys": schema.SetAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "List of JSON keys to allow",
							},
						},
					},
					"remove": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Fields to remove from the response",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaResourceAPIGWIfStatusCode(),
							"header_keys": schema.SetAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Header keys to remove",
							},
							"json_keys": schema.SetAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "JSON keys to remove from the response body",
							},
						},
					},
					"rename": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Rename response fields (from -> to)",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaResourceAPIGWIfStatusCode(),
							"headers":        schemaResourceAPIGWListFromTo(),
						},
					},
					"replace": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Replace values for keys or body",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaResourceAPIGWIfStatusCode(),
							"headers":        schemaResourceAPIGWListKV(),
							"json":           schemaResourceAPIGWListKV(),
							"body": schema.StringAttribute{
								Optional:    true,
								Description: "Replace whole response body with this string",
							},
						},
					},
					"add": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Add key/value pairs to response",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaResourceAPIGWIfStatusCode(),
							"headers":        schemaResourceAPIGWListKV(),
							"json":           schemaResourceAPIGWListKV(),
						},
					},
					"append": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Append values to existing keys in response",
						Attributes: map[string]schema.Attribute{
							"if_status_code": schemaResourceAPIGWIfStatusCode(),
							"headers":        schemaResourceAPIGWListKV(),
							"json":           schemaResourceAPIGWListKV(),
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Get information about an existing API Gateway Route.",
	}
}

func (r *apigwRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *apigwRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apigwRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	routeOp := apigw.NewRouteOp(r.client, uuid.MustParse(plan.ServiceID.ValueString()))
	created, err := routeOp.Create(ctx, expandAPIGWRouteRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create API Gateway Route: %s", err))
		return
	}
	if err := updateRouteExtra(ctx, r.client, uuid.MustParse(plan.ServiceID.ValueString()), created.ID.Value, &plan); err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update extra settings for API Gateway Route: %s", err))
		return
	}

	route := getAPIGWRoute(ctx, r.client, plan.ServiceID.ValueString(), created.ID.Value.String(), &resp.State, &resp.Diagnostics)
	if route == nil {
		return
	}

	if err := plan.updateState(ctx, r.client, route); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update state for API Gateway Route: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apigwRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route := getAPIGWRoute(ctx, r.client, data.ServiceID.ValueString(), data.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if route == nil {
		return
	}

	if err := data.updateState(ctx, r.client, route); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update state for API Gateway Route: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apigwRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apigwRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	route := getAPIGWRoute(ctx, r.client, plan.ServiceID.ValueString(), plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if route == nil {
		return
	}

	routeOp := apigw.NewRouteOp(r.client, uuid.MustParse(plan.ServiceID.ValueString()))
	err := routeOp.Update(ctx, expandAPIGWRouteRequest(&plan), route.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update API Gateway Route %s", err))
		return
	}
	if err := updateRouteExtra(ctx, r.client, uuid.MustParse(plan.ServiceID.ValueString()), uuid.MustParse(plan.ID.ValueString()), &plan); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update extra settings for API Gateway Route: %s", err))
		return
	}

	route = getAPIGWRoute(ctx, r.client, plan.ServiceID.ValueString(), plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if route == nil {
		return
	}

	if err := plan.updateState(ctx, r.client, route); err != nil {
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update state for API Gateway Route: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apigwRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	route := getAPIGWRoute(ctx, r.client, state.ServiceID.ValueString(), state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if route == nil {
		return
	}

	routeOp := apigw.NewRouteOp(r.client, uuid.MustParse(state.ServiceID.ValueString()))
	err := routeOp.Delete(ctx, route.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete API Gateway Route[%s]: %s", route.ID.Value.String(), err))
		return
	}
}

func getAPIGWRoute(ctx context.Context, client *v1.Client, serviceId, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.RouteDetail {
	routeOp := apigw.NewRouteOp(client, uuid.MustParse(serviceId))
	route, err := routeOp.Read(ctx, uuid.MustParse(id))
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read APIGW Route[%s]: %s", id, err))
		return nil
	}

	return route
}

func toV1Type[T ~string](s string) T {
	return T(s)
}

func updateRouteExtra(ctx context.Context, client *v1.Client, serviceId, routeId uuid.UUID, plan *apigwRouteResourceModel) error {
	ueOp := apigw.NewRouteExtraOp(client, serviceId, routeId)

	groups := make([]v1.RouteAuthorization, 0, len(plan.Groups))
	if len(plan.Groups) > 0 {
		for _, g := range plan.Groups {
			group := v1.RouteAuthorization{
				ID:      v1.NewOptUUID(uuid.MustParse(g.ID.ValueString())),
				Enabled: v1.NewOptBool(g.Enabled.ValueBool()),
			}
			if utils.IsKnown(g.Name) {
				group.Name = v1.NewOptName(v1.Name(g.Name.ValueString()))
			}
			groups = append(groups, group)
		}
	}
	if len(groups) > 0 {
		err := ueOp.EnableAuthorization(ctx, groups)
		if err != nil {
			return fmt.Errorf("failed to enables authorization to route[%s] with groups: %s", routeId.String(), err)
		}
	} else {
		err := ueOp.DisableAuthorization(ctx)
		if err != nil {
			return fmt.Errorf("failed to disables authorization to route[%s]: %s", routeId.String(), err)
		}
	}

	if plan.RequestTransformation != nil {
		reqModel := plan.RequestTransformation
		req := &v1.RequestTransformation{HttpMethod: v1.NewOptHTTPMethod(v1.HTTPMethod(reqModel.HTTPMethod.ValueString()))}
		if reqModel.Allow != nil {
			req.Allow = v1.NewOptRequestAllowDetail(
				v1.RequestAllowDetail{
					Body: common.MapTo(common.TsetToStrings(reqModel.Allow.Body), toV1Type[v1.JSONKey]),
				},
			)
		}
		if reqModel.Remove != nil {
			req.Remove = v1.NewOptRequestRemoveDetail(
				v1.RequestRemoveDetail{
					HeaderKeys:  common.MapTo(common.TsetToStrings(reqModel.Remove.HeaderKeys), toV1Type[v1.RequestHeaderKey]),
					QueryParams: common.MapTo(common.TsetToStrings(reqModel.Remove.QueryParams), toV1Type[v1.QueryParamKey]),
					Body:        common.MapTo(common.TsetToStrings(reqModel.Remove.Body), toV1Type[v1.JSONKey]),
				},
			)
		}
		if reqModel.Rename != nil {
			rename := v1.RequestRenameDetail{}
			if len(reqModel.Rename.Headers) > 0 {
				headers := make([]v1.RequestRenameDetailHeadersItem, 0, len(reqModel.Rename.Headers))
				for _, h := range reqModel.Rename.Headers {
					headers = append(headers, v1.RequestRenameDetailHeadersItem{
						From: v1.NewOptRequestHeaderKey(v1.RequestHeaderKey(h.From.ValueString())),
						To:   v1.NewOptRequestHeaderKey(v1.RequestHeaderKey(h.To.ValueString())),
					})
				}
				rename.Headers = headers
			}
			if len(reqModel.Rename.QueryParams) > 0 {
				params := make([]v1.RequestRenameDetailQueryParamsItem, 0, len(reqModel.Rename.QueryParams))
				for _, p := range reqModel.Rename.QueryParams {
					params = append(params, v1.RequestRenameDetailQueryParamsItem{
						From: v1.NewOptQueryParamKey(v1.QueryParamKey(p.From.ValueString())),
						To:   v1.NewOptQueryParamKey(v1.QueryParamKey(p.To.ValueString())),
					})
				}
				rename.QueryParams = params
			}
			if len(reqModel.Rename.Body) > 0 {
				bodies := make([]v1.RequestRenameDetailBodyItem, 0, len(reqModel.Rename.Body))
				for _, b := range reqModel.Rename.Body {
					bodies = append(bodies, v1.RequestRenameDetailBodyItem{
						From: v1.NewOptJSONKey(v1.JSONKey(b.From.ValueString())),
						To:   v1.NewOptJSONKey(v1.JSONKey(b.To.ValueString())),
					})
				}
				rename.Body = bodies
			}
			req.Rename = v1.NewOptRequestRenameDetail(rename)
		}
		if reqModel.Replace != nil {
			req.Replace = v1.NewOptRequestModificationDetail(expandAPIGRequestModificationKVs(reqModel.Replace))
		}
		if reqModel.Add != nil {
			req.Add = v1.NewOptRequestModificationDetail(expandAPIGRequestModificationKVs(reqModel.Add))
		}
		if reqModel.Append != nil {
			req.Append = v1.NewOptRequestModificationDetail(expandAPIGRequestModificationKVs(reqModel.Append))
		}

		err := ueOp.UpdateRequestTransformation(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to update request transformation to route[%s]: %s", routeId.String(), err)
		}
	}

	if plan.ResponseTransformation != nil {
		resModel := plan.ResponseTransformation
		res := &v1.ResponseTransformation{}
		if resModel.Allow != nil {
			res.Allow = v1.NewOptResponseAllowDetail(
				v1.ResponseAllowDetail{
					JsonKeys: common.MapTo(common.TsetToStrings(resModel.Allow.JSONKeys), toV1Type[v1.JSONKey]),
				},
			)
		}
		if resModel.Remove != nil {
			res.Remove = v1.NewOptResponseRemoveDetail(
				v1.ResponseRemoveDetail{
					IfStatusCode: common.TsetToInts(resModel.Remove.IfStatusCode),
					HeaderKeys:   common.MapTo(common.TsetToStrings(resModel.Remove.HeaderKeys), toV1Type[v1.ResponseHeaderKey]),
					JsonKeys:     common.MapTo(common.TsetToStrings(resModel.Remove.JSONKeys), toV1Type[v1.JSONKey]),
				},
			)
		}
		if resModel.Rename != nil {
			rename := v1.ResponseRenameDetail{IfStatusCode: common.TsetToInts(resModel.Rename.IfStatusCode)}
			if len(resModel.Rename.Headers) > 0 {
				headers := make([]v1.ResponseRenameDetailHeadersItem, 0, len(resModel.Rename.Headers))
				for _, h := range resModel.Rename.Headers {
					headers = append(headers, v1.ResponseRenameDetailHeadersItem{
						From: v1.NewOptResponseHeaderKey(v1.ResponseHeaderKey(h.From.ValueString())),
						To:   v1.NewOptResponseHeaderKey(v1.ResponseHeaderKey(h.To.ValueString())),
					})
				}
				rename.Headers = headers
			}
			res.Rename = v1.NewOptResponseRenameDetail(rename)
		}
		if resModel.Replace != nil {
			replace := v1.ResponseReplaceDetail{IfStatusCode: common.TsetToInts(resModel.Replace.IfStatusCode)}
			if len(resModel.Replace.Headers) > 0 {
				headers := make([]v1.ResponseReplaceDetailHeadersItem, 0, len(resModel.Replace.Headers))
				for _, h := range resModel.Replace.Headers {
					headers = append(headers, v1.ResponseReplaceDetailHeadersItem{
						Key:   v1.NewOptResponseHeaderKey(v1.ResponseHeaderKey(h.Key.ValueString())),
						Value: v1.NewOptRequestHeaderValue(v1.RequestHeaderValue(h.Value.ValueString())),
					})
				}
				replace.Headers = headers
			}
			if len(resModel.Replace.JSON) > 0 {
				jsons := make([]v1.ResponseReplaceDetailJSONItem, 0, len(resModel.Replace.JSON))
				for _, j := range resModel.Replace.JSON {
					jsons = append(jsons, v1.ResponseReplaceDetailJSONItem{
						Key:   v1.NewOptJSONKey(v1.JSONKey(j.Key.ValueString())),
						Value: v1.NewOptString(j.Value.ValueString()),
					})
				}
				replace.JSON = jsons
			}
			if utils.IsKnown(resModel.Replace.Body) {
				replace.Body = v1.NewOptString(resModel.Replace.Body.ValueString())
			}
			res.Replace = v1.NewOptResponseReplaceDetail(replace)
		}
		if resModel.Add != nil {
			res.Add = v1.NewOptResponseModificationDetail(expandAPIGResponseModificationKVs(resModel.Add))
		}
		if resModel.Append != nil {
			res.Append = v1.NewOptResponseModificationDetail(expandAPIGResponseModificationKVs(resModel.Append))
		}

		err := ueOp.UpdateResponseTransformation(ctx, res)
		if err != nil {
			return fmt.Errorf("failed to update response transformation to route[%s]: %s", routeId.String(), err)
		}
	}

	return nil
}

func expandAPIGRequestModificationKVs(model *apigwRequestTransformKVsModel) v1.RequestModificationDetail {
	res := v1.RequestModificationDetail{}

	if len(model.Headers) > 0 {
		headers := make([]v1.RequestModificationDetailHeadersItem, 0, len(model.Headers))
		for _, h := range model.Headers {
			headers = append(headers, v1.RequestModificationDetailHeadersItem{
				Key:   v1.NewOptRequestHeaderKey(v1.RequestHeaderKey(h.Key.ValueString())),
				Value: v1.NewOptRequestHeaderValue(v1.RequestHeaderValue(h.Value.ValueString())),
			})
		}
		res.Headers = headers
	}
	if len(model.QueryParams) > 0 {
		params := make([]v1.RequestModificationDetailQueryParamsItem, 0, len(model.QueryParams))
		for _, p := range model.QueryParams {
			params = append(params, v1.RequestModificationDetailQueryParamsItem{
				Key:   v1.NewOptQueryParamKey(v1.QueryParamKey(p.Key.ValueString())),
				Value: v1.NewOptQueryParamValue(v1.QueryParamValue(p.Value.ValueString())),
			})
		}
		res.QueryParams = params
	}
	if len(model.Body) > 0 {
		bodies := make([]v1.RequestModificationDetailBodyItem, 0, len(model.Body))
		for _, b := range model.Body {
			bodies = append(bodies, v1.RequestModificationDetailBodyItem{
				Key:   v1.NewOptJSONKey(v1.JSONKey(b.Key.ValueString())),
				Value: v1.NewOptString(b.Value.ValueString()),
			})
		}
		res.Body = bodies
	}

	return res
}

func expandAPIGResponseModificationKVs(model *apigwResponseTransformKVsModel) v1.ResponseModificationDetail {
	res := v1.ResponseModificationDetail{IfStatusCode: common.TsetToInts(model.IfStatusCode)}

	if len(model.Headers) > 0 {
		headers := make([]v1.ResponseModificationDetailHeadersItem, 0, len(model.Headers))
		for _, h := range model.Headers {
			headers = append(headers, v1.ResponseModificationDetailHeadersItem{
				Key:   v1.NewOptResponseHeaderKey(v1.ResponseHeaderKey(h.Key.ValueString())),
				Value: v1.NewOptRequestHeaderValue(v1.RequestHeaderValue(h.Value.ValueString())),
			})
		}
		res.Headers = headers
	}
	if len(model.JSON) > 0 {
		json := make([]v1.ResponseModificationDetailJSONItem, 0, len(model.JSON))
		for _, p := range model.JSON {
			json = append(json, v1.ResponseModificationDetailJSONItem{
				Key:   v1.NewOptJSONKey(v1.JSONKey(p.Key.ValueString())),
				Value: v1.NewOptString(p.Value.ValueString()),
			})
		}
		res.JSON = json
	}

	return res
}

func expandAPIGWRouteRequest(plan *apigwRouteResourceModel) *v1.RouteDetail {
	res := &v1.RouteDetail{
		Name:                    v1.NewOptName(v1.Name(plan.Name.ValueString())),
		Tags:                    common.TsetToStrings(plan.Tags),
		Path:                    v1.NewOptString(plan.Path.ValueString()),
		Protocols:               v1.NewOptRouteDetailProtocols(v1.RouteDetailProtocols(plan.Protocols.ValueString())),
		Methods:                 common.MapTo(common.TsetToStrings(plan.Methods), ToHTTPMethod),
		HttpsRedirectStatusCode: v1.NewOptRouteDetailHttpsRedirectStatusCode(v1.RouteDetailHttpsRedirectStatusCode(plan.HttpsRedirectStatusCode.ValueInt32())),
		RegexPriority:           v1.NewOptInt(int(plan.RegexPriority.ValueInt32())),
		StripPath:               v1.NewOptBool(plan.StripPath.ValueBool()),
		PreserveHost:            v1.NewOptBool(plan.PreserveHost.ValueBool()),
		RequestBuffering:        v1.NewOptBool(plan.RequestBuffering.ValueBool()),
		ResponseBuffering:       v1.NewOptBool(plan.ResponseBuffering.ValueBool()),
	}
	if utils.IsKnown(plan.Hosts) {
		res.Hosts = common.TlistToStrings(plan.Hosts)
	}
	if plan.IPRestriction != nil {
		res.IpRestrictionConfig = v1.NewOptIpRestrictionConfig(
			v1.IpRestrictionConfig{
				Protocols:    v1.IpRestrictionConfigProtocols(plan.IPRestriction.Protocols.ValueString()),
				RestrictedBy: v1.IpRestrictionConfigRestrictedBy(plan.IPRestriction.RestrictedBy.ValueString()),
				Ips:          common.TsetToStrings(plan.IPRestriction.Ips),
			},
		)
	}

	return res
}
