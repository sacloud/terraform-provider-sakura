// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package nlb

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaas "github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/accessor"
	"github.com/sacloud/iaas-api-go/helper/power"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/iaas-service-go/setup"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type nlbResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &nlbResource{}
	_ resource.ResourceWithConfigure   = &nlbResource{}
	_ resource.ResourceWithImportState = &nlbResource{}
)

func NewNLBResource() resource.Resource {
	return &nlbResource{}
}

func (r *nlbResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nlb"
}

func (r *nlbResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type nlbResourceModel struct {
	nlbBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *nlbResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("NLB"),
			"name":        common.SchemaResourceName("NLB"),
			"description": common.SchemaResourceDescription("NLB"),
			"tags":        common.SchemaResourceTags("NLB"),
			"zone":        common.SchemaResourceZone("NLB"),
			"icon_id":     common.SchemaResourceIconID("NLB"),
			"plan":        common.SchemaResourcePlan("NLB", "standard", []string{"standard", "highspec"}),
			"network_interface": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Network interface for NLB",
				Attributes: map[string]schema.Attribute{
					"vswitch_id": common.SchemaResourceSwitchID("NLB"),
					"vrid": schema.Int64Attribute{
						Required:    true,
						Description: "The Virtual Router Identifier",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
					"ip_addresses": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "A list of IP address to assign to the load balancer.",
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.SizeAtMost(2),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
					"netmask": schema.Int32Attribute{
						Required:    true,
						Description: desc.Sprintf("The bit length of the subnet assigned to the load balancer. %s", desc.Range(8, 29)),
						Validators: []validator.Int32{
							int32validator.Between(8, 29),
						},
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"gateway": schema.StringAttribute{
						Optional:    true,
						Description: "The IP address of the gateway used by load balancer.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
				},
			},
			"vip": schema.ListNestedAttribute{
				Optional: true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(20),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"vip": schema.StringAttribute{
							Required:    true,
							Description: "The virtual IP address",
						},
						"port": schema.Int32Attribute{
							Required:    true,
							Description: desc.Sprintf("The target port number for load-balancing. %s", desc.Range(1, 65535)),
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
							},
						},
						"delay_loop": schema.Int32Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int32default.StaticInt32(10),
							Description: desc.Sprintf("The interval in seconds between checks. %s", desc.Range(10, 2147483647)),
							Validators: []validator.Int32{
								int32validator.Between(10, 2147483647),
							},
						},
						"sorry_server": schema.StringAttribute{
							Optional:    true,
							Description: "The IP address of the SorryServer. This will be used when all servers under this VIP are down",
						},
						"description": common.SchemaResourceDescription("LoadBalancer's VIP"),
						"server": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"ip_address": schema.StringAttribute{
										Required:    true,
										Description: "The IP address of the destination server",
									},
									"protocol": schema.StringAttribute{
										Required:    true,
										Description: desc.Sprintf("The protocol used for health checks. This must be one of [%s]", iaastypes.LoadBalancerHealthCheckProtocolStrings),
									},
									"path": schema.StringAttribute{
										Optional:    true,
										Description: "The path used when checking by HTTP/HTTPS",
									},
									"status": schema.Int32Attribute{
										Optional:    true,
										Description: "The response code to expect when checking by HTTP/HTTPS",
									},
									"enabled": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(true),
										Description: "The flag to enable as destination of load balancing",
									},
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manage a NLB",
	}
}

