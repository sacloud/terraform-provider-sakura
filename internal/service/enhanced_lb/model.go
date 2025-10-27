// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package enhanced_lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type enhancedLBBaseModel struct {
	common.SakuraBaseModel
	IconID               types.String                `tfsdk:"icon_id"`
	Plan                 types.Int64                 `tfsdk:"plan"`
	VIPFailover          types.Bool                  `tfsdk:"vip_failover"`
	StickySession        types.Bool                  `tfsdk:"sticky_session"`
	Gzip                 types.Bool                  `tfsdk:"gzip"`
	BackendHttpKeepAlive types.String                `tfsdk:"backend_http_keep_alive"`
	ProxyProtocol        types.Bool                  `tfsdk:"proxy_protocol"`
	Timeout              types.Int64                 `tfsdk:"timeout"`
	Region               types.String                `tfsdk:"region"`
	Syslog               *enhancedLBSyslogModel      `tfsdk:"syslog"`
	BindPort             []enhancedLBBindPortModel   `tfsdk:"bind_port"`
	HealthCheck          *enhancedLBHealthCheckModel `tfsdk:"health_check"`
	SorryServer          *enhancedLBSorryServerModel `tfsdk:"sorry_server"`
	Server               []enhancedLBServerModel     `tfsdk:"server"`
	Rule                 []enhancedLBRuleModel       `tfsdk:"rule"`
	Certificate          types.Object                `tfsdk:"certificate"`
	LetsEncrypt          types.Object                `tfsdk:"letsencrypt"`
	FQDN                 types.String                `tfsdk:"fqdn"`
	VIP                  types.String                `tfsdk:"vip"`
	ProxyNetworks        types.List                  `tfsdk:"proxy_networks"`
}

type enhancedLBSyslogModel struct {
	Server types.String `tfsdk:"server"`
	Port   types.Int32  `tfsdk:"port"`
}

type enhancedLBBindPortModel struct {
	ProxyMode       types.String                    `tfsdk:"proxy_mode"`
	Port            types.Int32                     `tfsdk:"port"`
	RedirectToHTTPS types.Bool                      `tfsdk:"redirect_to_https"`
	SupportHTTP2    types.Bool                      `tfsdk:"support_http2"`
	SSLPolicy       types.String                    `tfsdk:"ssl_policy"`
	ResponseHeader  []enhancedLBResponseHeaderModel `tfsdk:"response_header"`
}

type enhancedLBResponseHeaderModel struct {
	Header types.String `tfsdk:"header"`
	Value  types.String `tfsdk:"value"`
}

type enhancedLBHealthCheckModel struct {
	Protocol   types.String `tfsdk:"protocol"`
	DelayLoop  types.Int64  `tfsdk:"delay_loop"`
	HostHeader types.String `tfsdk:"host_header"`
	Path       types.String `tfsdk:"path"`
}

type enhancedLBSorryServerModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	Port      types.Int32  `tfsdk:"port"`
}

type enhancedLBServerModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	Port      types.Int32  `tfsdk:"port"`
	Group     types.String `tfsdk:"group"`
	Enabled   types.Bool   `tfsdk:"enabled"`
}

type enhancedLBRuleModel struct {
	Host                         types.String `tfsdk:"host"`
	Path                         types.String `tfsdk:"path"`
	SourceIPs                    types.String `tfsdk:"source_ips"`
	RequestHeaderName            types.String `tfsdk:"request_header_name"`
	RequestHeaderValue           types.String `tfsdk:"request_header_value"`
	RequestHeaderValueIgnoreCase types.Bool   `tfsdk:"request_header_value_ignore_case"`
	RequestHeaderValueNotMatch   types.Bool   `tfsdk:"request_header_value_not_match"`
	Group                        types.String `tfsdk:"group"`
	Action                       types.String `tfsdk:"action"`
	RedirectLocation             types.String `tfsdk:"redirect_location"`
	RedirectStatusCode           types.String `tfsdk:"redirect_status_code"`
	FixedStatusCode              types.String `tfsdk:"fixed_status_code"`
	FixedContentType             types.String `tfsdk:"fixed_content_type"`
	FixedMessageBody             types.String `tfsdk:"fixed_message_body"`
}

