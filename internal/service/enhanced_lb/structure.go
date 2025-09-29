// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enhanced_lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

func expandEnhancedLBCreateRequest(model *enhancedLBResourceModel) *iaas.ProxyLBCreateRequest {
	return &iaas.ProxyLBCreateRequest{
		Plan:                 types.EProxyLBPlan(model.Plan.ValueInt64()),
		HealthCheck:          expandEnhancedLBHealthCheck(model),
		SorryServer:          expandEnhancedLBSorryServer(model),
		BindPorts:            expandEnhancedLBBindPorts(model),
		Servers:              expandEnhancedLBServers(model),
		Rules:                expandEnhancedLBRules(model),
		StickySession:        expandEnhancedLBStickySession(model),
		Gzip:                 expandEnhancedLBGzip(model),
		BackendHttpKeepAlive: expandEnhancedLBBackendHttpKeepAlive(model),
		ProxyProtocol:        expandEnhancedLBProxyProtocol(model),
		Syslog:               expandEnhancedLBSyslog(model),
		Timeout:              expandEnhancedLBTimeout(model),
		UseVIPFailover:       model.VIPFailover.ValueBool(),
		Region:               types.EProxyLBRegion(model.Region.ValueString()),
		Name:                 model.Name.ValueString(),
		Description:          model.Description.ValueString(),
		Tags:                 common.TsetToStrings(model.Tags),
		IconID:               common.ExpandSakuraCloudID(model.IconID),
		//LetsEncryptフィールドはenhanced_lb_acmeで管理するためCreate時には設定しない
	}
}

func expandEnhancedLBUpdateRequest(model, state *enhancedLBResourceModel) *iaas.ProxyLBUpdateRequest {
	return &iaas.ProxyLBUpdateRequest{
		HealthCheck:          expandEnhancedLBHealthCheck(model),
		SorryServer:          expandEnhancedLBSorryServer(model),
		BindPorts:            expandEnhancedLBBindPorts(model),
		Servers:              expandEnhancedLBServers(model),
		Rules:                expandEnhancedLBRules(model),
		LetsEncrypt:          expandEnhancedLBACMESetting(model, state),
		StickySession:        expandEnhancedLBStickySession(model),
		Gzip:                 expandEnhancedLBGzip(model),
		BackendHttpKeepAlive: expandEnhancedLBBackendHttpKeepAlive(model),
		ProxyProtocol:        expandEnhancedLBProxyProtocol(model),
		Syslog:               expandEnhancedLBSyslog(model),
		Timeout:              expandEnhancedLBTimeout(model),
		Name:                 model.Name.ValueString(),
		Description:          model.Description.ValueString(),
		Tags:                 common.TsetToStrings(model.Tags),
		IconID:               common.ExpandSakuraCloudID(model.IconID),
	}
}

func expandEnhancedLBStickySession(model *enhancedLBResourceModel) *iaas.ProxyLBStickySession {
	if model.StickySession.ValueBool() {
		return &iaas.ProxyLBStickySession{
			Enabled: true,
			Method:  "cookie",
		}
	}
	return nil
}

func expandEnhancedLBGzip(model *enhancedLBResourceModel) *iaas.ProxyLBGzip {
	if model.Gzip.ValueBool() {
		return &iaas.ProxyLBGzip{
			Enabled: true,
		}
	}
	return nil
}

func expandEnhancedLBBackendHttpKeepAlive(model *enhancedLBResourceModel) *iaas.ProxyLBBackendHttpKeepAlive {
	s := model.BackendHttpKeepAlive.ValueString()
	if s == "" {
		s = types.ProxyLBBackendHttpKeepAlive.Safe.String()
	}

	return &iaas.ProxyLBBackendHttpKeepAlive{
		Mode: types.EProxyLBBackendHttpKeepAlive(s),
	}
}

func expandEnhancedLBProxyProtocol(model *enhancedLBResourceModel) *iaas.ProxyLBProxyProtocol {
	if model.ProxyProtocol.ValueBool() {
		return &iaas.ProxyLBProxyProtocol{
			Enabled: true,
		}
	}
	return nil
}

func expandEnhancedLBSyslog(model *enhancedLBResourceModel) *iaas.ProxyLBSyslog {
	if model.Syslog != nil {
		return &iaas.ProxyLBSyslog{
			Server: model.Syslog.Server.ValueString(),
			Port:   int(model.Syslog.Port.ValueInt32()),
		}
	}
	return &iaas.ProxyLBSyslog{Port: 514}
}

func expandEnhancedLBBindPorts(model *enhancedLBResourceModel) []*iaas.ProxyLBBindPort {
	var results []*iaas.ProxyLBBindPort
	for _, bindPort := range model.BindPort {
		var headers []*iaas.ProxyLBResponseHeader
		if len(bindPort.ResponseHeader) > 0 {
			for _, rh := range bindPort.ResponseHeader {
				headers = append(headers, &iaas.ProxyLBResponseHeader{
					Header: rh.Header.ValueString(),
					Value:  rh.Value.ValueString(),
				})
			}
		}

		results = append(results, &iaas.ProxyLBBindPort{
			ProxyMode:         types.EProxyLBProxyMode(bindPort.ProxyMode.ValueString()),
			Port:              int(bindPort.Port.ValueInt32()),
			RedirectToHTTPS:   bindPort.RedirectToHTTPS.ValueBool(),
			SupportHTTP2:      bindPort.SupportHTTP2.ValueBool(),
			SSLPolicy:         bindPort.SSLPolicy.ValueString(),
			AddResponseHeader: headers,
		})
	}
	return results
}

