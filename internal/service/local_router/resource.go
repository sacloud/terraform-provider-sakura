// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package local_router

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	localrouter "github.com/sacloud/iaas-service-go/localrouter/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type localRouterResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &localRouterResource{}
	_ resource.ResourceWithConfigure   = &localRouterResource{}
	_ resource.ResourceWithImportState = &localRouterResource{}
)

func NewLocalRouterResource() resource.Resource {
	return &localRouterResource{}
}

type localRouterResourceModel struct {
	localRouterBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *localRouterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_router"
}

func (r *localRouterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

func (r *localRouterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Local Router"),
			"name":        common.SchemaResourceName("Local Router"),
			"description": common.SchemaResourceDescription("Local Router"),
			"tags":        common.SchemaResourceTags("Local Router"),
			"icon_id":     common.SchemaResourceIconID("Local Router"),
			// sakura_vswitch以外の他のサービスのスイッチとも繋がるので、パラメータ名はswitchのままにする
			"switch": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"code": schema.StringAttribute{
						Required:    true,
						Description: "The resource ID of the Switch",
						Validators: []validator.String{
							sacloudvalidator.SakuraIDValidator(),
						},
					},
					"category": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("cloud"),
						Description: "The category name of connected services (e.g. `cloud`, `vps`)",
					},
					"zone": schema.StringAttribute{
						Required:    true,
						Description: "The name of the Zone",
					},
				},
			},
			"network_interface": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"vip": schema.StringAttribute{
						Required:    true,
						Description: "The virtual IP address",
						Validators: []validator.String{
							sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
						},
					},
					"ip_addresses": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The list of the IP address assigned",
						Validators: []validator.List{
							listvalidator.ValueStringsAre(sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4)),
							listvalidator.SizeAtLeast(2),
							listvalidator.SizeAtMost(2),
						},
					},
					"netmask": schema.Int32Attribute{
						Required:    true,
						Description: "The bit length of the subnet assigned to the network interface",
						Validators: []validator.Int32{
							int32validator.Between(8, 29),
						},
					},
					"vrid": schema.Int64Attribute{
						Required:    true,
						Description: "The Virtual Router Identifier",
					},
				},
			},
			"peer": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"peer_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the peer LocalRouter",
							Validators: []validator.String{
								sacloudvalidator.SakuraIDValidator(),
							},
						},
						"secret_key": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The secret key of the peer LocalRouter",
						},
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "The flag to enable the LocalRouter",
						},
						"description": common.SchemaResourceDescription("Local Router Peer"),
					},
				},
			},
			"static_route": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"prefix": schema.StringAttribute{
							Required:    true,
							Description: "The CIDR block of destination",
						},
						"next_hop": schema.StringAttribute{
							Required:    true,
							Description: "The IP address of the next hop",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
					},
				},
			},
			"secret_keys": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Sensitive:   true,
				Description: "A list of secret key used for peering from other LocalRouters",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Local Router.",
	}
}

func (r *localRouterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *localRouterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan localRouterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	builder := expandLocalRouterBuilder(&plan, r.client)
	if err := builder.Validate(ctx); err != nil {
		resp.Diagnostics.AddError("Create: Validation Error", fmt.Sprintf("failed to validate parameter for LocalRouter: %s", err))
		return
	}
	lr, err := builder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create LocalRouter: %s", err))
		return
	}

	plan.updateState(lr)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *localRouterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state localRouterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lr := getLocalRouter(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if lr == nil {
		return
	}

	state.updateState(lr)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *localRouterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan localRouterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	id := plan.ID.ValueString()
	common.SakuraMutexKV.Lock(id)
	defer common.SakuraMutexKV.Unlock(id)

	builder := expandLocalRouterBuilder(&plan, r.client)
	if err := builder.Validate(ctx); err != nil {
		resp.Diagnostics.AddError("Update: Validation Error", fmt.Sprintf("failed to validate parameter for LocalRouter[%s]: %s", id, err))
		return
	}
	updated, err := builder.Update(ctx, common.SakuraCloudID(id))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update LocalRouter[%s]: %s", id, err))
		return
	}

	lr := getLocalRouter(ctx, r.client, updated.ID.String(), &resp.State, &resp.Diagnostics)
	if lr == nil {
		return
	}

	plan.updateState(lr)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *localRouterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state localRouterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	id := state.ID.ValueString()
	common.SakuraMutexKV.Lock(id)
	defer common.SakuraMutexKV.Unlock(id)

	lr := getLocalRouter(ctx, r.client, id, &resp.State, &resp.Diagnostics)
	if lr == nil {
		return
	}

	if err := iaas.NewLocalRouterOp(r.client).Delete(ctx, lr.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete LocalRouter[%s]: %s", id, err))
		return
	}
}

func getLocalRouter(ctx context.Context, client *common.APIClient, id string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.LocalRouter {
	lrOp := iaas.NewLocalRouterOp(client)
	lr, err := lrOp.Read(ctx, common.SakuraCloudID(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read LocalRouter[%s]: %s", id, err))
		return nil
	}
	return lr
}

func expandLocalRouterBuilder(model *localRouterResourceModel, client *common.APIClient) *localrouter.Builder {
	b := &localrouter.Builder{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Tags:        common.TsetToStrings(model.Tags),
		IconID:      common.ExpandSakuraCloudID(model.IconID),
		Client:      localrouter.NewAPIClient(client),
	}

	if model.Switch != nil {
		b.Switch = &iaas.LocalRouterSwitch{
			Code:     model.Switch.Code.ValueString(),
			Category: model.Switch.Category.ValueString(),
			ZoneID:   model.Switch.Zone.ValueString(),
		}
	}

	if model.NetworkInterface != nil {
		b.Interface = &iaas.LocalRouterInterface{
			VirtualIPAddress: model.NetworkInterface.VIP.ValueString(),
			IPAddress:        common.TlistToStrings(model.NetworkInterface.IPAddresses),
			NetworkMaskLen:   int(model.NetworkInterface.Netmask.ValueInt32()),
			VRID:             int(model.NetworkInterface.VRID.ValueInt64()),
		}
	}

	if len(model.Peer) > 0 {
		for _, p := range model.Peer {
			b.Peers = append(b.Peers, &iaas.LocalRouterPeer{
				ID:          common.ExpandSakuraCloudID(p.PeerID),
				SecretKey:   p.SecretKey.ValueString(),
				Enabled:     p.Enabled.ValueBool(),
				Description: p.Description.ValueString(),
			})
		}
	}

	if len(model.StaticRoute) > 0 {
		for _, sr := range model.StaticRoute {
			b.StaticRoutes = append(b.StaticRoutes, &iaas.LocalRouterStaticRoute{
				Prefix:  sr.Prefix.ValueString(),
				NextHop: sr.NextHop.ValueString(),
			})
		}
	}

	return b
}
