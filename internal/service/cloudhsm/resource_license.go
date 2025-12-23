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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/cloudhsm-api-go"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type cloudHSMLicenseResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &cloudHSMLicenseResource{}
	_ resource.ResourceWithConfigure   = &cloudHSMLicenseResource{}
	_ resource.ResourceWithImportState = &cloudHSMLicenseResource{}
)

func NewCloudHSMLicenseResource() resource.Resource {
	return &cloudHSMLicenseResource{}
}

func (r *cloudHSMLicenseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudhsm_license"
}

func (r *cloudHSMLicenseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type cloudHSMLicenseResourceModel struct {
	cloudHSMLicenseBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *cloudHSMLicenseResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("CloudHSM License"),
			"name":        common.SchemaResourceName("CloudHSM License"),
			"description": common.SchemaResourceDescription("CloudHSM License"),
			"tags":        common.SchemaResourceTags("CloudHSM License"),
			"zone":        schemaResourceZone("CloudHSM License"),
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation date of the CloudHSM License",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "The modification date of the CloudHSM License",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a CloudHSM License.",
	}
}

func (r *cloudHSMLicenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *cloudHSMLicenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudHSMLicenseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	licenseOp := cloudhsm.NewLicenseOp(client)
	created, err := licenseOp.Create(ctx, cloudhsm.CloudHSMSoftwareLicenseCreateParams{
		Name:        plan.Name.ValueString(),
		Description: common.Ptr(plan.Description.ValueString()),
		Tags:        common.TsetToStrings(plan.Tags),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create CloudHSM License: %s", err.Error()))
		return
	}

	license := getCloudHSMLicense(ctx, client, created.ID, &resp.State, &resp.Diagnostics)
	if license == nil {
		return
	}

	plan.updateState(license, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudHSMLicenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudHSMLicenseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	license := getCloudHSMLicense(ctx, client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if license == nil {
		return
	}

	state.updateState(license, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudHSMLicenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cloudHSMLicenseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	updated, err := cloudhsm.NewLicenseOp(client).Update(ctx, plan.ID.ValueString(), cloudhsm.CloudHSMSoftwareLicenseUpdateParams{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Tags:        common.TsetToStrings(plan.Tags),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update CloudHSM License[%s]: %s", plan.ID.ValueString(), err.Error()))
		return
	}

	plan.updateState(updated, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudHSMLicenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudHSMLicenseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	license := getCloudHSMLicense(ctx, client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if license == nil {
		return
	}

	if err := cloudhsm.NewLicenseOp(client).Delete(ctx, license.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete CloudHSM License[%s]: %s", license.ID, err.Error()))
		return
	}
}

func getCloudHSMLicense(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.CloudHSMSoftwareLicense {
	licenseOp := cloudhsm.NewLicenseOp(client)
	license, err := licenseOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read CloudHSM License[%s]: %s", id, err))
		return nil
	}
	return license
}