func expandEnhancedLBHealthCheck(model *enhancedLBResourceModel) *iaas.ProxyLBHealthCheck {
	if model.HealthCheck != nil {
		protocol := model.HealthCheck.Protocol.ValueString()
		switch protocol {
		case "http":
			return &iaas.ProxyLBHealthCheck{
				Protocol:  types.ProxyLBProtocols.HTTP,
				Path:      model.HealthCheck.Path.ValueString(),
				Host:      model.HealthCheck.HostHeader.ValueString(),
				DelayLoop: int(model.HealthCheck.DelayLoop.ValueInt64()),
			}
		case "tcp":
			return &iaas.ProxyLBHealthCheck{
				Protocol:  types.ProxyLBProtocols.TCP,
				DelayLoop: int(model.HealthCheck.DelayLoop.ValueInt64()),
			}
		}
	}
	return nil
}

func expandEnhancedLBSorryServer(model *enhancedLBResourceModel) *iaas.ProxyLBSorryServer {
	if model.SorryServer != nil {
		return &iaas.ProxyLBSorryServer{
			IPAddress: model.SorryServer.IPAddress.ValueString(),
			Port:      int(model.SorryServer.Port.ValueInt32()),
		}
	}
	return nil
}

func expandEnhancedLBServers(model *enhancedLBResourceModel) []*iaas.ProxyLBServer {
	var results []*iaas.ProxyLBServer
	if len(model.Server) > 0 {
		for _, server := range model.Server {
			results = append(results, &iaas.ProxyLBServer{
				IPAddress:   server.IPAddress.ValueString(),
				Port:        int(server.Port.ValueInt32()),
				Enabled:     server.Enabled.ValueBool(),
				ServerGroup: server.Group.ValueString(),
			})
		}
	}
	return results
}

func expandEnhancedLBRules(model *enhancedLBResourceModel) []*iaas.ProxyLBRule {
	var results []*iaas.ProxyLBRule
	if len(model.Rule) > 0 {
		for _, rule := range model.Rule {
			results = append(results, &iaas.ProxyLBRule{
				Host:                         rule.Host.ValueString(),
				Path:                         rule.Path.ValueString(),
				SourceIPs:                    rule.SourceIPs.ValueString(),
				RequestHeaderName:            rule.RequestHeaderName.ValueString(),
				RequestHeaderValue:           rule.RequestHeaderValue.ValueString(),
				RequestHeaderValueIgnoreCase: rule.RequestHeaderValueIgnoreCase.ValueBool(),
				RequestHeaderValueNotMatch:   rule.RequestHeaderValueNotMatch.ValueBool(),
				ServerGroup:                  rule.Group.ValueString(),
				Action:                       types.EProxyLBRuleAction(rule.Action.ValueString()),
				RedirectLocation:             rule.RedirectLocation.ValueString(),
				RedirectStatusCode:           types.EProxyLBRedirectStatusCode(common.MustAtoI(rule.RedirectStatusCode.ValueString())),
				FixedStatusCode:              types.EProxyLBFixedStatusCode(common.MustAtoI(rule.FixedStatusCode.ValueString())),
				FixedContentType:             types.EProxyLBFixedContentType(rule.FixedContentType.ValueString()),
				FixedMessageBody:             rule.FixedMessageBody.ValueString(),
			})
		}
	}
	return results
}

func expandEnhancedLBTimeout(model *enhancedLBResourceModel) *iaas.ProxyLBTimeout {
	return &iaas.ProxyLBTimeout{InactiveSec: int(model.Timeout.ValueInt64())}
}

func expandEnhancedLBCerts(model *enhancedLBResourceModel) *iaas.ProxyLBCertificates {
	if model.Certificate.IsNull() || model.Certificate.IsUnknown() {
		return nil
	}

	var cm enhancedLBCertificateModel
	diags := model.Certificate.As(context.Background(), &cm, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	cert := &iaas.ProxyLBCertificates{
		PrimaryCert: &iaas.ProxyLBPrimaryCert{
			ServerCertificate:       cm.ServerCert.ValueString(),
			IntermediateCertificate: cm.IntermediateCert.ValueString(),
			PrivateKey:              cm.PrivateKey.ValueString(),
		},
	}

	if len(cm.AdditionalCertificate) > 0 {
		for _, c := range cm.AdditionalCertificate {
			cert.AdditionalCerts = append(cert.AdditionalCerts, &iaas.ProxyLBAdditionalCert{
				ServerCertificate:       c.ServerCert.ValueString(),
				IntermediateCertificate: c.IntermediateCert.ValueString(),
				PrivateKey:              c.PrivateKey.ValueString(),
			})
		}
	}

	return cert
}

func expandEnhancedLBACMESetting(plan, state *enhancedLBResourceModel) *iaas.ProxyLBACMESetting {
	// LetsEncryptの設定はenhanced_lb_acmeで管理するため、Update時にはstateも見て状態を維持する
	model := plan.LetsEncrypt
	if model.IsNull() || model.IsUnknown() {
		if state.LetsEncrypt.IsNull() || state.LetsEncrypt.IsUnknown() {
			return nil
		} else {
			model = state.LetsEncrypt
		}
	}

	var lm enhancedLBLetsEncryptModel
	diags := model.As(context.Background(), &lm, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil
	}

	return &iaas.ProxyLBACMESetting{
		CommonName:      lm.CommonName.ValueString(),
		Enabled:         lm.Enabled.ValueBool(),
		SubjectAltNames: common.TsetToStrings(lm.SubjectAltNames),
	}
}
