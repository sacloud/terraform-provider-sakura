// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_monitor

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type simpleMonitorBaseModel struct {
	ID                 types.String `tfsdk:"id"`
	Target             types.String `tfsdk:"target"`
	DelayLoop          types.Int32  `tfsdk:"delay_loop"`
	MaxCheckAttempts   types.Int32  `tfsdk:"max_check_attempts"`
	RetryInterval      types.Int32  `tfsdk:"retry_interval"`
	Timeout            types.Int32  `tfsdk:"timeout"`
	IconID             types.String `tfsdk:"icon_id"`
	Description        types.String `tfsdk:"description"`
	Tags               types.Set    `tfsdk:"tags"`
	NotifyEmailEnabled types.Bool   `tfsdk:"notify_email_enabled"`
	NotifyEmailHTML    types.Bool   `tfsdk:"notify_email_html"`
	NotifySlackEnabled types.Bool   `tfsdk:"notify_slack_enabled"`
	NotifySlackWebhook types.String `tfsdk:"notify_slack_webhook"`
	NotifyInterval     types.Int32  `tfsdk:"notify_interval"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	MonitoringSuite    types.Object `tfsdk:"monitoring_suite"`
}

type simpleMonitorHealthCheckModel struct {
	Protocol       types.String `tfsdk:"protocol"`
	HostHeader     types.String `tfsdk:"host_header"`
	Path           types.String `tfsdk:"path"`
	Status         types.Int32  `tfsdk:"status"`
	ContainsString types.String `tfsdk:"contains_string"`
	SNI            types.Bool   `tfsdk:"sni"`
	Username       types.String `tfsdk:"username"`
	Port           types.Int32  `tfsdk:"port"`
	QName          types.String `tfsdk:"qname"`
	ExpectedData   types.String `tfsdk:"expected_data"`
	Community      types.String `tfsdk:"community"`
	SnmpVersion    types.String `tfsdk:"snmp_version"`
	Oid            types.String `tfsdk:"oid"`
	RemainingDays  types.Int32  `tfsdk:"remaining_days"`
	Http2          types.Bool   `tfsdk:"http2"`
	Ftps           types.String `tfsdk:"ftps"`
	VerifySni      types.Bool   `tfsdk:"verify_sni"`
}

func (m *simpleMonitorBaseModel) updateState(sm *iaas.SimpleMonitor) {
	m.ID = types.StringValue(sm.ID.String())
	m.Description = types.StringValue(sm.Description)
	m.Tags = common.FlattenTags(sm.Tags)
	m.Target = types.StringValue(sm.Target)
	m.DelayLoop = types.Int32Value(int32(sm.DelayLoop))
	m.MaxCheckAttempts = types.Int32Value(int32(sm.MaxCheckAttempts))
	m.RetryInterval = types.Int32Value(int32(sm.RetryInterval))
	m.Timeout = types.Int32Value(int32(sm.Timeout))
	m.Enabled = types.BoolValue(sm.Enabled.Bool())
	m.NotifyEmailEnabled = types.BoolValue(sm.NotifyEmailEnabled.Bool())
	m.NotifyEmailHTML = types.BoolValue(sm.NotifyEmailHTML.Bool())
	m.NotifySlackEnabled = types.BoolValue(sm.NotifySlackEnabled.Bool())
	m.NotifySlackWebhook = types.StringValue(sm.SlackWebhooksURL)
	m.NotifyInterval = types.Int32Value(int32(flattenSimpleMonitorNotifyInterval(sm)))
	m.MonitoringSuite = common.FlattenMonitoringSuiteLog(sm.MonitoringSuiteLog)
	if sm.IconID.IsEmpty() {
		m.IconID = types.StringNull()
	} else {
		m.IconID = types.StringValue(sm.IconID.String())
	}
}

func flattenSimpleMonitorNotifyInterval(simpleMonitor *iaas.SimpleMonitor) int {
	interval := simpleMonitor.NotifyInterval
	if interval == 0 {
		return 0
	}
	// seconds => hours
	return interval / 60 / 60
}

func (m *simpleMonitorHealthCheckModel) updateState(simpleMonitor *iaas.SimpleMonitor) {
	hc := simpleMonitor.HealthCheck
	switch hc.Protocol {
	case iaastypes.SimpleMonitorProtocols.HTTP:
		m.Path = types.StringValue(hc.Path)
		m.Status = types.Int32Value(int32(hc.Status.Int()))
		m.Port = types.Int32Value(int32(hc.Port.Int()))
		if len(hc.ContainsString) > 0 {
			m.ContainsString = types.StringValue(hc.ContainsString)
		} else {
			m.ContainsString = types.StringNull()
		}
		if len(hc.Host) > 0 {
			m.HostHeader = types.StringValue(hc.Host)
		} else {
			m.HostHeader = types.StringNull()
		}
		if len(hc.BasicAuthUsername) > 0 {
			m.Username = types.StringValue(hc.BasicAuthUsername)
		} else {
			m.Username = types.StringNull()
		}
	case iaastypes.SimpleMonitorProtocols.HTTPS:
		m.Path = types.StringValue(hc.Path)
		m.Status = types.Int32Value(int32(hc.Status.Int()))
		m.Port = types.Int32Value(int32(hc.Port.Int()))
		m.SNI = types.BoolValue(hc.SNI.Bool())
		if len(hc.ContainsString) > 0 {
			m.ContainsString = types.StringValue(hc.ContainsString)
		} else {
			m.ContainsString = types.StringNull()
		}
		if len(hc.Host) > 0 {
			m.HostHeader = types.StringValue(hc.Host)
		} else {
			m.HostHeader = types.StringNull()
		}
		if len(hc.BasicAuthUsername) > 0 {
			m.Username = types.StringValue(hc.BasicAuthUsername)
		} else {
			m.Username = types.StringNull()
		}
		m.Http2 = types.BoolValue(hc.HTTP2.Bool())
	case iaastypes.SimpleMonitorProtocols.TCP, iaastypes.SimpleMonitorProtocols.SSH, iaastypes.SimpleMonitorProtocols.SMTP, iaastypes.SimpleMonitorProtocols.POP3:
		m.Port = types.Int32Value(int32(hc.Port.Int()))
	case iaastypes.SimpleMonitorProtocols.SNMP:
		m.Community = types.StringValue(hc.Community)
		m.SnmpVersion = types.StringValue(hc.SNMPVersion)
		m.Oid = types.StringValue(hc.OID)
		if len(hc.ExpectedData) > 0 {
			m.ExpectedData = types.StringValue(hc.ExpectedData)
		} else {
			m.ExpectedData = types.StringNull()
		}
	case iaastypes.SimpleMonitorProtocols.DNS:
		m.QName = types.StringValue(hc.QName)
		if len(hc.ExpectedData) > 0 {
			m.ExpectedData = types.StringValue(hc.ExpectedData)
		} else {
			m.ExpectedData = types.StringNull()
		}
	case iaastypes.SimpleMonitorProtocols.FTP:
		if len(hc.FTPS.String()) > 0 {
			m.Ftps = types.StringValue(hc.FTPS.String())
		} else {
			m.Ftps = types.StringNull()
		}
	case iaastypes.SimpleMonitorProtocols.SSLCertificate:
		m.VerifySni = types.BoolValue(hc.VerifySNI.Bool())
	}

	// These codes should be in each case, but hard to manage simple_monitor's complicated optional/computed code in terraform-plugin-framework.
	// Existing implementation can't avoid 'Provider produced inconsistent result after apply' error
	m.SNI = types.BoolValue(hc.SNI.Bool())
	m.VerifySni = types.BoolValue(hc.VerifySNI.Bool())
	m.Http2 = types.BoolValue(hc.HTTP2.Bool())
	days := hc.RemainingDays
	if days == 0 {
		days = 30
	}
	m.RemainingDays = types.Int32Value(int32(days))

	m.Protocol = types.StringValue(string(hc.Protocol))
	m.Port = types.Int32Value(int32(hc.Port.Int()))
}

func flattenSimpleMonitorHealthCheck(sm *iaas.SimpleMonitor) *simpleMonitorHealthCheckModel {
	res := simpleMonitorHealthCheckModel{}
	res.updateState(sm)
	return &res
}
