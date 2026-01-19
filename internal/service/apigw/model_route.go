// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwRouteBaseModel struct {
	ID                      types.String                 `tfsdk:"id"`
	Name                    types.String                 `tfsdk:"name"`
	Tags                    types.Set                    `tfsdk:"tags"`
	CreatedAt               types.String                 `tfsdk:"created_at"`
	UpdatedAt               types.String                 `tfsdk:"updated_at"`
	ServiceID               types.String                 `tfsdk:"service_id"`
	Protocols               types.String                 `tfsdk:"protocols"`
	Path                    types.String                 `tfsdk:"path"`
	Host                    types.String                 `tfsdk:"host"`
	Hosts                   types.List                   `tfsdk:"hosts"`
	Methods                 types.Set                    `tfsdk:"methods"`
	HttpsRedirectStatusCode types.Int32                  `tfsdk:"https_redirect_status_code"`
	RegexPriority           types.Int32                  `tfsdk:"regex_priority"`
	StripPath               types.Bool                   `tfsdk:"strip_path"`
	PreserveHost            types.Bool                   `tfsdk:"preserve_host"`
	RequestBuffering        types.Bool                   `tfsdk:"request_buffering"`
	ResponseBuffering       types.Bool                   `tfsdk:"response_buffering"`
	IPRestriction           *apigwIpRestrictionModel     `tfsdk:"ip_restriction"`
	Groups                  []apigwGroupAuthModel        `tfsdk:"groups"`
	RequestTransformation   *apigwRequestTransformModel  `tfsdk:"request_transformation"`
	ResponseTransformation  *apigwResponseTransformModel `tfsdk:"response_transformation"`
}

type apigwGroupAuthModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (m apigwGroupAuthModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":      types.StringType,
		"name":    types.StringType,
		"enabled": types.BoolType,
	}
}

type apigwRequestTransformModel struct {
	HTTPMethod types.String                      `tfsdk:"http_method"`
	Allow      *apigwRequestTransformAllowModel  `tfsdk:"allow"`
	Remove     *apigwRequestTransformRemoveModel `tfsdk:"remove"`
	Rename     *apigwRequestTransformRenameModel `tfsdk:"rename"`
	Replace    *apigwRequestTransformKVsModel    `tfsdk:"replace"`
	Add        *apigwRequestTransformKVsModel    `tfsdk:"add"`
	Append     *apigwRequestTransformKVsModel    `tfsdk:"append"`
}

type apigwRequestTransformAllowModel struct {
	Body types.Set `tfsdk:"body"`
}

type apigwRequestTransformRemoveModel struct {
	HeaderKeys  types.Set `tfsdk:"header_keys"`
	QueryParams types.Set `tfsdk:"query_params"`
	Body        types.Set `tfsdk:"body"`
}

type apigwRequestTransformRenameModel struct {
	Headers     []apigwTransformFromToModel `tfsdk:"headers"`
	QueryParams []apigwTransformFromToModel `tfsdk:"query_params"`
	Body        []apigwTransformFromToModel `tfsdk:"body"`
}

type apigwTransformFromToModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type apigwRequestTransformKVsModel struct {
	Headers     []apigwTransformKVModel `tfsdk:"headers"`
	QueryParams []apigwTransformKVModel `tfsdk:"query_params"`
	Body        []apigwTransformKVModel `tfsdk:"body"`
}

type apigwTransformKVModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type apigwResponseTransformModel struct {
	Allow   *apigwResponseTransformAllowModel       `tfsdk:"allow"`
	Remove  *apigwResponseTransformRemoveModel      `tfsdk:"remove"`
	Rename  *apigwResponseTransformRenameModel      `tfsdk:"rename"`
	Replace *apigwResponseTransformKVsWithBodyModel `tfsdk:"replace"`
	Add     *apigwResponseTransformKVsModel         `tfsdk:"add"`
	Append  *apigwResponseTransformKVsModel         `tfsdk:"append"`
}

type apigwResponseTransformAllowModel struct {
	JSONKeys types.Set `tfsdk:"json_keys"`
}

