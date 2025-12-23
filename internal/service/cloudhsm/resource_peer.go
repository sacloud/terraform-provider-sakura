// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cloudhsm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sacloud/cloudhsm-api-go"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type cloudHSMPeerResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &cloudHSMPeerResource{}
	_ resource.ResourceWithConfigure   = &cloudHSMPeerResource{}
	_ resource.ResourceWithImportState = &cloudHSMPeerResource{}
)

func NewCloudHSMPeerResource() resource.Resource {
	return &cloudHSMPeerResource{}
}

func (r *cloudHSMPeerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudhsm_peer"
}

func (r *cloudHSMPeerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type cloudHSMPeerResourceModel struct {
	cloudHSMPeerBaseModel
	RouterID  types.String   `tfsdk:"router_id"`
	SecretKey types.String   `tfsdk:"secret_key"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func (r *cloudHSMPeerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaResourceId("CloudHSM Peer"),
			"zone": schemaResourceZone("CloudHSM Peer"),
			"cloudhsm_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the CloudHSM to associate with the peer",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"router_id": schema.StringAttribute{
				Required:    true,
				Description: "The router ID to associate with the peer",
			},
			"secret_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The secret key for the CloudHSM Peer",
			},
			"index": schema.Int64Attribute{
				Computed:    true,
				Description: "The index of the CloudHSM Peer",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the CloudHSM Peer",
			},
			"routes": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The routes for the CloudHSM Peer",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a CloudHSM Peer.",
	}
}

func (r *cloudHSMPeerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *cloudHSMPeerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudHSMPeerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	chsm := getCloudHSM(ctx, client, plan.CloudHSMID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	peerOp, err := cloudhsm.NewPeerOp(client, chsm)
	if err != nil {
		resp.Diagnostics.AddError("Create: Client Error", err.Error())
		return
	}

	err = peerOp.Create(ctx, cloudhsm.CloudHSMPeerCreateParams{
		RouterID:  plan.RouterID.ValueString(),
		SecretKey: plan.SecretKey.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create CloudHSM Peer: %s", err))
		return
	}

	// CloudHSM PeerのIDはRouterIDと同一
	chsmPeer := getCloudHSMPeer(ctx, client, chsm, plan.RouterID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsmPeer == nil {
		return
	}
	plan.updateState(chsmPeer, zone, chsm.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudHSMPeerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudHSMPeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	chsm := getCloudHSM(ctx, client, state.CloudHSMID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	chsmPeer := getCloudHSMPeer(ctx, client, chsm, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsmPeer == nil {
		return
	}

	state.updateState(chsmPeer, zone, chsm.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudHSMPeerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Error", "CloudHSM Peer resource does not support update operation.")
}

func (r *cloudHSMPeerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudHSMPeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	chsm := getCloudHSM(ctx, client, state.CloudHSMID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	chsmPeer := getCloudHSMPeer(ctx, client, chsm, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsmPeer == nil {
		return
	}

	peerOp, _ := cloudhsm.NewPeerOp(client, chsm)
	if err := peerOp.Delete(ctx, chsmPeer.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete CloudHSM Peer[%s]: %s", chsmPeer.ID, err))
		return
	}
}

func getCloudHSMPeer(ctx context.Context, client *v1.Client, chsm *v1.CloudHSM, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.CloudHSMPeer {
	peerOp, err := cloudhsm.NewPeerOp(client, chsm)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("failed to create CloudHSM Peer operation: %s", err))
		return nil
	}

	peers, err := peerOp.List(ctx)
	if err != nil {
		diags.AddError("API List Error", fmt.Sprintf("failed to list CloudHSM Peers for CloudHSM[%s]: %s", chsm.ID, err))
		return nil
	}

	for _, p := range peers {
		if p.ID == id {
			return &p
		}
	}

	state.RemoveResource(ctx)
	return nil
}
