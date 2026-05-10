// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/webaccel-api-go"
)

type webAccelBaseModel struct {
	ID               types.String            `tfsdk:"id"`
	Name             types.String            `tfsdk:"name"`
	Domain           types.String            `tfsdk:"domain"`
	DomainType       types.String            `tfsdk:"domain_type"`
	Origin           types.String            `tfsdk:"origin"`
	RequestProtocol  types.String            `tfsdk:"request_protocol"`
	Subdomain        types.String            `tfsdk:"subdomain"`
	CNAMERecordValue types.String            `tfsdk:"cname_record_value"`
	TXTRecordValue   types.String            `tfsdk:"txt_record_value"`
	DefaultCacheTTL  types.Int32             `tfsdk:"default_cache_ttl"`
	VarySupport      types.Bool              `tfsdk:"vary_support"`
	NormalizeAE      types.String            `tfsdk:"normalize_ae"`
	CorsRules        []webAccelCorsRuleModel `tfsdk:"cors_rules"`
}

func (m *webAccelBaseModel) updateState(site *webaccel.Site) error {
	m.ID = types.StringValue(site.ID)
	m.Name = types.StringValue(site.Name)
	m.DomainType = types.StringValue(site.DomainType)
	m.Domain = types.StringValue(site.Domain)
	m.Subdomain = types.StringValue(site.Subdomain)
	m.Origin = types.StringValue(site.Origin)
	m.CNAMERecordValue = types.StringValue(site.Subdomain + ".")
	m.TXTRecordValue = types.StringValue(fmt.Sprintf("webaccel=%s", site.Subdomain))

	requestProtocol, err := mapWebAccelRequestProtocol(site)
	if err != nil {
		return fmt.Errorf("invalid request_protocol: %w", err)
	}
	m.RequestProtocol = types.StringValue(requestProtocol)
	m.DefaultCacheTTL = types.Int32Value(int32(site.DefaultCacheTTL))
	m.VarySupport = types.BoolValue(site.VarySupport == webaccel.VarySupportEnabled)
	normalize, err := mapWebAccelNormalizeAE(site)
	if err != nil {
		return fmt.Errorf("invalid normalize_ae: %w", err)
	}
	m.NormalizeAE = types.StringValue(normalize)

	corsRules, err := flattenWebAccelCorsRules(site.CORSRules)
	if err != nil {
		return fmt.Errorf("invalid CORS rules: %w", err)
	}
	m.CorsRules = corsRules

	return nil
}

type webAccelOriginParamModel struct {
	Type       types.String `tfsdk:"type"`
	Origin     types.String `tfsdk:"origin"`
	Protocol   types.String `tfsdk:"protocol"`
	HostHeader types.String `tfsdk:"host_header"`
	Endpoint   types.String `tfsdk:"endpoint"`
	Region     types.String `tfsdk:"region"`
	BucketName types.String `tfsdk:"bucket_name"`
	DocIndex   types.Bool   `tfsdk:"doc_index"`
}

type webAccelCorsRuleModel struct {
	AllowAll       types.Bool `tfsdk:"allow_all"`
	AllowedOrigins types.List `tfsdk:"allowed_origins"`
}

type webAccelLoggingModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	Endpoint   types.String `tfsdk:"endpoint"`
	Region     types.String `tfsdk:"region"`
	BucketName types.String `tfsdk:"bucket_name"`
}

func flattenWebAccelCorsRules(rules []*webaccel.CORSRule) ([]webAccelCorsRuleModel, error) {
	if len(rules) == 0 {
		return nil, nil
	}
	if len(rules) > 1 {
		// NOTE: ウェブアクセラレーターAPIの現仕様では、CORSRules配列の最大長は`1`。 仕様が変更された場合、サポートを追加する。
		return nil, fmt.Errorf("duplicated CORS rule is unsupported: %d", len(rules))
	}

	rule := rules[0]
	if rule.AllowsAnyOrigin && len(rule.AllowedOrigins) != 0 {
		return nil, fmt.Errorf("allow_all and allowed_origins should not be specified together")
	}
	// NOTE: resourceのRead系処理では `cors_rules` を指定しない場合には値を代入しない。
	// これにより、レスポンス内のデフォルト値を無視することができ、差分が発生することを防ぐ。
	if !rule.AllowsAnyOrigin && len(rule.AllowedOrigins) == 0 {
		return nil, nil
	}

	result := webAccelCorsRuleModel{}
	if rule.AllowsAnyOrigin {
		result.AllowAll = types.BoolValue(true)
		result.AllowedOrigins = types.ListNull(types.StringType)
	} else {
		result.AllowAll = types.BoolValue(false)
		result.AllowedOrigins = common.StringsToTlist(rule.AllowedOrigins)
	}

	return []webAccelCorsRuleModel{result}, nil
}

func mapWebAccelRequestProtocol(site *webaccel.Site) (string, error) {
	switch site.RequestProtocol {
	case webaccel.RequestProtocolsHttpAndHttps:
		return "http+https", nil
	case webaccel.RequestProtocolsHttpsOnly:
		return "https", nil
	case webaccel.RequestProtocolsRedirectToHttps:
		return "https-redirect", nil
	default:
		return "", fmt.Errorf("invalid request protocol: %s", site.RequestProtocol)
	}
}

func mapWebAccelNormalizeAE(site *webaccel.Site) (string, error) {
	if site.NormalizeAE != "" {
		switch site.NormalizeAE {
		case webaccel.NormalizeAEBrGz:
			return "br+gzip", nil
		case webaccel.NormalizeAEGz:
			return "gzip", nil
		default:
			return "", fmt.Errorf("invalid normalize_ae parameter: '%s'", site.NormalizeAE)
		}
	}
	//NOTE: APIが返却するデフォルト値は""。
	// このフィールドで "gzip" と "" が持つ効果は同一であるため、
	// "gzip" として正規化する
	return "gzip", nil
}
