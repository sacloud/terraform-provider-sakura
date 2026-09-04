// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/saclient-go"
	networkingsuite "github.com/sacloud/sacloud-sdk-go/api/networking-suite"
	v1 "github.com/sacloud/sacloud-sdk-go/api/networking-suite/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sctypes "github.com/sacloud/terraform-provider-sakura/internal/types"
)

type subnetResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &subnetResource{}
	_ resource.ResourceWithConfigure   = &subnetResource{}
	_ resource.ResourceWithImportState = &subnetResource{}
)

func NewSubnetResource() resource.Resource {
	return &subnetResource{}
}

func (r *subnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networking_suite_subnet"
}

func (r *subnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	r.client = apiClient
}

type subnetResourceModel struct {
	SRN              sctypes.SRN          `tfsdk:"srn"`
	Name             types.String         `tfsdk:"name"`
	Description      types.String         `tfsdk:"description"`
	SubnetGroupSRN   sctypes.SRN          `tfsdk:"subnet_group_srn"`
	IPv4AddressRange cidrtypes.IPv4Prefix `tfsdk:"ipv4_address_range"`
	Zone             types.String         `tfsdk:"zone"`
	Timeouts         timeouts.Value       `tfsdk:"timeouts"`
}

func (r *subnetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"srn":              common.SchemaResourceSRN("Networking Suite Subnet"),
			"name":             common.SchemaResourceName("Networking Suite Subnet"),
			"description":      common.SchemaResourceDescription("Networking Suite Subnet"),
			"subnet_group_srn": common.SchemaResourceSRNAttr("The Networking Suite Subnet Group's SRN associated with the Networking Suite Subnet", true),
			"ipv4_address_range": schema.StringAttribute{
				CustomType:  cidrtypes.IPv4PrefixType{},
				Required:    true,
				Description: "The IPv4 address range in CIDR notation for the subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of zone that the subnet will be created (e.g. `is1c`, `tk1a`)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Networking Suite Subnet.",
	}
}

func (r *subnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("srn"), req, resp)
}

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := createClient(zone, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetsOp(client)
	created, err := op.Create(ctx, v1.CreateSubnet{
		Name:                 plan.Name.ValueString(),
		Description:          plan.Description.ValueString(),
		IPv4AddressRangeCIDR: plan.IPv4AddressRange.ValueString(),
		Zone:                 v1.Zone{Code: zone},
		SubnetGroup:          v1.SakuraResourceNameRef{SRN: plan.SubnetGroupSRN.ValueString()},
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create subnet: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZoneFromSRN(state.SRN.ValueSRN())
	if zone == "" {
		resp.Diagnostics.AddError("Read: Attribute Error", fmt.Sprintf("failed to get zone from srn: %s", state.SRN.ValueString()))
		return
	}

	client, err := createClient(zone, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetsOp(client)
	found, err := op.Read(ctx, state.SRN.ValueSRN())
	if err != nil {
		if saclient.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read subnet[%s]: %s", state.SRN.ValueString(), err))
		return
	}

	state.updateState(found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := createClient(zone, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetsOp(client)
	updated, err := op.Update(ctx, plan.SRN.ValueSRN(), v1.UpdateSubnet{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update subnet: %s", err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := createClient(zone, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetsOp(client)
	if err := op.Delete(ctx, state.SRN.ValueSRN()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete subnet[%s]: %s", state.SRN.ValueString(), err))
	}
}

func (m *subnetResourceModel) updateState(subnet *v1.ReadSubnet) {
	m.SRN = sctypes.SRNValue(subnet.SRN)
	m.Name = types.StringValue(subnet.Name)
	m.Description = types.StringValue(subnet.Description)
	m.SubnetGroupSRN = sctypes.SRNValue(subnet.SubnetGroup.SRN)
	m.IPv4AddressRange = cidrtypes.NewIPv4PrefixValue(subnet.IPv4AddressRangeCIDR)
	m.Zone = types.StringValue(subnet.Zone.Code)
}
