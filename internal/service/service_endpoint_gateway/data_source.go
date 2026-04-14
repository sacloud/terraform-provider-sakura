// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package service_endpoint_gateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	seg "github.com/sacloud/service-endpoint-gateway-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type segDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &segDataSource{}
	_ datasource.DataSourceWithConfigure = &segDataSource{}
)

func NewSEGDataSource() datasource.DataSource {
	return &segDataSource{}
}

func (d *segDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_endpoint_gateway"
}

func (d *segDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type segDataSourceModel struct {
	segBaseModel
}

func (d *segDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	const resourceName = "Service Endpoint Gateway"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaDataSourceId(resourceName),
			"zone":       common.SchemaResourceZone(resourceName),
			"vswitch_id": common.SchemaDataSourceVSwitchID(resourceName),
			"server_ip_addresses": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The IP server addresslist to connect the Service Endpoint Gateway",
				Computed:    true,
			},
			"netmask": schema.Int32Attribute{
				Description: desc.Sprintf("The bit length of the subnet to assign to the Service Endpoint Gateway. %s", desc.Range(8, 29)),
				Computed:    true,
			},
			"endpoint_setting": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The endpoint settings of the Service Endpoint Gateway",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"object_storage_endpoints": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "The list of sakura object storage endpoints to connect to the Service Endpoint Gateway",
						Computed:    true,
					},
					"monitoring_suite_endpoints": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "The list of monitoring suite endpoints to connect to the Service Endpoint Gateway",
						Computed:    true,
					},
					"container_registry_endpoints": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "The list of sakura container registry endpoints to connect to the Service Endpoint Gateway",
						Computed:    true,
					},
					"ai_engine_endpoints": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "The list of AI engine endpoints to connect to the Service Endpoint Gateway",
						Computed:    true,
					},
					"apprun_dedicated_control_enabled": schema.BoolAttribute{
						Optional:    true,
						Description: "The flag to enable AppRun Dedicated Control Plane endpoint on the Service Endpoint Gateway",
						Computed:    true,
					},
				},
			},
			"monitoring_suite_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "The flag to enable monitoring suite endpoint on the Service Endpoint Gateway",
				Computed:    true,
			},
			"dns_forwarding": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The DNS forwarding settings of the Service Endpoint Gateway",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "The flag to enable DNS forwarding on the Service Endpoint Gateway",
						Computed:    true,
					},
					"private_hosted_zone": schema.StringAttribute{
						Description: "The private hosted zone name for DNS forwarding",
						Computed:    true,
					},
					"upstream_dns_1": schema.StringAttribute{
						Description: "The IP address of the first upstream DNS server for DNS forwarding",
						Computed:    true,
					},
					"upstream_dns_2": schema.StringAttribute{
						Description: "The IP address of the second upstream DNS server for DNS forwarding",
						Computed:    true,
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing seg.",
	}
}

func (d *segDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data segDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := getServiceEndpointGatewayAPIClient(d.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Client Error", fmt.Sprintf("failed to create API client for Service Endpoint Gateway in zone %s: %s", data.Zone.ValueString(), err))
		return
	}

	segOp := seg.NewServiceEndpointGatewayOp(apiClient)
	res, err := segOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read seg[%s]: %s", data.ID.ValueString(), err))
		return
	}

	err = data.updateState(&res.Appliance, zone)
	if err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update state for seg resource: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
