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
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	"github.com/sacloud/iaas-api-go/search"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type privateHostResource struct {
	client *APIClient
}

var (
	_ resource.Resource                = &privateHostResource{}
	_ resource.ResourceWithConfigure   = &privateHostResource{}
	_ resource.ResourceWithImportState = &privateHostResource{}
)

func NewPrivateHostResource() resource.Resource {
	return &privateHostResource{}
}

func (r *privateHostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_host"
}

func (r *privateHostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type privateHostResourceModel struct {
	sakuraPrivateHostBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *privateHostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	classes := []string{iaastypes.PrivateHostClassDynamic, iaastypes.PrivateHostClassWindows}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaResourceId("PrivateHost"),
			"name":        schemaResourceName("PrivateHost"),
			"description": schemaResourceDescription("PrivateHost"),
			"tags":        schemaResourceTags("PrivateHost"),
			"icon_id":     schemaResourceIconID("PrivateHost"),
			"zone":        schemaResourceZone("PrivateHost"),
			"class": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The class of the PrivateHost. This will be one of [%s]", classes),
				Default:     stringdefault.StaticString(iaastypes.PrivateHostClassDynamic),
				Validators: []validator.String{
					stringvalidator.OneOf(classes...),
				},
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the private host",
			},
			"assigned_core": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPUs assigned to servers on the private host",
			},
			"assigned_memory": schema.Int32Attribute{
				Computed:    true,
				Description: "The total size of memory assigned to servers on the private host",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *privateHostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *privateHostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan privateHostResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutCreate(ctx, plan.Timeouts, timeout5min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	phOp := iaas.NewPrivateHostOp(r.client)
	planID, err := expandPrivateHostPlanID(ctx, &plan, r.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	ph, err := phOp.Create(ctx, zone, expandPrivateHostCreateRequest(&plan, planID))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud PrivateHost is failed: %s", err))
		return
	}

	updateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, ph.ID.String(), zone)
}

func (r *privateHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state privateHostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	phOp := iaas.NewPrivateHostOp(r.client)
	ph, err := phOp.Read(ctx, zone, expandSakuraCloudID(state.ID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not read SakuraCloud PrivateHost[%s]: %s", state.ID.ValueString(), err))
		return
	}

	state.updateState(ph, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *privateHostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan privateHostResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutUpdate(ctx, plan.Timeouts, timeout5min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	phOp := iaas.NewPrivateHostOp(r.client)
	ph, err := phOp.Read(ctx, zone, expandSakuraCloudID(plan.ID))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("could not read SakuraCloud PrivateHost[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	_, err = phOp.Update(ctx, zone, ph.ID, expandPrivateHostUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud PrivateHost[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	updateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, ph.ID.String(), zone)
}

func (r *privateHostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state privateHostResourceModel
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

	phOp := iaas.NewPrivateHostOp(r.client)
	ph, err := phOp.Read(ctx, zone, expandSakuraCloudID(state.ID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not read SakuraCloud PrivateHost[%s]: %s", state.ID.ValueString(), err))
		return
	}

	if err := cleanup.DeletePrivateHost(ctx, r.client, zone, ph.ID, r.client.checkReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud PrivateHost[%s] is failed: %s", state.ID.ValueString(), err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func expandPrivateHostPlanID(ctx context.Context, d *privateHostResourceModel, client *APIClient, zone string) (iaastypes.ID, error) {
	op := iaas.NewPrivateHostPlanOp(client)
	searched, err := op.Find(ctx, zone, &iaas.FindCondition{
		Filter: search.Filter{search.Key("Class"): search.ExactMatch(d.Class.ValueString())},
	})
	if err != nil {
		return iaastypes.ID(0), err
	}
	if searched.Count == 0 {
		return iaastypes.ID(0), errors.New("finding PrivateHostPlan is failed: plan is not found")
	}

	return searched.PrivateHostPlans[0].ID, nil
}

func expandPrivateHostCreateRequest(model *privateHostResourceModel, planID iaastypes.ID) *iaas.PrivateHostCreateRequest {
	return &iaas.PrivateHostCreateRequest{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Tags:        tsetToStrings(model.Tags),
		IconID:      expandSakuraCloudID(model.IconID),
		PlanID:      planID,
	}
}

func expandPrivateHostUpdateRequest(model *privateHostResourceModel) *iaas.PrivateHostUpdateRequest {
	return &iaas.PrivateHostUpdateRequest{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Tags:        tsetToStrings(model.Tags),
		IconID:      expandSakuraCloudID(model.IconID),
	}
}
