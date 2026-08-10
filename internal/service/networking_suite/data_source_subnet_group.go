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
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	networkingsuite "github.com/sacloud/terraform-provider-sakura/internal/ns-sdk"
	v1 "github.com/sacloud/terraform-provider-sakura/internal/ns-sdk/apis/v1"
	sctypes "github.com/sacloud/terraform-provider-sakura/internal/types"
)

type subnetGroupDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &subnetGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &subnetGroupDataSource{}
)

func NewSubnetGroupDataSource() datasource.DataSource {
	return &subnetGroupDataSource{}
}

func (r *subnetGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networking_suite_subnet_group"
}

func (r *subnetGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	r.client = apiClient
}

type subnetGroupDataSourceModel struct {
	SRN                  sctypes.SRN          `tfsdk:"srn"`
	Name                 types.String         `tfsdk:"name"`
	Description          types.String         `tfsdk:"description"`
	IPv4AddressRangeCIDR cidrtypes.IPv4Prefix `tfsdk:"ipv4_address_range_cidr"`
	Region               types.String         `tfsdk:"region"`
}

func (r *subnetGroupDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"srn":         common.SchemaDataSourceSRN("Networking Suite Subnet Group"),
			"name":        common.SchemaDataSourceName("Networking Suite Subnet Group"),
			"description": common.SchemaDataSourceDescription("Networking Suite Subnet Group"),
			"ipv4_address_range_cidr": schema.StringAttribute{
				CustomType:  cidrtypes.IPv4PrefixType{},
				Computed:    true,
				Description: "The IPv4 address range in CIDR format",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "The target region code of the subnet group",
			},
		},
		MarkdownDescription: "Get information of a Networking Suite Subnet Group.",
	}
}

func (r *subnetGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subnetGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := networkingsuite.NewClient(r.client.SaClient2)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetGroupsOp(client)
	var found *v1.ReadSubnetGroup
	if utils.IsKnown(data.SRN) {
		found, err = op.Read(ctx, data.SRN.ValueSRN())
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read subnet group[%s]: %s", data.SRN.ValueString(), err))
			return
		}
	} else {
		list, err := op.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list subnet groups: %s", err))
			return
		}
		found, err = filterSubnetGroupByName(list, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	}

	data.updateState(found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *subnetGroupDataSourceModel) updateState(group *v1.ReadSubnetGroup) {
	m.SRN = sctypes.SRNValue(group.SRN)
	m.Name = types.StringValue(group.Name)
	m.Description = types.StringValue(group.Description)
	m.IPv4AddressRangeCIDR = cidrtypes.NewIPv4PrefixValue(group.IPv4AddressRangeCIDR)
	m.Region = types.StringValue(group.Region.Code)
}

func filterSubnetGroupByName(groups []v1.ReadSubnetGroup, name string) (*v1.ReadSubnetGroup, error) {
	match := slices.Collect(func(yield func(v1.ReadSubnetGroup) bool) {
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
		return nil, fmt.Errorf("multiple subnet groups found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