type apigwResponseTransformRemoveModel struct {
	IfStatusCode types.Set `tfsdk:"if_status_code"`
	HeaderKeys   types.Set `tfsdk:"header_keys"`
	JSONKeys     types.Set `tfsdk:"json_keys"`
}

type apigwResponseTransformRenameModel struct {
	IfStatusCode types.Set                   `tfsdk:"if_status_code"`
	Headers      []apigwTransformFromToModel `tfsdk:"headers"`
	//JSON         []apigwTransformFromToModel `tfsdk:"json"`
}

type apigwResponseTransformKVsModel struct {
	IfStatusCode types.Set               `tfsdk:"if_status_code"`
	Headers      []apigwTransformKVModel `tfsdk:"headers"`
	JSON         []apigwTransformKVModel `tfsdk:"json"`
}

type apigwResponseTransformKVsWithBodyModel struct {
	apigwResponseTransformKVsModel
	Body types.String `tfsdk:"body"`
}

func (m *apigwRouteBaseModel) updateState(ctx context.Context, client *v1.Client, route *v1.RouteDetail) error {
	reOp := apigw.NewRouteExtraOp(client, route.ServiceId.Value, route.ID.Value)
	authz, err := reOp.ReadAuthorization(ctx)
	if err != nil {
		return err
	}
	reqTrans, err := reOp.ReadRequestTransformation(ctx)
	if err != nil {
		return err
	}
	resTrans, err := reOp.ReadResponseTransformation(ctx)
	if err != nil {
		return err
	}

	m.ID = types.StringValue(route.ID.Value.String())
	m.Name = types.StringValue(string(route.Name.Value))
	m.Tags = common.StringsToTset(route.Tags)
	m.CreatedAt = types.StringValue(route.CreatedAt.Value.String())
	m.UpdatedAt = types.StringValue(route.UpdatedAt.Value.String())
	m.ServiceID = types.StringValue(string(route.ServiceId.Value.String()))
	m.Protocols = types.StringValue(string(route.Protocols.Value))
	m.Path = types.StringValue(string(route.Path.Value))
	m.Host = types.StringValue(string(route.Host.Value))
	m.Hosts = common.StringsToTlist(route.Hosts)
	m.Methods = common.StringsToTset(common.MapTo(route.Methods, common.ToString))
	m.HttpsRedirectStatusCode = types.Int32Value(int32(route.HttpsRedirectStatusCode.Value))
	m.RegexPriority = types.Int32Value(int32(route.RegexPriority.Value))
	m.StripPath = types.BoolValue(bool(route.StripPath.Value))
	m.PreserveHost = types.BoolValue(bool(route.PreserveHost.Value))
	m.RequestBuffering = types.BoolValue(bool(route.RequestBuffering.Value))
	m.ResponseBuffering = types.BoolValue(bool(route.ResponseBuffering.Value))

	if route.IpRestrictionConfig.IsSet() {
		ipr := route.IpRestrictionConfig.Value
		m.IPRestriction = &apigwIpRestrictionModel{
			Protocols:    types.StringValue(string(ipr.Protocols)),
			RestrictedBy: types.StringValue(string(ipr.RestrictedBy)),
			Ips:          common.StringsToTset(ipr.Ips),
		}
	}

	m.Groups = flattenGroups(m.Groups, authz)
	m.RequestTransformation = flattenAPIGWRequestTransformation(reqTrans)
	m.ResponseTransformation = flattenAPIGWResponseTransformation(resTrans)

	return nil
}

func flattenGroups(groupModels []apigwGroupAuthModel, authz *v1.RouteAuthorizationDetailResponse) []apigwGroupAuthModel {
	var groups []apigwGroupAuthModel
	for _, gm := range groupModels {
		for _, g := range authz.Groups {
			if gm.ID.ValueString() == g.ID.Value.String() || gm.Name.ValueString() == string(g.Name.Value) {
				group := apigwGroupAuthModel{
					ID:      types.StringValue(g.ID.Value.String()),
					Name:    types.StringValue(string(g.Name.Value)),
					Enabled: types.BoolValue(g.Enabled.Value),
				}
				groups = append(groups, group)
				break
			}
		}
	}
	return groups
}

