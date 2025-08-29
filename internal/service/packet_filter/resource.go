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

package packet_filter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

type packetFilterResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &packetFilterResource{}
	_ resource.ResourceWithConfigure   = &packetFilterResource{}
	_ resource.ResourceWithImportState = &packetFilterResource{}
)

func NewPacketFilterResource() resource.Resource {
	return &packetFilterResource{}
}

func (r *packetFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_packet_filter"
}

func (r *packetFilterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type packetFilterResourceModel struct {
	packetFilterBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *packetFilterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Packet Filter"),
			"name":        common.SchemaResourceName("Packet Filter"),
			"description": common.SchemaResourceDescription("Packet Filter"),
			"zone":        common.SchemaResourceZone("Packet Filter"),
			"expression":  schemaPacketFilterExpression(),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *packetFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *packetFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan packetFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf, err := pfOp.Create(ctx, zone, &iaas.PacketFilterCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Expression:  expandPacketFilterExpressions(plan.Expression),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud PacketFilter is failed: %s", err))
		return
	}

	common.UpdateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, pf.ID.String(), zone)
}

func (r *packetFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state packetFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf, err := pfOp.Read(ctx, zone, common.ExpandSakuraCloudID(state.ID))
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

func (r *packetFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state packetFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf, err := pfOp.Read(ctx, zone, common.ExpandSakuraCloudID(plan.ID))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("could not read SakuraCloud PacketFilter[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	_, err = pfOp.Update(ctx, zone, pf.ID, expandPacketFilterUpdateRequest(&plan, &state, pf), pf.ExpressionHash)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud PacketFilter[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	common.UpdateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, pf.ID.String(), zone)
}

func (r *packetFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state packetFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	pfOp := iaas.NewPacketFilterOp(r.client)
	pf, err := pfOp.Read(ctx, zone, common.ExpandSakuraCloudID(state.ID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not read SakuraCloud PacketFilter[%s]: %s", state.ID.ValueString(), err))
		return
	}

	if err := cleanup.DeletePacketFilter(ctx, r.client, zone, pf.ID, r.client.CheckReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud PacketFilter[%s] is failed: %s", state.ID.ValueString(), err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func expandPacketFilterUpdateRequest(plan *packetFilterResourceModel, state *packetFilterResourceModel, pf *iaas.PacketFilter) *iaas.PacketFilterUpdateRequest {
	expressions := pf.Expression
	if common.HasChange(plan.Expression, state.Expression) {
		expressions = expandPacketFilterExpressions(plan.Expression)
	}

	return &iaas.PacketFilterUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Expression:  expressions,
	}
}