func (r *nlbResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *nlbResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nlbResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	lbOp := iaas.NewLoadBalancerOp(r.client)
	builder := &setup.RetryableSetup{
		Create: func(ctx context.Context, zone string) (accessor.ID, error) {
			req := expandLoadBalancerCreateRequest(&plan)
			return lbOp.Create(ctx, zone, req)
		},
		ProvisionBeforeUp: func(ctx context.Context, zone string, id iaastypes.ID, _ interface{}) error {
			return lbOp.Config(ctx, zone, id)
		},
		Read: func(ctx context.Context, zone string, id iaastypes.ID) (interface{}, error) {
			return lbOp.Read(ctx, zone, id)
		},
		Delete: func(ctx context.Context, zone string, id iaastypes.ID) error {
			return lbOp.Delete(ctx, zone, id)
		},
		IsWaitForUp:   true,
		IsWaitForCopy: true,
		Options: &setup.Options{
			RetryCount: 3,
		},
	}

	res, err := builder.Setup(ctx, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create NLB: %s", err))
		return
	}
	lb, ok := res.(*iaas.LoadBalancer)
	if !ok {
		resp.Diagnostics.AddError("Create: API Error", "created resource is not *iaas.LoadBalancer")
		return
	}
	if lb.Availability.IsFailed() {
		resp.Diagnostics.AddError("Create: State Error", fmt.Sprintf("got unexpected state: NLB[%s].Availability is failed", lb.ID.String()))
		return
	}

	plan.updateState(lb, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *nlbResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nlbResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	lb := getLoadBalancer(ctx, r.client, zone, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if lb == nil {
		return
	}

	state.updateState(lb, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *nlbResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan nlbResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	lbOp := iaas.NewLoadBalancerOp(r.client)
	found, err := lbOp.Read(ctx, zone, common.ExpandSakuraCloudID(plan.ID))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read NLB[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	if _, err := iaas.NewLoadBalancerOp(r.client).Update(ctx, zone, found.ID, expandLoadBalancerUpdateRequest(&plan, found)); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update NLB[%s]: %s", plan.ID.ValueString(), err))
		return
	}
	if err := iaas.NewLoadBalancerOp(r.client).Config(ctx, zone, found.ID); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to configure NLB[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	lb := getLoadBalancer(ctx, r.client, zone, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if lb == nil {
		return
	}

	plan.updateState(lb, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *nlbResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nlbResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	lbOp := iaas.NewLoadBalancerOp(r.client)
	found := getLoadBalancer(ctx, r.client, zone, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if found == nil {
		return
	}

	if err := power.ShutdownLoadBalancer(ctx, lbOp, zone, found.ID, true); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to shutdown NLB[%s]: %s", state.ID.ValueString(), err))
		return
	}

	if err := lbOp.Delete(ctx, zone, found.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete NLB[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getLoadBalancer(ctx context.Context, client *common.APIClient, zone string, id string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.LoadBalancer {
	lbOp := iaas.NewLoadBalancerOp(client)
	lb, err := lbOp.Read(ctx, zone, common.SakuraCloudID(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read NLB[%s]: %s", id, err))
		return nil
	}
	if lb.Availability.IsFailed() {
		state.RemoveResource(ctx)
		diags.AddError("State Error", fmt.Sprintf("got unexpected state: NLB[%s].Availability is failed", id))
		return nil
	}
	return lb
}

func expandLoadBalancerCreateRequest(model *nlbResourceModel) *iaas.LoadBalancerCreateRequest {
	nic := model.NetworkInterface
	return &iaas.LoadBalancerCreateRequest{
		SwitchID:           common.ExpandSakuraCloudID(nic.VSwitchID),
		PlanID:             expandLoadBalancerPlanID(model),
		VRID:               int(nic.VRID.ValueInt64()),
		IPAddresses:        common.TlistToStrings(nic.IPAddresses),
		NetworkMaskLen:     int(nic.Netmask.ValueInt32()),
		DefaultRoute:       nic.Gateway.ValueString(),
		Name:               model.Name.ValueString(),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		VirtualIPAddresses: expandLoadBalancerVIPs(model),
	}
}

func expandLoadBalancerUpdateRequest(model *nlbResourceModel, lb *iaas.LoadBalancer) *iaas.LoadBalancerUpdateRequest {
	return &iaas.LoadBalancerUpdateRequest{
		Name:               model.Name.ValueString(),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		VirtualIPAddresses: expandLoadBalancerVIPs(model),
		SettingsHash:       lb.SettingsHash,
	}
}

func expandLoadBalancerPlanID(model *nlbResourceModel) iaastypes.ID {
	if model.Plan.ValueString() == "standard" {
		return iaastypes.LoadBalancerPlans.Standard
	}
	return iaastypes.LoadBalancerPlans.HighSpec
}

func expandLoadBalancerVIPs(model *nlbResourceModel) []*iaas.LoadBalancerVirtualIPAddress {
	if model == nil || len(model.VIP) == 0 {
		return nil
	}

	var results []*iaas.LoadBalancerVirtualIPAddress
	for _, vip := range model.VIP {
		var servers []*iaas.LoadBalancerServer
		for _, s := range vip.Server {
			server := &iaas.LoadBalancerServer{
				IPAddress: s.IPAddress.ValueString(),
				Port:      iaastypes.StringNumber(int(vip.Port.ValueInt32())),
				Enabled:   iaastypes.StringFlag(s.Enabled.ValueBool()),
				HealthCheck: &iaas.LoadBalancerServerHealthCheck{
					Protocol:     iaastypes.ELoadBalancerHealthCheckProtocol(s.Protocol.ValueString()),
					Path:         s.Path.ValueString(),
					ResponseCode: iaastypes.StringNumber(int(s.Status.ValueInt32())),
				},
			}
			servers = append(servers, server)
		}
		vip := &iaas.LoadBalancerVirtualIPAddress{
			VirtualIPAddress: vip.VIP.ValueString(),
			Port:             iaastypes.StringNumber(int(vip.Port.ValueInt32())),
			DelayLoop:        iaastypes.StringNumber(int(vip.DelayLoop.ValueInt32())),
			SorryServer:      vip.SorryServer.ValueString(),
			Description:      vip.Description.ValueString(),
			Servers:          servers,
		}
		results = append(results, vip)
	}
	return results
}