type enhancedLBCertificateModel struct {
	ServerCert            types.String                           `tfsdk:"server_cert"`
	IntermediateCert      types.String                           `tfsdk:"intermediate_cert"`
	PrivateKey            types.String                           `tfsdk:"private_key"`
	CommonName            types.String                           `tfsdk:"common_name"`
	SubjectAltNames       types.String                           `tfsdk:"subject_alt_names"`
	AdditionalCertificate []enhancedLBAdditionalCertificateModel `tfsdk:"additional_certificate"`
}

func (m enhancedLBCertificateModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"server_cert":            types.StringType,
		"intermediate_cert":      types.StringType,
		"private_key":            types.StringType,
		"common_name":            types.StringType,
		"subject_alt_names":      types.StringType,
		"additional_certificate": types.ListType{ElemType: types.ObjectType{AttrTypes: enhancedLBAdditionalCertificateModel{}.AttributeTypes()}},
	}
}

type enhancedLBAdditionalCertificateModel struct {
	ServerCert       types.String `tfsdk:"server_cert"`
	IntermediateCert types.String `tfsdk:"intermediate_cert"`
	PrivateKey       types.String `tfsdk:"private_key"`
}

func (m enhancedLBAdditionalCertificateModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"server_cert":       types.StringType,
		"intermediate_cert": types.StringType,
		"private_key":       types.StringType,
	}
}

type enhancedLBLetsEncryptModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	CommonName      types.String `tfsdk:"common_name"`
	SubjectAltNames types.Set    `tfsdk:"subject_alt_names"`
}

func (m enhancedLBLetsEncryptModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":           types.BoolType,
		"common_name":       types.StringType,
		"subject_alt_names": types.SetType{ElemType: types.StringType},
	}
}

func (model *enhancedLBBaseModel) updateState(ctx context.Context, client *common.APIClient, data *iaas.ProxyLB) error {
	elbOp := iaas.NewProxyLBOp(client)
	certs, err := elbOp.GetCertificates(ctx, data.ID)
	if err != nil {
		// even if certificate is deleted, it will not result in an error
		return err
	}
	health, err := elbOp.HealthStatus(ctx, data.ID)
	if err != nil {
		return err
	}

	model.UpdateBaseState(data.ID.String(), data.Name, data.Description, data.Tags)

	model.Plan = types.Int64Value(int64(data.Plan.Int()))
	model.VIPFailover = types.BoolValue(data.UseVIPFailover)
	model.StickySession = types.BoolValue(flattenEnhancedLBStickySession(data))
	model.Gzip = types.BoolValue(flattenEnhancedLBGzip(data))
	model.BackendHttpKeepAlive = types.StringValue(flattenEnhancedLBBackendHttpKeepAlive(data))
	model.ProxyProtocol = types.BoolValue(flattenEnhancedLBProxyProtocol(data))
	model.Timeout = types.Int64Value(int64(flattenEnhancedLBTimeout(data)))
	model.Region = types.StringValue(data.Region.String())
	model.FQDN = types.StringValue(data.FQDN)
	model.VIP = types.StringValue(health.CurrentVIP)
	model.ProxyNetworks = common.StringsToTlist(data.ProxyNetworks)
	model.Syslog = flattenEnhancedLBSyslog(data)
	model.BindPort = flattenEnhancedLBBindPorts(data)
	model.HealthCheck = flattenEnhancedLBHealthCheck(data)
	model.SorryServer = flattenEnhancedLBSorryServer(data)
	model.Server = flattenEnhancedLBServers(data)
	model.Rule = flattenEnhancedLBRules(data)
	model.LetsEncrypt = flattenEnhancedLBACMESetting(data)
	model.Certificate = flattenEnhancedLBCerts(certs)

	return nil
}

