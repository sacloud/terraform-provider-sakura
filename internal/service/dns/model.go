// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dns

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

const defaultTTL = 3600

type dnsBaseModel struct {
	ID              types.String `tfsdk:"id"`
	Description     types.String `tfsdk:"description"`
	IconID          types.String `tfsdk:"icon_id"`
	Tags            types.Set    `tfsdk:"tags"`
	Zone            types.String `tfsdk:"zone"`
	DNSServers      types.List   `tfsdk:"dns_servers"`
	MonitoringSuite types.Object `tfsdk:"monitoring_suite"`
}

type dnsRecordModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int32  `tfsdk:"priority"`
	Weight   types.Int32  `tfsdk:"weight"`
	Port     types.Int32  `tfsdk:"port"`
}

func (m dnsRecordModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.StringType,
		"type":     types.StringType,
		"value":    types.StringType,
		"ttl":      types.Int64Type,
		"priority": types.Int32Type,
		"weight":   types.Int32Type,
		"port":     types.Int32Type,
	}
}

func (model *dnsBaseModel) updateState(dns *iaas.DNS) {
	model.ID = types.StringValue(dns.ID.String())
	model.Description = types.StringValue(dns.Description)
	model.Tags = common.StringsToTset(dns.Tags)
	model.Zone = types.StringValue(dns.DNSZone)
	model.DNSServers = common.StringsToTlist(dns.DNSNameServers)
	model.MonitoringSuite = common.FlattenMonitoringSuiteLog(dns.MonitoringSuiteLog)
	if dns.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(dns.IconID.String())
	}
}

func flattenDNSRecords(dns *iaas.DNS) []dnsRecordModel {
	var records []dnsRecordModel
	for _, record := range dns.Records {
		records = append(records, flattenDNSRecord(record))
	}
	return records
}

func flattenDNSRecord(record *iaas.DNSRecord) dnsRecordModel {
	r := dnsRecordModel{
		Name:  types.StringValue(record.Name),
		Type:  types.StringValue(string(record.Type)),
		Value: types.StringValue(record.RData),
		TTL:   types.Int64Value(int64(record.TTL)),
	}

	switch record.Type {
	case "MX":
		// ex. record.RData = "10 example.com."
		values := strings.SplitN(record.RData, " ", 2)
		r.Value = types.StringValue(values[1])
		r.Priority = types.Int32Value(int32(common.MustAtoI(values[0])))
	case "SRV":
		values := strings.SplitN(record.RData, " ", 4)
		r.Value = types.StringValue(values[3])
		r.Priority = types.Int32Value(int32(common.MustAtoI(values[0])))
		r.Weight = types.Int32Value(int32(common.MustAtoI(values[1])))
		r.Port = types.Int32Value(int32(common.MustAtoI(values[2])))
	}

	return r
}
