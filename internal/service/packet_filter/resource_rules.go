// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package packet_filter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type packetFilterRulesResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &packetFilterRulesResource{}
	_ resource.ResourceWithConfigure   = &packetFilterRulesResource{}
	_ resource.ResourceWithImportState = &packetFilterRulesResource{}
)

func NewPacketFilterRulesResource() resource.Resource {
	return &packetFilterRulesResource{}
}

func (r *packetFilterRulesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_packet_filter_rules"
}

func (r *packetFilterRulesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type packetFilterRulesResourceModel struct {
	ID             types.String                  `tfsdk:"id"`
	Zone           types.String                  `tfsdk:"zone"`
	PacketFilterID types.String                  `tfsdk:"packet_filter_id"`
	Expressions    []packetFilterExpressionModel `tfsdk:"expression"`
	Timeouts       timeouts.Value                `tfsdk:"timeouts"`
}

func (r *packetFilterRulesResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaResourceId("Packet Filter Rules"),
			"zone": common.SchemaResourceZone("Packet Filter Rules"),
			"packet_filter_id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the packet filter that set expressions to",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expression": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of packet filter expressions",
				Validators: []validator.List{
					listvalidator.SizeAtMost(30),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The protocol used for filtering. This must be one of [%s]", iaastypes.PacketFilterProtocolStrings),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.PacketFilterProtocolStrings...),
							},
						},
						"source_network": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "A source IP address or CIDR block used for filtering (e.g. `192.0.2.1`, `192.0.2.0/24`)",
						},
						"source_port": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "A source port number or port range used for filtering (e.g. `1024`, `1024-2048`)",
						},
						"destination_port": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "A destination port number or port range used for filtering (e.g. `1024`, `1024-2048`)",
						},
						"allow": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "The flag to allow the packet through the filter",
						},
						"description": schema.StringAttribute{
							Optional:    true,
							Description: "The description of this packet filter expression",
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Packet Filter's rules",
	}
}

func (r *packetFilterRulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *packetFilterRulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan packetFilterRulesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	callPacketFilterRulesUpdate(ctx, r, &plan, &resp.State, &resp.Diagnostics)
}

func (r *packetFilterRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state packetFilterRulesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	pf := getPacketFilter(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if pf == nil {
		return
	}

	state.updateState(pf, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *packetFilterRulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan packetFilterRulesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	callPacketFilterRulesUpdate(ctx, r, &plan, &resp.State, &resp.Diagnostics)
}

func (r *packetFilterRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state packetFilterRulesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	pfID := state.PacketFilterID.ValueString()

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	common.SakuraMutexKV.Lock(pfID)
	defer common.SakuraMutexKV.Unlock(pfID)

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf := getPacketFilter(ctx, r.client, common.SakuraCloudID(pfID), zone, &resp.State, &resp.Diagnostics)
	if pf == nil {
		return
	}

	_, err := pfOp.Update(ctx, zone, pf.ID, &iaas.PacketFilterUpdateRequest{
		Name:        pf.Name,
		Description: pf.Description,
		Expression:  []*iaas.PacketFilterExpression{}, // Set empty expressions to delete all rules
	}, pf.ExpressionHash)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to update SakuraCloud PacketFilter[%s]: %s", pfID, err))
		return
	}
}

func (model *packetFilterRulesResourceModel) updateState(pf *iaas.PacketFilter, zone string) {
	model.ID = types.StringValue(pf.ID.String())
	model.Zone = types.StringValue(zone)
	model.PacketFilterID = types.StringValue(pf.ID.String())
	model.Expressions = flattenPacketFilterExpressions(pf)
}

func callPacketFilterRulesUpdate(ctx context.Context, r *packetFilterRulesResource, plan *packetFilterRulesResourceModel, state *tfsdk.State, diags *diag.Diagnostics) {
	zone := common.GetZone(plan.Zone, r.client, diags)
	if diags.HasError() {
		return
	}
	pfID := plan.PacketFilterID.ValueString()

	common.SakuraMutexKV.Lock(pfID)
	defer common.SakuraMutexKV.Unlock(pfID)

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf, err := pfOp.Read(ctx, zone, common.SakuraCloudID(pfID))
	if err != nil {
		diags.AddError("Update: API Error", fmt.Sprintf("failed to read PacketFilter[%s]: %s", pfID, err))
		return
	}

	_, err = pfOp.Update(ctx, zone, pf.ID, &iaas.PacketFilterUpdateRequest{
		Name:        pf.Name,
		Description: pf.Description,
		Expression:  expandPacketFilterExpressions(plan.Expressions),
	}, pf.ExpressionHash)
	if err != nil {
		diags.AddError("Update: API Error", fmt.Sprintf("failed to update PacketFilter[%s]: %s", pfID, err))
		return
	}

	pf = getPacketFilter(ctx, r.client, common.SakuraCloudID(pfID), zone, state, diags)
	if pf == nil {
		return
	}

	plan.updateState(pf, zone)
	diags.Append(state.Set(ctx, &plan)...)
}
