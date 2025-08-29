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

package sw1tch

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/validators"
)

type switchResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &switchResource{}
	_ resource.ResourceWithConfigure   = &switchResource{}
	_ resource.ResourceWithImportState = &switchResource{}
)

func NewSwitchResource() resource.Resource {
	return &switchResource{}
}

func (r *switchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch"
}

func (r *switchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type switchResourceModel struct {
	switchBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *switchResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Switch"),
			"name":        common.SchemaDataSourceName("Switch"),
			"icon_id":     common.SchemaResourceIconID("Switch"),
			"description": common.SchemaResourceDescription("Switch"),
			"tags":        common.SchemaResourceTags("Switch"),
			"zone":        common.SchemaResourceZone("Switch"),
			"bridge_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The bridge id attached to the switch",
				Validators: []validator.String{
					validators.SakuraIDValidator(),
				},
			},
			"server_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of server ids connected to the switch",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.SakuraIDValidator()),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *switchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *switchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan switchResourceModel
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

	swOp := iaas.NewSwitchOp(r.client)
	sw, err := swOp.Create(ctx, zone, &iaas.SwitchCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Tags:        common.TsetToStrings(plan.Tags),
		IconID:      common.ExpandSakuraCloudID(plan.IconID),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud Switch is failed: %s", err))
		return
	}

	if !plan.BridgeID.IsNull() && plan.BridgeID.ValueString() != "" {
		brId := common.ExpandSakuraCloudID(plan.BridgeID)
		if err := swOp.ConnectToBridge(ctx, zone, sw.ID, brId); err != nil {
			resp.Diagnostics.AddError("Bridge Connect Error",
				fmt.Sprintf("connecting Switch[%s] to Bridge[%s] is failed: %s", sw.ID, brId, err))
			return
		}
	}

	common.UpdateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, sw.ID.String(), zone)
}

func (r *switchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state switchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	swOp := iaas.NewSwitchOp(r.client)
	sw, err := swOp.Read(ctx, zone, common.SakuraCloudID(state.ID.ValueString()))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error",
			fmt.Sprintf("could not read SakuraCloud Switch[%s] : %s", state.ID.ValueString(), err))
		return
	}

	if err := state.updateState(ctx, r.client, sw, zone); err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *switchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state switchResourceModel
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

	sid := state.ID.ValueString()
	common.SakuraMutexKV.Lock(sid)
	defer common.SakuraMutexKV.Unlock(sid)

	swOp := iaas.NewSwitchOp(r.client)
	sw, err := swOp.Read(ctx, zone, common.SakuraCloudID(sid))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}

	sw, err = swOp.Update(ctx, zone, sw.ID, &iaas.SwitchUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Tags:        common.TsetToStrings(plan.Tags),
		IconID:      common.ExpandSakuraCloudID(plan.IconID),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Error",
			fmt.Sprintf("updating SakuraCloud Switch[%s] is failed : %s", plan.ID.ValueString(), err))
		return
	}

	if plan.BridgeID.ValueString() != state.BridgeID.ValueString() { // HasChange in SDK v2
		if !plan.BridgeID.IsNull() {
			brId := plan.BridgeID.ValueString()
			if brId == "" && !sw.BridgeID.IsEmpty() {
				if err := swOp.DisconnectFromBridge(ctx, zone, sw.ID); err != nil {
					resp.Diagnostics.AddError("Update Error",
						fmt.Sprintf("disconnecting from Bridge[%s] is failed: %s", sw.BridgeID, err))
					return
				}
			} else {
				if err := swOp.ConnectToBridge(ctx, zone, sw.ID, common.SakuraCloudID(brId)); err != nil {
					resp.Diagnostics.AddError("Update Error",
						fmt.Sprintf("connecting to Bridge[%s] is failed: %s", brId, err))
					return
				}
			}
		}
	}

	common.UpdateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, sw.ID.String(), zone)
}

func (r *switchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state switchResourceModel
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

	sid := state.ID.ValueString()
	common.SakuraMutexKV.Lock(sid)
	defer common.SakuraMutexKV.Unlock(sid)

	swOp := iaas.NewSwitchOp(r.client)
	sw, err := swOp.Read(ctx, zone, common.SakuraCloudID(sid))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error",
			fmt.Sprintf("could not read SakuraCloud Switch[%s]: %s", state.ID.ValueString(), err))
		return
	}

	if !sw.BridgeID.IsEmpty() {
		if err := swOp.DisconnectFromBridge(ctx, zone, sw.ID); err != nil {
			resp.Diagnostics.AddError("Delete Error",
				fmt.Sprintf("disconnecting Switch[%s] from Bridge[%s] is failed: %s", sw.ID, sw.BridgeID, err))
			return
		}
	}

	if err := cleanup.DeleteSwitch(ctx, r.client, zone, sw.ID, r.client.CheckReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete Error",
			fmt.Sprintf("deleting SakuraCloud Switch[%s] is failed: %s", state.ID.ValueString(), err))
		return
	}

	resp.State.RemoveResource(ctx)
}
