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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
)

type packetFilterRulesResource struct {
	client *APIClient
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
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type packetFilterRulesResourceModel struct {
	ID             types.String                         `tfsdk:"id"`
	Zone           types.String                         `tfsdk:"zone"`
	PacketFilterID types.String                         `tfsdk:"packet_filter_id"`
	Expression     []*sakuraPacketFilterExpressionModel `tfsdk:"expression"`
	Timeouts       timeouts.Value                       `tfsdk:"timeouts"`
}

func (r *packetFilterRulesResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   schemaResourceId("Packet Filter Rules"),
			"zone": schemaResourceZone("Packet Filter Rules"),
			"packet_filter_id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the packet filter that set expressions to",
				Validators: []validator.String{
					sakuraIDValidator(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"expression": schemaPacketFilterExpression(),
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
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

	ctx, cancel := setupTimeoutCreate(ctx, plan.Timeouts, timeout5min)
	defer cancel()

	callPacketFilterRulesUpdate(ctx, r, &plan, &resp.State, &resp.Diagnostics)
}

func (r *packetFilterRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state packetFilterRulesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf, err := pfOp.Read(ctx, zone, expandSakuraCloudID(state.ID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not read SakuraCloud PacketFilter[%s]: %s", state.ID.ValueString(), err))
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

	ctx, cancel := setupTimeoutUpdate(ctx, plan.Timeouts, timeout5min)
	defer cancel()

	callPacketFilterRulesUpdate(ctx, r, &plan, &resp.State, &resp.Diagnostics)
}

func (r *packetFilterRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state packetFilterRulesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	pfID := state.PacketFilterID.ValueString()

	ctx, cancel := setupTimeoutDelete(ctx, state.Timeouts, timeout5min)
	defer cancel()

	sakuraMutexKV.Lock(pfID)
	defer sakuraMutexKV.Unlock(pfID)

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf, err := pfOp.Read(ctx, zone, sakuraCloudID(pfID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not read SakuraCloud PacketFilter[%s]: %s", pfID, err))
		return
	}

	_, err = pfOp.Update(ctx, zone, pf.ID, &iaas.PacketFilterUpdateRequest{
		Name:        pf.Name,
		Description: pf.Description,
		Expression:  []*iaas.PacketFilterExpression{}, // Set empty expressions to delete all rules
	}, pf.ExpressionHash)
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("updating SakuraCloud PacketFilter[%s] is failed: %s", pfID, err))
		return
	}
}

func (model *packetFilterRulesResourceModel) updateState(pf *iaas.PacketFilter, zone string) {
	model.ID = types.StringValue(pf.ID.String())
	model.Zone = types.StringValue(zone)
	model.PacketFilterID = types.StringValue(pf.ID.String())
	model.Expression = flattenPacketFilterExpressions(pf)
}

func callPacketFilterRulesUpdate(ctx context.Context, r *packetFilterRulesResource, plan *packetFilterRulesResourceModel, state *tfsdk.State, diags *diag.Diagnostics) {
	zone := getZone(plan.Zone, r.client, diags)
	if diags.HasError() {
		return
	}
	pfID := plan.PacketFilterID.ValueString()

	sakuraMutexKV.Lock(pfID)
	defer sakuraMutexKV.Unlock(pfID)

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf, err := pfOp.Read(ctx, zone, sakuraCloudID(pfID))
	if err != nil {
		diags.AddError("Update Error", fmt.Sprintf("could not read SakuraCloud PacketFilter[%s]: %s", pfID, err))
		return
	}

	_, err = pfOp.Update(ctx, zone, pf.ID, &iaas.PacketFilterUpdateRequest{
		Name:        pf.Name,
		Description: pf.Description,
		Expression:  expandPacketFilterExpressions(plan.Expression),
	}, pf.ExpressionHash)
	if err != nil {
		diags.AddError("Update Error", fmt.Sprintf("updating SakuraCloud PacketFilter[%s] is failed: %s", pfID, err))
		return
	}

	updateResourceByReadWithZone(ctx, r, state, diags, pfID, zone)
}
