// Copyright 2016-2025 terraform-provider-sakuracloud authors
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

package sakura

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
)

type bridgeResource struct {
	client *APIClient
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
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

// TODO: model.goに切り出してdata sourceと共通化する
type bridgeResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Zone        types.String   `tfsdk:"zone"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *bridgeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaResourceId("Bridge"),
			"name":        schemaResourceName("Bridge"),
			"description": schemaResourceDescription("Bridge"),
			"zone":        schemaResourceZone("Bridge"),
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

	ctx, cancel := setupTimeoutCreate(ctx, plan.Timeouts, timeout20min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
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
	plan.updateState(ctx, bridge, zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bridgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bridgeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridgeOp := iaas.NewBridgeOp(r.client)
	bridge, err := bridgeOp.Read(ctx, zone, sakuraCloudID(state.ID.ValueString()))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Could not read Bridge: %s", err))
		return
	}
	state.updateState(ctx, bridge, zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bridgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bridgeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutUpdate(ctx, plan.Timeouts, timeout20min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridgeOp := iaas.NewBridgeOp(r.client)
	bridge, err := bridgeOp.Update(ctx, zone, sakuraCloudID(plan.ID.ValueString()), &iaas.BridgeUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("Could not update Bridge: %s", err))
		return
	}
	plan.updateState(ctx, bridge, zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bridgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bridgeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutDelete(ctx, state.Timeouts, timeout20min)
	defer cancel()

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridgeOp := iaas.NewBridgeOp(r.client)
	bridge, err := bridgeOp.Read(ctx, zone, sakuraCloudID(state.ID.ValueString()))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Could not read Bridge[%s]: %s", state.ID.ValueString(), err))
		return
	}

	if err := cleanup.DeleteBridge(ctx, r.client, zone, r.client.zones, bridge.ID, r.client.checkReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Could not delete Bridge[%s]: %s", state.ID.ValueString(), err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (model *bridgeResourceModel) updateState(ctx context.Context, bridge *iaas.Bridge, zone string) {
	model.ID = types.StringValue(bridge.ID.String())
	model.Name = types.StringValue(bridge.Name)
	model.Description = types.StringValue(bridge.Description)
	model.Zone = types.StringValue(zone)
}
