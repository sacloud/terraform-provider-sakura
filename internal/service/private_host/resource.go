// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package private_host

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	"github.com/sacloud/iaas-api-go/search"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type privateHostResource struct {
	client *common.APIClient
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
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type privateHostResourceModel struct {
	privateHostBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *privateHostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	classes := []string{iaastypes.PrivateHostClassDynamic, iaastypes.PrivateHostClassWindows}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("PrivateHost"),
			"name":        common.SchemaResourceName("PrivateHost"),
			"description": common.SchemaResourceDescription("PrivateHost"),
			"tags":        common.SchemaResourceTags("PrivateHost"),
			"icon_id":     common.SchemaResourceIconID("PrivateHost"),
			"zone":        common.SchemaResourceZone("PrivateHost"),
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
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Private Host.",
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

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	phOp := iaas.NewPrivateHostOp(r.client)
	planID, err := expandPrivateHostPlanID(ctx, &plan, r.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create: Expand Error", err.Error())
		return
	}

	ph, err := phOp.Create(ctx, zone, expandPrivateHostCreateRequest(&plan, planID))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create PrivateHost: %s", err))
		return
	}

	plan.updateState(ph, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *privateHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state privateHostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ph := getPrivateHost(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if ph == nil {
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

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	phOp := iaas.NewPrivateHostOp(r.client)
	_, err := phOp.Update(ctx, zone, common.ExpandSakuraCloudID(plan.ID), expandPrivateHostUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update PrivateHost[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	ph := getPrivateHost(ctx, r.client, zone, common.ExpandSakuraCloudID(plan.ID), &resp.State, &resp.Diagnostics)
	if ph == nil {
		return
	}

	plan.updateState(ph, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *privateHostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state privateHostResourceModel
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

	ph := getPrivateHost(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if ph == nil {
		return
	}

	if err := cleanup.DeletePrivateHost(ctx, r.client, zone, ph.ID, r.client.CheckReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete PrivateHost[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getPrivateHost(ctx context.Context, client *common.APIClient, zone string, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.PrivateHost {
	phOp := iaas.NewPrivateHostOp(client)
	ph, err := phOp.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read PrivateHost[%s]: %s", id.String(), err))
		return nil
	}

	return ph
}

func expandPrivateHostPlanID(ctx context.Context, d *privateHostResourceModel, client *common.APIClient, zone string) (iaastypes.ID, error) {
	op := iaas.NewPrivateHostPlanOp(client)
	searched, err := op.Find(ctx, zone, &iaas.FindCondition{
		Filter: search.Filter{search.Key("Class"): search.ExactMatch(d.Class.ValueString())},
	})
	if err != nil {
		return iaastypes.ID(0), err
	}
	if searched.Count == 0 {
		return iaastypes.ID(0), errors.New("failed to find PrivateHostPlan: plan is not found")
	}

	return searched.PrivateHostPlans[0].ID, nil
}

func expandPrivateHostCreateRequest(model *privateHostResourceModel, planID iaastypes.ID) *iaas.PrivateHostCreateRequest {
	return &iaas.PrivateHostCreateRequest{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Tags:        common.TsetToStrings(model.Tags),
		IconID:      common.ExpandSakuraCloudID(model.IconID),
		PlanID:      planID,
	}
}

func expandPrivateHostUpdateRequest(model *privateHostResourceModel) *iaas.PrivateHostUpdateRequest {
	return &iaas.PrivateHostUpdateRequest{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Tags:        common.TsetToStrings(model.Tags),
		IconID:      common.ExpandSakuraCloudID(model.IconID),
	}
}
