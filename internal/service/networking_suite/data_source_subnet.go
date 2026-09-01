// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	networkingsuite "github.com/sacloud/sacloud-sdk-go/api/networking-suite"
	v1 "github.com/sacloud/sacloud-sdk-go/api/networking-suite/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	sctypes "github.com/sacloud/terraform-provider-sakura/internal/types"
)

type subnetDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &subnetDataSource{}
	_ datasource.DataSourceWithConfigure = &subnetDataSource{}
)

func NewSubnetDataSource() datasource.DataSource {
	return &subnetDataSource{}
}

func (r *subnetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networking_suite_subnet"
}

func (r *subnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	r.client = apiClient
}

type subnetDataSourceModel struct {
	SRN                  sctypes.SRN          `tfsdk:"srn"`
	Name                 types.String         `tfsdk:"name"`
	Description          types.String         `tfsdk:"description"`
	SubnetGroupSRN       sctypes.SRN          `tfsdk:"subnet_group_srn"`
	IPv4AddressRangeCIDR cidrtypes.IPv4Prefix `tfsdk:"ipv4_address_range_cidr"`
	Zone                 types.String         `tfsdk:"zone"`
}

func (r *subnetDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"srn":              common.SchemaDataSourceSRN("Networking Suite Subnet"),
			"name":             common.SchemaDataSourceName("Networking Suite Subnet"),
			"description":      common.SchemaDataSourceDescription("Networking Suite Subnet"),
			"subnet_group_srn": common.SchemaDataSourceSRN("Networking Suite Subnet Group associated with the Networking Suite Subnet"),
			"ipv4_address_range_cidr": schema.StringAttribute{
				CustomType:  cidrtypes.IPv4PrefixType{},
				Computed:    true,
				Description: "The IPv4 address range in CIDR format",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of zone that the subnet is in (e.g. `is1a`, `tk1a`)",
			},
		},
		MarkdownDescription: "Get information of a Networking Suite Subnet.",
	}
}

func (r *subnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := createClient(zone, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetsOp(client)
	var found *v1.ReadSubnet
	switch {
	case utils.IsKnown(data.SRN):
		found, err = op.Read(ctx, data.SRN.ValueSRN())
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read subnet[%s]: %s", data.SRN.ValueString(), err))
			return
		}
	case utils.IsKnown(data.Name):
		if !utils.IsKnown(data.SubnetGroupSRN) {
			resp.Diagnostics.AddError("Read: Attribute Error", "subnet_group_srn is required when name based search")
			return
		}
		list, err := op.List(ctx, data.SubnetGroupSRN.ValueSRN())
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list subnets: %s", err))
			return
		}
		found, err = filterSubnetByName(list, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	default:
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'srn' or 'name' must be specified")
		return
	}

	data.updateState(found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *subnetDataSourceModel) updateState(subnet *v1.ReadSubnet) {
	m.SRN = sctypes.SRNValue(subnet.SRN)
	m.Name = types.StringValue(subnet.Name)
	m.Description = types.StringValue(subnet.Description)
	m.SubnetGroupSRN = sctypes.SRNValue(subnet.SubnetGroup.SRN)
	m.IPv4AddressRangeCIDR = cidrtypes.NewIPv4PrefixValue(subnet.IPv4AddressRangeCIDR)
	m.Zone = types.StringValue(subnet.Zone.Code)
}

func filterSubnetByName(groups []v1.ReadSubnet, name string) (*v1.ReadSubnet, error) {
	match := slices.Collect(func(yield func(v1.ReadSubnet) bool) {
		for _, v := range groups {
			if name != v.Name {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple subnets found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