func flattenAPIGWRequestTransformation(rt *v1.RequestTransformation) *apigwRequestTransformModel {
	if rt == nil {
		return nil
	}

	model := &apigwRequestTransformModel{
		HTTPMethod: types.StringValue(string(rt.HttpMethod.Value)),
	}

	if rt.Allow.IsSet() {
		if len(rt.Allow.Value.Body) > 0 {
			model.Allow = &apigwRequestTransformAllowModel{
				Body: common.StringsToTset(common.MapTo(rt.Allow.Value.Body, common.ToString)),
			}
		}
	}
	if rt.Remove.IsSet() {
		remove := rt.Remove.Value
		// APIGWのAPIが現状削除された項目をnullではなく{}で返してくるため、厳密に判定する
		if !(len(remove.HeaderKeys) == 0 && len(remove.QueryParams) == 0 && len(remove.Body) == 0) {
			model.Remove = &apigwRequestTransformRemoveModel{
				HeaderKeys:  common.StringsToTset(common.MapTo(remove.HeaderKeys, common.ToString)),
				QueryParams: common.StringsToTset(common.MapTo(remove.QueryParams, common.ToString)),
				Body:        common.StringsToTset(common.MapTo(remove.Body, common.ToString)),
			}
		}
	}
	if rt.Rename.IsSet() {
		rename := rt.Rename.Value
		if !(len(rename.Headers) == 0 && len(rename.QueryParams) == 0 && len(rename.Body) == 0) {
			var headers []apigwTransformFromToModel
			for _, h := range rename.Headers {
				headers = append(headers, apigwTransformFromToModel{
					From: types.StringValue(string(h.From.Value)),
					To:   types.StringValue(string(h.To.Value)),
				})
			}
			var queryParams []apigwTransformFromToModel
			for _, q := range rename.QueryParams {
				queryParams = append(queryParams, apigwTransformFromToModel{
					From: types.StringValue(string(q.From.Value)),
					To:   types.StringValue(string(q.To.Value)),
				})
			}
			var body []apigwTransformFromToModel
			for _, b := range rename.Body {
				body = append(body, apigwTransformFromToModel{
					From: types.StringValue(string(b.From.Value)),
					To:   types.StringValue(string(b.To.Value)),
				})
			}
			model.Rename = &apigwRequestTransformRenameModel{
				Headers:     headers,
				QueryParams: queryParams,
				Body:        body,
			}
		}
	}
	if rt.Replace.IsSet() {
		replace := rt.Replace.Value
		model.Replace = flattenAPIGWRequestTransformKVs(&replace)
	}
	if rt.Add.IsSet() {
		add := rt.Add.Value
		model.Add = flattenAPIGWRequestTransformKVs(&add)
	}
	if rt.Append.IsSet() {
		append := rt.Append.Value
		model.Append = flattenAPIGWRequestTransformKVs(&append)
	}

	return model
}

func flattenAPIGWRequestTransformKVs(req *v1.RequestModificationDetail) *apigwRequestTransformKVsModel {
	if req == nil {
		return nil
	}
	if len(req.Headers) == 0 && len(req.QueryParams) == 0 && len(req.Body) == 0 {
		return nil
	}

	var headers []apigwTransformKVModel
	for _, h := range req.Headers {
		headers = append(headers, apigwTransformKVModel{
			Key:   types.StringValue(string(h.Key.Value)),
			Value: types.StringValue(string(h.Value.Value)),
		})
	}
	var queryParams []apigwTransformKVModel
	for _, q := range req.QueryParams {
		queryParams = append(queryParams, apigwTransformKVModel{
			Key:   types.StringValue(string(q.Key.Value)),
			Value: types.StringValue(string(q.Value.Value)),
		})
	}
	var body []apigwTransformKVModel
	for _, b := range req.Body {
		body = append(body, apigwTransformKVModel{
			Key:   types.StringValue(string(b.Key.Value)),
			Value: types.StringValue(string(b.Value.Value)),
		})
	}

	return &apigwRequestTransformKVsModel{
		Headers:     headers,
		QueryParams: queryParams,
		Body:        body,
	}
}

