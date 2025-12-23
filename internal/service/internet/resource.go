// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package internet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	"github.com/sacloud/iaas-api-go/helper/query"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	internetBuilder "github.com/sacloud/iaas-service-go/internet/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type internetResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &internetResource{}
	_ resource.ResourceWithConfigure   = &internetResource{}
	_ resource.ResourceWithImportState = &internetResource{}
)

func NewInternetResource() resource.Resource {
	return &internetResource{}
}

func (r *internetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet"
}

func (r *internetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type internetResourceModel struct {
	internetBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *internetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceName := "Internet(switch+router)"

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			"tags":        common.SchemaResourceTags(resourceName),
			"icon_id":     common.SchemaResourceIconID(resourceName),
			"zone":        common.SchemaResourceZone(resourceName),
			"netmask": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The bit length of the subnet assigned to the %s. %s", resourceName, iaastypes.InternetNetworkMaskLengths),
				Default:     int32default.StaticInt32(28),
				Validators: []validator.Int32{
					int32validator.OneOf(common.MapTo(iaastypes.InternetNetworkMaskLengths, common.IntToInt32)...),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"band_width": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The bandwidth of the network connected to the Internet in Mbps. %s", iaastypes.InternetBandWidths),
				Default:     int32default.StaticInt32(100),
				Validators: []validator.Int32{
					int32validator.OneOf(common.MapTo(iaastypes.InternetBandWidths, common.IntToInt32)...),
				},
			},
			"enable_ipv6": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to enable IPv6",
			},
			"vswitch_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The id of the vSwitch"),
			},
			"server_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: desc.Sprintf("A set of the ID of Servers connected to the %s", resourceName),
			},
			"network_address": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The IPv4 network address assigned to the %s", resourceName),
			},
			"gateway": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The IP address of the gateway used by the %s", resourceName),
			},
			"min_ip_address": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("Minimum IP address in assigned global addresses to the %s", resourceName),
			},
			"max_ip_address": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("Maximum IP address in assigned global addresses to the %s", resourceName),
			},
			"ip_addresses": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: desc.Sprintf("A set of assigned global address to the %s", resourceName),
			},
			"ipv6_prefix": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The network prefix of assigned IPv6 addresses to the %s", resourceName),
			},
			"ipv6_prefix_len": schema.Int32Attribute{
				Computed:    true,
				Description: "The bit length of IPv6 network prefix",
			},
			"ipv6_network_address": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The IPv6 network address assigned to the %s", resourceName),
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Internet(Switch + Router).",
	}
}

func (r *internetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *internetResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, state *internetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if state == nil || plan == nil {
		return
	}

	if plan.BandWidth.ValueInt32() != state.BandWidth.ValueInt32() {
		// FrameworkではSDK v2と違いPlan/Stateの比較がされるため、既存のコードでは"Provider produced inconsistent result after apply"エラーが出る
		// PlanのIDをUnknownにして強制的に(known after apply)状態にすることで、制限を回避する。このアプローチはComputedのみの属性で有効
		resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())
	}
}

func (r *internetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan internetResourceModel
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

	builder := expandInternetBuilder(&plan, r.client)
	internet, err := builder.Build(ctx, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create SakuraCloud Internet: %s", err))
		return
	}

	if err := plan.updateState(ctx, r.client, zone, internet); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update Internet[%s] state: %s", internet.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *internetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state internetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	internet := getInternet(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if internet == nil {
		return
	}

	if err := state.updateState(ctx, r.client, zone, internet); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update Internet[%s] state: %s", internet.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *internetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state internetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	internetId := state.ID.ValueString() // ModifyPlanでIDがUnknownにされている場合があるため、StateからIDを取得する

	common.SakuraMutexKV.Lock(internetId)
	defer common.SakuraMutexKV.Unlock(internetId)

	builder := expandInternetBuilder(&plan, r.client)
	_, err := builder.Update(ctx, zone, common.SakuraCloudID(internetId))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Internet[%s]: %s", internetId, err))
		return
	}

	internet := getInternet(ctx, r.client, zone, common.SakuraCloudID(internetId), &resp.State, &resp.Diagnostics)
	if internet == nil {
		return
	}

	// NOTE: 帯域変更後はIDが変更になる
	if err := state.updateState(ctx, r.client, zone, internet); err != nil {
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update Internet[%s] state: %s", internet.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *internetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state internetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	internetId := state.ID.ValueString()

	common.SakuraMutexKV.Lock(internetId)
	defer common.SakuraMutexKV.Unlock(internetId)

	internetOp := iaas.NewInternetOp(r.client)
	internet, err := internetOp.Read(ctx, zone, common.ExpandSakuraCloudID(state.ID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to read Internet[%s]: %s", internetId, err))
		return
	}

	if err := query.WaitWhileSwitchIsReferenced(ctx, r.client, zone, internet.Switch.ID, r.client.CheckReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to wait deletion: Internet[%s] still used by others: %s", internet.ID, err))
		return
	}

	if err := cleanup.DeleteInternet(ctx, internetOp, zone, internet.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Internet[%s]: %s", internet.ID, err))
		return
	}
}

func getInternet(ctx context.Context, client *common.APIClient, zone string, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Internet {
	// @previous-idも考慮するため、internetOp.Readではなくquery.ReadRouterを利用する
	internet, err := query.ReadRouter(ctx, client, zone, id)
	if err != nil {
		if iaas.IsNoResultsError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Internet[%s]: %s", id, err))
		return nil
	}

	return internet
}

func expandInternetBuilder(model *internetResourceModel, client *common.APIClient) *internetBuilder.Builder {
	return &internetBuilder.Builder{
		Name:           model.Name.ValueString(),
		Description:    model.Description.ValueString(),
		Tags:           common.TsetToStrings(model.Tags),
		IconID:         common.ExpandSakuraCloudID(model.IconID),
		NetworkMaskLen: int(model.Netmask.ValueInt32()),
		BandWidthMbps:  int(model.BandWidth.ValueInt32()),
		EnableIPv6:     model.EnableIPv6.ValueBool(),
		Client:         internetBuilder.NewAPIClient(client),
	}
}
