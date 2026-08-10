// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	networkingsuite "github.com/sacloud/terraform-provider-sakura/internal/ns-sdk"
	v1 "github.com/sacloud/terraform-provider-sakura/internal/ns-sdk/apis/v1"
	sctypes "github.com/sacloud/terraform-provider-sakura/internal/types"
)

type subnetGroupResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &subnetGroupResource{}
	_ resource.ResourceWithConfigure   = &subnetGroupResource{}
	_ resource.ResourceWithImportState = &subnetGroupResource{}
)

func NewSubnetGroupResource() resource.Resource {
	return &subnetGroupResource{}
}

func (r *subnetGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networking_suite_subnet_group"
}

func (r *subnetGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	r.client = apiClient
}

type subnetGroupResourceModel struct {
	SRN                  sctypes.SRN          `tfsdk:"srn"`
	Name                 types.String         `tfsdk:"name"`
	Description          types.String         `tfsdk:"description"`
	IPv4AddressRangeCIDR cidrtypes.IPv4Prefix `tfsdk:"ipv4_address_range_cidr"`
	Region               types.String         `tfsdk:"region"`
	Timeouts             timeouts.Value       `tfsdk:"timeouts"`
}

func (r *subnetGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"srn":         common.SchemaResourceSRN("Networking Suite Subnet Group"),
			"name":        common.SchemaResourceName("Networking Suite Subnet Group"),
			"description": common.SchemaResourceDescription("Networking Suite Subnet Group"),
			"ipv4_address_range_cidr": schema.StringAttribute{
				CustomType:  cidrtypes.IPv4PrefixType{},
				Required:    true,
				Description: "The IPv4 address range in CIDR format",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "The target region code of the subnet group",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Networking Suite Subnet Group.",
	}
}

func (r *subnetGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("srn"), req, resp)
}

func (r *subnetGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	client, err := networkingsuite.NewClient(r.client.SaClient2)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetGroupsOp(client)
	created, err := op.Create(ctx, &v1.CreateSubnetGroup{
		Name:                 plan.Name.ValueString(),
		Description:          plan.Description.ValueString(),
		IPv4AddressRangeCIDR: plan.IPv4AddressRangeCIDR.ValueString(),
		Region:               v1.CreateRegion{Code: plan.Region.ValueString()},
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create subnet group: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := networkingsuite.NewClient(r.client.SaClient2)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetGroupsOp(client)
	found, err := op.Read(ctx, state.SRN.ValueSRN())
	if err != nil {
		if saclient.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read subnet group[%s]: %s", state.SRN.ValueString(), err))
		return
	}

	state.updateState(found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subnetGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subnetGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	client, err := networkingsuite.NewClient(r.client.SaClient2)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}
	/*
		parsed, err := parseSRN(plan.SRN.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Update: Attribute Error", fmt.Sprintf("failed to parse SRN[%s]: %s", plan.SRN.ValueString(), err))
			return
		}
	*/

	op := networkingsuite.NewSubnetGroupsOp(client)
	updated, err := op.Update(ctx, plan.SRN.ValueSRN(), &v1.UpdateSubnetGroup{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update subnet group[%s]: %s", plan.SRN.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	client, err := networkingsuite.NewClient(r.client.SaClient2)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Client Error", fmt.Sprintf("failed to create networking suite API client: %s", err))
		return
	}

	op := networkingsuite.NewSubnetGroupsOp(client)
	/*
		parsed, err := parseSRN(state.SRN.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Delete: Attribute Error", fmt.Sprintf("failed to parse SRN[%s]: %s", state.SRN.ValueString(), err))
			return
		}
	*/
	if err := op.Delete(ctx, state.SRN.ValueSRN()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete subnet group[%s]: %s", state.SRN.ValueString(), err))
	}
}

func (m *subnetGroupResourceModel) updateState(group *v1.ReadSubnetGroup) {
	m.SRN = sctypes.SRNValue(group.SRN)
	m.Name = types.StringValue(group.Name)
	m.Description = types.StringValue(group.Description)
	m.IPv4AddressRangeCIDR = cidrtypes.NewIPv4PrefixValue(group.IPv4AddressRangeCIDR)
	m.Region = types.StringValue(group.Region.Code)
}
