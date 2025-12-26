// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package subnet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

var _ resource.Resource = &subnetResource{}

func NewSubnetResource() resource.Resource {
	return &subnetResource{}
}

type subnetResource struct {
	client *common.APIClient
}

func (r *subnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *subnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type subnetResourceModel struct {
	subnetBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *subnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaResourceId("Subnet"),
			"zone": common.SchemaResourceZone("Subnet"),
			"internet_id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the Internet(switch+router) resource that the Subnet belongs",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"next_hop": schema.StringAttribute{
				Required:    true,
				Description: "The ip address of the next-hop at the Subnet",
			},
			"netmask": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(28),
				Description: desc.Sprintf("The bit length of the subnet to assign to the Subnet. %s", desc.Range(26, 28)),
				Validators: []validator.Int32{
					int32validator.Between(26, 28),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"vswitch_id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the vSwitch connected from the Subnet",
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
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of assigned global address to the Subnet",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Subnet.",
	}
}

func (r *subnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	internetID := common.ExpandSakuraCloudID(plan.InternetID)

	common.SakuraMutexKV.Lock(internetID.String())
	defer common.SakuraMutexKV.Unlock(internetID.String())

	internetOp := iaas.NewInternetOp(r.client)
	internet, err := internetOp.Read(ctx, zone, internetID)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read Internet(switch+router)[%s] for Subnet: %s", internetID, err))
		return
	}

	created, err := internetOp.AddSubnet(ctx, zone, internet.ID, &iaas.InternetAddSubnetRequest{
		NetworkMaskLen: int(plan.Netmask.ValueInt32()),
		NextHop:        plan.NextHop.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to add Subnet to Internet[%s]: %s", internet.ID, err))
		return
	}

	subnet := getSubnet(ctx, r.client, zone, created.ID, nil, &resp.Diagnostics)
	if subnet == nil {
		return
	}

	if err := plan.updateState(zone, subnet); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update Subnet state: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	subnet := getSubnet(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if subnet == nil {
		return
	}

	if err := state.updateState(zone, subnet); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update Subnet state: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	internetID := common.ExpandSakuraCloudID(plan.InternetID)

	common.SakuraMutexKV.Lock(internetID.String())
	defer common.SakuraMutexKV.Unlock(internetID.String())

	subnet := getSubnet(ctx, r.client, zone, common.ExpandSakuraCloudID(plan.ID), &resp.State, &resp.Diagnostics)
	if subnet == nil {
		return
	}

	internetOp := iaas.NewInternetOp(r.client)
	if _, err := internetOp.UpdateSubnet(ctx, zone, internetID, subnet.ID, &iaas.InternetUpdateSubnetRequest{
		NextHop: plan.NextHop.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Subnet[%s]: %s", subnet.ID, err))
		return
	}

	subnet = getSubnet(ctx, r.client, zone, subnet.ID, &resp.State, &resp.Diagnostics)
	if subnet == nil {
		return
	}

	if err := plan.updateState(zone, subnet); err != nil {
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update Subnet state: %s", err))
		return
	}
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

	internetOp := iaas.NewInternetOp(r.client)
	internetID := common.ExpandSakuraCloudID(state.InternetID)

	common.SakuraMutexKV.Lock(internetID.String())
	defer common.SakuraMutexKV.Unlock(internetID.String())

	subnet := getSubnet(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if subnet == nil {
		return
	}

	if err := internetOp.DeleteSubnet(ctx, zone, internetID, subnet.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Subnet[%s] from Internet[%s]: %s", subnet.ID, internetID, err))
		return
	}
}

func getSubnet(ctx context.Context, client *common.APIClient, zone string, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Subnet {
	subnetOp := iaas.NewSubnetOp(client)
	subnet, err := subnetOp.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			if state != nil {
				state.RemoveResource(ctx)
			}
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Subnet[%s]: %s", id.String(), err))
		return nil
	}
	return subnet
}
