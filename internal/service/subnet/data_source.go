// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package subnet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaas "github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

var (
	_ datasource.DataSource              = &subnetDataSource{}
	_ datasource.DataSourceWithConfigure = &subnetDataSource{}
)

type subnetDataSource struct {
	client *common.APIClient
}

func NewSubnetDataSource() datasource.DataSource {
	return &subnetDataSource{}
}

type subnetDataSourceModel struct {
	subnetBaseModel
	Index types.Int64 `tfsdk:"index"`
}

func (d *subnetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *subnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

func (d *subnetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaDataSourceId("Subnet"),
			"zone": common.SchemaDataSourceZone("Subnet"),
			"internet_id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the Internet(switch+router) resource that the Subnet belongs",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"index": schema.Int64Attribute{
				Required:    true,
				Description: "The index of the subnet in assigned to the Internet(switch+router)",
			},
			"vswitch_id": common.SchemaDataSourceVSwitchID("Subnet"),
			"netmask":    common.SchemaDataSourceNetMask("Subnet"),
			"next_hop": schema.StringAttribute{
				Computed:    true,
				Description: "The ip address of the next-hop at the Subnet",
			},
			"network_address": schema.StringAttribute{
				Computed:    true,
				Description: "The IPv4 network address assigned to the Subnet",
			},
			"min_ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "Minimum IP address in assigned global addresses to the Subnet",
			},
			"max_ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "Maximum IP address in assigned global addresses to the Subnet",
			},
			"ip_addresses": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A list of assigned global address to the Subnet",
			},
		},
		MarkdownDescription: "Get information about an existing Subnet.",
	}
}

func (d *subnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	internetOp := iaas.NewInternetOp(d.client)
	subnetOp := iaas.NewSubnetOp(d.client)
	internetID := common.ExpandSakuraCloudID(data.InternetID)
	subnetIndex := int(data.Index.ValueInt64())

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := internetOp.Read(ctx, zone, internetID)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", "failed to find Internet(switch+router) for Subnet: "+err.Error())
		return
	}
	if subnetIndex >= len(res.Switch.Subnets) {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find Subnet: invalid subnet index: %d", subnetIndex))
		return
	}

	subnetID := res.Switch.Subnets[subnetIndex].ID
	subnet, err := subnetOp.Read(ctx, zone, subnetID)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find Subnet[%s]: %s", subnetID, err))
		return
	}

	if err := data.updateState(zone, subnet); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update Subnet state: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
