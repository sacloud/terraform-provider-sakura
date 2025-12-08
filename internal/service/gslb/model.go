// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package gslb

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type gslbBaseModel struct {
	common.SakuraBaseModel
	IconID          types.String          `tfsdk:"icon_id"`
	FQDN            types.String          `tfsdk:"fqdn"`
	HealthCheck     *gslbHealthCheckModel `tfsdk:"health_check"`
	Weighted        types.Bool            `tfsdk:"weighted"`
	SorryServer     types.String          `tfsdk:"sorry_server"`
	Server          []gslbServerModel     `tfsdk:"server"`
	MonitoringSuite types.Object          `tfsdk:"monitoring_suite"`
}

type gslbHealthCheckModel struct {
	Protocol   types.String `tfsdk:"protocol"`
	DelayLoop  types.Int32  `tfsdk:"delay_loop"`
	HostHeader types.String `tfsdk:"host_header"`
	Path       types.String `tfsdk:"path"`
	Status     types.String `tfsdk:"status"`
	Port       types.Int32  `tfsdk:"port"`
}

type gslbServerModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Weight    types.Int32  `tfsdk:"weight"`
}

func (model *gslbBaseModel) updateState(data *iaas.GSLB) {
	model.UpdateBaseState(data.ID.String(), data.Name, data.Description, data.Tags)
	model.FQDN = types.StringValue(data.FQDN)
	model.SorryServer = types.StringValue(data.SorryServer)
	model.Weighted = types.BoolValue(data.Weighted.Bool())
	if data.HealthCheck != nil {
		model.HealthCheck = &gslbHealthCheckModel{
			Protocol:   types.StringValue(string(data.HealthCheck.Protocol)),
			DelayLoop:  types.Int32Value(int32(data.DelayLoop)),
			HostHeader: types.StringValue(data.HealthCheck.HostHeader),
			Path:       types.StringValue(data.HealthCheck.Path),
			Status:     types.StringValue(data.HealthCheck.ResponseCode.String()),
			Port:       types.Int32Value(int32(data.HealthCheck.Port.Int())),
		}
	}
	if len(data.DestinationServers) > 0 {
		servers := make([]gslbServerModel, 0, len(data.DestinationServers))
		for _, s := range data.DestinationServers {
			servers = append(servers, gslbServerModel{
				IPAddress: types.StringValue(s.IPAddress),
				Enabled:   types.BoolValue(s.Enabled.Bool()),
				Weight:    types.Int32Value(int32(s.Weight.Int())),
			})
		}
		model.Server = servers
	}
	model.MonitoringSuite = common.FlattenMonitoringSuiteLog(data.MonitoringSuiteLog)
	if data.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(data.IconID.String())
	}
}