func flattenEnhancedLBSyslog(elb *iaas.ProxyLB) *enhancedLBSyslogModel {
	syslog := elb.Syslog
	if syslog != nil && syslog.Port > 0 && syslog.Server != "" {
		return &enhancedLBSyslogModel{
			Server: types.StringValue(syslog.Server),
			Port:   types.Int32Value(int32(syslog.Port)),
		}
	}
	return nil
}

func flattenEnhancedLBBindPorts(elb *iaas.ProxyLB) []enhancedLBBindPortModel {
	var bindPorts []enhancedLBBindPortModel
	for _, bindPort := range elb.BindPorts {
		var headers []enhancedLBResponseHeaderModel
		for _, header := range bindPort.AddResponseHeader {
			headers = append(headers, enhancedLBResponseHeaderModel{
				Header: types.StringValue(header.Header),
				Value:  types.StringValue(header.Value),
			})
		}

		bindPorts = append(bindPorts, enhancedLBBindPortModel{
			ProxyMode:       types.StringValue(string(bindPort.ProxyMode)),
			Port:            types.Int32Value(int32(bindPort.Port)),
			RedirectToHTTPS: types.BoolValue(bindPort.RedirectToHTTPS),
			SupportHTTP2:    types.BoolValue(bindPort.SupportHTTP2),
			SSLPolicy:       types.StringValue(bindPort.SSLPolicy),
			ResponseHeader:  headers,
		})
	}
	return bindPorts
}

func flattenEnhancedLBHealthCheck(elb *iaas.ProxyLB) *enhancedLBHealthCheckModel {
	if elb.HealthCheck != nil {
		return &enhancedLBHealthCheckModel{
			Protocol:   types.StringValue(string(elb.HealthCheck.Protocol)),
			DelayLoop:  types.Int64Value(int64(elb.HealthCheck.DelayLoop)),
			HostHeader: types.StringValue(elb.HealthCheck.Host),
			Path:       types.StringValue(elb.HealthCheck.Path),
		}
	}
	return nil
}

func flattenEnhancedLBSorryServer(elb *iaas.ProxyLB) *enhancedLBSorryServerModel {
	if elb.SorryServer != nil && elb.SorryServer.IPAddress != "" {
		return &enhancedLBSorryServerModel{
			IPAddress: types.StringValue(elb.SorryServer.IPAddress),
			Port:      types.Int32Value(int32(elb.SorryServer.Port)),
		}
	}
	return nil
}

func flattenEnhancedLBServers(elb *iaas.ProxyLB) []enhancedLBServerModel {
	var results []enhancedLBServerModel
	for _, server := range elb.Servers {
		results = append(results, enhancedLBServerModel{
			IPAddress: types.StringValue(server.IPAddress),
			Port:      types.Int32Value(int32(server.Port)),
			Enabled:   types.BoolValue(server.Enabled),
			Group:     types.StringValue(server.ServerGroup), // TODO: Only Optional?
		})
	}
	return results
}

func flattenEnhancedLBRules(elb *iaas.ProxyLB) []enhancedLBRuleModel {
	var results []enhancedLBRuleModel
	for _, rule := range elb.Rules {
		results = append(results, enhancedLBRuleModel{
			Host:                         types.StringValue(rule.Host),
			Path:                         types.StringValue(rule.Path),
			SourceIPs:                    types.StringValue(rule.SourceIPs),
			RequestHeaderName:            types.StringValue(rule.RequestHeaderName),
			RequestHeaderValue:           types.StringValue(rule.RequestHeaderValue),
			RequestHeaderValueIgnoreCase: types.BoolValue(rule.RequestHeaderValueIgnoreCase),
			RequestHeaderValueNotMatch:   types.BoolValue(rule.RequestHeaderValueNotMatch),
			Group:                        types.StringValue(rule.ServerGroup),
			Action:                       types.StringValue(rule.Action.String()),
			RedirectLocation:             types.StringValue(rule.RedirectLocation),
			RedirectStatusCode:           types.StringValue(rule.RedirectStatusCode.String()),
			FixedStatusCode:              types.StringValue(rule.FixedStatusCode.String()),
			FixedContentType:             types.StringValue(rule.FixedContentType.String()),
			FixedMessageBody:             types.StringValue(rule.FixedMessageBody),
		})
	}
	return results
}

