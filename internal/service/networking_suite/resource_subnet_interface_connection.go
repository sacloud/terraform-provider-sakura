// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	networkingsuite "github.com/sacloud/terraform-provider-sakura/internal/ns-sdk"
	sctypes "github.com/sacloud/terraform-provider-sakura/internal/types"
)

type subnetInterfaceConnectionResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &subnetInterfaceConnectionResource{}
	_ resource.ResourceWithConfigure   = &subnetInterfaceConnectionResource{}
	_ resource.ResourceWithImportState = &subnetInterfaceConnectionResource{}
)

func NewSubnetInterfaceConnectionResource() resource.Resource {
	return &subnetInterfaceConnectionResource{}
}

func (r *subnetInterfaceConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networking_suite_subnet_interface_connection"
}

func (r *subnetInterfaceConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	r.client = apiClient
}

type subnetInterfaceConnectionResourceModel struct {
	SRN                  sctypes.SRN                             `tfsdk:"srn"`
	SubnetSRN            sctypes.SRN                             `tfsdk:"subnet_srn"`
	InterfaceSRN         sctypes.SRN                             `tfsdk:"interface_srn"`
	EphemeralIPv4Address iptypes.IPv4Address                     `tfsdk:"ephemeral_ipv4_address"`
	Settings             *subnetInterfaceConnectionSettingsModel `tfsdk:"settings"`
	Timeouts             timeouts.Value                          `tfsdk:"timeouts"`
}

type subnetInterfaceConnectionSettingsModel struct {
	EnableIpForwarding types.Bool `tfsdk:"enable_ip_forwarding"`
}

func (r *subnetInterfaceConnectionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"srn":           common.SchemaResourceSRN("Networking Suite Interface Connection"),
			"subnet_srn":    common.SchemaResourceSRNAttr("The SRN of the target Subnet."),
			"interface_srn": common.SchemaResourceSRNAttr("The interface SRN to connect."),
			"ephemeral_ipv4_address": schema.StringAttribute{
				CustomType:  iptypes.IPv4AddressType{},
				Optional:    true,
				Description: "Ephemeral IPv4 address used on connect",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"settings": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Settings for the connection",
				Attributes: map[string]schema.Attribute{
					"enable_ip_forwarding": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Enable IP forwarding on the interface",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a connection between an interface and a Networking Suite Subnet.",
	}
}

func (r *subnetInterfaceConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Import: Not Supported", "networking_suite_subnet_interface_connection does not support import")
}

func (r *subnetInterfaceConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetInterfaceConnectionResourceModel
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

	op := networkingsuite.NewInterfaceConnectionOp(client)
	var ipAddress *string
	if utils.IsKnown(plan.EphemeralIPv4Address) && plan.EphemeralIPv4Address.ValueString() != "" {
		v := plan.EphemeralIPv4Address.ValueString()
		ipAddress = &v
	}
	enableIPForwarding := false
	if plan.Settings != nil {
		enableIPForwarding = plan.Settings.EnableIpForwarding.ValueBool()
	}

	created, err := op.Create(ctx, plan.InterfaceSRN.ValueSRN(), plan.SubnetSRN.ValueSRN(), ipAddress, enableIPForwarding)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to connect interface[%s] to subnet[%s]: %s", plan.InterfaceSRN.ValueString(), plan.SubnetSRN.ValueString(), err))
		return
	}

	plan.SRN = sctypes.SRNValue(created.SRN)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetInterfaceConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetInterfaceConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The current networking-suite API has no read endpoint for interface connection
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subnetInterfaceConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Supported", "networking_suite_subnet_interface_connection does not support updates")
}

func (r *subnetInterfaceConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetInterfaceConnectionResourceModel
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

	op := networkingsuite.NewInterfaceConnectionOp(client)
	if err := op.Delete(ctx, state.SRN.ValueSRN()); err != nil && !saclient.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to disconnect interface[%s] from subnet[%s]: %s", state.InterfaceSRN.ValueString(), state.SubnetSRN.ValueString(), err))
	}
}