func flattenAPIGWResponseTransformation(rt *v1.ResponseTransformation) *apigwResponseTransformModel {
	if rt == nil {
		return nil
	}

	model := &apigwResponseTransformModel{}

	if rt.Allow.IsSet() {
		if len(rt.Allow.Value.JsonKeys) > 0 {
			model.Allow = &apigwResponseTransformAllowModel{
				JSONKeys: common.StringsToTset(common.MapTo(rt.Allow.Value.JsonKeys, common.ToString)),
			}
		}
	}
	if rt.Remove.IsSet() {
		remove := rt.Remove.Value
		// APIGWのAPIが現状削除された項目をnullではなく{}で返してくるため、厳密に判定する
		if !(len(remove.IfStatusCode) == 0 && len(remove.HeaderKeys) == 0 && len(remove.JsonKeys) == 0) {
			model.Remove = &apigwResponseTransformRemoveModel{
				IfStatusCode: common.IntsToTset32(remove.IfStatusCode),
				HeaderKeys:   common.StringsToTset(common.MapTo(remove.HeaderKeys, common.ToString)),
				JSONKeys:     common.StringsToTset(common.MapTo(remove.JsonKeys, common.ToString)),
			}
		}
	}
	if rt.Rename.IsSet() {
		rename := rt.Rename.Value
		if !(len(rename.IfStatusCode) == 0 && len(rename.Headers) == 0) {
			var headers []apigwTransformFromToModel
			for _, h := range rename.Headers {
				headers = append(headers, apigwTransformFromToModel{
					From: types.StringValue(string(h.From.Value)),
					To:   types.StringValue(string(h.To.Value)),
				})
			}
			model.Rename = &apigwResponseTransformRenameModel{
				IfStatusCode: common.IntsToTset32(rename.IfStatusCode),
				Headers:      headers,
			}
		}
	}
	if rt.Replace.IsSet() {
		replace := rt.Replace.Value
		if !(len(replace.IfStatusCode) == 0 && len(replace.Headers) == 0 && len(replace.JSON) == 0 && replace.Body.Value == "") {
			var headers []apigwTransformKVModel
			for _, h := range replace.Headers {
				headers = append(headers, apigwTransformKVModel{
					Key:   types.StringValue(string(h.Key.Value)),
					Value: types.StringValue(string(h.Value.Value)),
				})
			}
			var jsons []apigwTransformKVModel
			for _, j := range replace.JSON {
				jsons = append(jsons, apigwTransformKVModel{
					Key:   types.StringValue(string(j.Key.Value)),
					Value: types.StringValue(string(j.Value.Value)),
				})
			}
			model.Replace = &apigwResponseTransformKVsWithBodyModel{
				apigwResponseTransformKVsModel{
					IfStatusCode: common.IntsToTset32(replace.IfStatusCode),
					Headers:      headers,
					JSON:         jsons,
				},
				types.StringValue(string(replace.Body.Value)),
			}
		}
	}
	if rt.Add.IsSet() {
		v := rt.Add.Value
		model.Add = flattenAPIGWResponseTransformKVs(&v)
	}
	if rt.Append.IsSet() {
		v := rt.Append.Value
		model.Append = flattenAPIGWResponseTransformKVs(&v)
	}

	return model
}

func flattenAPIGWResponseTransformKVs(res *v1.ResponseModificationDetail) *apigwResponseTransformKVsModel {
	if res == nil {
		return nil
	}
	if len(res.Headers) == 0 && len(res.JSON) == 0 {
		return nil
	}

	var headers []apigwTransformKVModel
	for _, h := range res.Headers {
		headers = append(headers, apigwTransformKVModel{
			Key:   types.StringValue(string(h.Key.Value)),
			Value: types.StringValue(string(h.Value.Value)),
		})
	}
	var jsons []apigwTransformKVModel
	for _, j := range res.JSON {
		jsons = append(jsons, apigwTransformKVModel{
			Key:   types.StringValue(string(j.Key.Value)),
			Value: types.StringValue(string(j.Value.Value)),
		})
	}

	return &apigwResponseTransformKVsModel{
		IfStatusCode: common.IntsToTset32(res.IfStatusCode),
		Headers:      headers,
		JSON:         jsons,
	}
}
