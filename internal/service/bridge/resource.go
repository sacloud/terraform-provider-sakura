// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bridge

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type bridgeResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &bridgeResource{}
	_ resource.ResourceWithConfigure   = &bridgeResource{}
	_ resource.ResourceWithImportState = &bridgeResource{}
)

func NewBridgeResource() resource.Resource {
	return &bridgeResource{}
}

func (r *bridgeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bridge"
}

func (r *bridgeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type bridgeResourceModel struct {
	bridgeBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *bridgeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Bridge"),
			"name":        common.SchemaResourceName("Bridge"),
			"description": common.SchemaResourceDescription("Bridge"),
			"zone":        common.SchemaResourceZone("Bridge"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *bridgeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *bridgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bridgeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridgeOp := iaas.NewBridgeOp(r.client)
	bridge, err := bridgeOp.Create(ctx, zone, &iaas.BridgeCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("Could not create Bridge: %s", err))
		return
	}

	plan.updateState(bridge, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bridgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bridgeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridge := getBridge(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if bridge == nil || resp.Diagnostics.HasError() {
		return
	}

	state.updateState(bridge, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bridgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bridgeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridgeOp := iaas.NewBridgeOp(r.client)
	bridge, err := bridgeOp.Update(ctx, zone, common.SakuraCloudID(plan.ID.ValueString()), &iaas.BridgeUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Could not update Bridge: %s", err))
		return
	}

	plan.updateState(bridge, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bridgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bridgeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridge := getBridge(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if bridge == nil || resp.Diagnostics.HasError() {
		return
	}

	if err := cleanup.DeleteBridge(ctx, r.client, zone, r.client.GetZones(), bridge.ID, r.client.CheckReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Could not delete Bridge[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getBridge(ctx context.Context, client *common.APIClient, zone string, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Bridge {
	bridgeOp := iaas.NewBridgeOp(client)
	bridge, err := bridgeOp.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("Could not read SakuraCloud Bridge[%s]: %s", id.String(), err))
		return nil
	}
	return bridge
}