func flattenEnhancedLBCerts(certs *iaas.ProxyLBCertificates) types.Object {
	v := types.ObjectNull(enhancedLBCertificateModel{}.AttributeTypes())
	if certs == nil {
		return v
	}

	elbCert := enhancedLBCertificateModel{}
	if certs.PrimaryCert != nil {
		elbCert.ServerCert = types.StringValue(certs.PrimaryCert.ServerCertificate)
		elbCert.IntermediateCert = types.StringValue(certs.PrimaryCert.IntermediateCertificate)
		elbCert.PrivateKey = types.StringValue(certs.PrimaryCert.PrivateKey)
		elbCert.CommonName = types.StringValue(certs.PrimaryCert.CertificateCommonName)
		elbCert.SubjectAltNames = types.StringValue(certs.PrimaryCert.CertificateAltNames)
	}
	if len(certs.AdditionalCerts) > 0 {
		var additionalCerts []enhancedLBAdditionalCertificateModel
		for _, ac := range certs.AdditionalCerts {
			additionalCerts = append(additionalCerts, enhancedLBAdditionalCertificateModel{
				ServerCert:       types.StringValue(ac.ServerCertificate),
				IntermediateCert: types.StringValue(ac.IntermediateCertificate),
				PrivateKey:       types.StringValue(ac.PrivateKey),
			})
		}
		elbCert.AdditionalCertificate = additionalCerts
	}

	value, diags := types.ObjectValueFrom(context.Background(), elbCert.AttributeTypes(), elbCert)
	if diags.HasError() {
		return v
	}
	return value
}

func flattenEnhancedLBStickySession(elb *iaas.ProxyLB) bool {
	if elb.StickySession != nil {
		return elb.StickySession.Enabled
	}
	return false
}

func flattenEnhancedLBGzip(elb *iaas.ProxyLB) bool {
	if elb.Gzip != nil {
		return elb.Gzip.Enabled
	}
	return false
}

func flattenEnhancedLBBackendHttpKeepAlive(elb *iaas.ProxyLB) string {
	if elb.BackendHttpKeepAlive != nil {
		return elb.BackendHttpKeepAlive.Mode.String()
	}
	return ""
}

func flattenEnhancedLBProxyProtocol(elb *iaas.ProxyLB) bool {
	if elb.ProxyProtocol != nil {
		return elb.ProxyProtocol.Enabled
	}
	return false
}

func flattenEnhancedLBTimeout(elb *iaas.ProxyLB) int {
	if elb.Timeout != nil {
		return elb.Timeout.InactiveSec
	}
	return 0
}

func flattenEnhancedLBACMESetting(elb *iaas.ProxyLB) types.Object {
	v := types.ObjectNull(enhancedLBLetsEncryptModel{}.AttributeTypes())
	if elb.LetsEncrypt != nil {
		m := enhancedLBLetsEncryptModel{
			Enabled:         types.BoolValue(elb.LetsEncrypt.Enabled),
			CommonName:      types.StringValue(elb.LetsEncrypt.CommonName),
			SubjectAltNames: common.StringsToTset(elb.LetsEncrypt.SubjectAltNames),
		}
		value, diags := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		if diags.HasError() {
			return v
		}
		return value
	}
	return v
}
