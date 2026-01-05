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
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/cloudhsm-api-go"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type cloudHSMClientResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &cloudHSMClientResource{}
	_ resource.ResourceWithConfigure   = &cloudHSMClientResource{}
	_ resource.ResourceWithImportState = &cloudHSMClientResource{}
)

func NewCloudHSMClientResource() resource.Resource {
	return &cloudHSMClientResource{}
}

func (r *cloudHSMClientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudhsm_client"
}

func (r *cloudHSMClientResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type cloudHSMClientResourceModel struct {
	cloudHSMClientBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *cloudHSMClientResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaResourceId("CloudHSM Client"),
			"name": common.SchemaResourceName("CloudHSM Client"),
			"zone": schemaResourceZone("CloudHSM Client"),
			"cloudhsm_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the CloudHSM to associate with the client",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"certificate": schema.StringAttribute{
				Required:    true,
				Description: "The certificate for the CloudHSM Client",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation date of the CloudHSM Client",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "The modification date of the CloudHSM Client",
			},
			"availability": schema.StringAttribute{
				Computed:    true,
				Description: "The availability status of the CloudHSM Client",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a CloudHSM Client.",
	}
}

func (r *cloudHSMClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *cloudHSMClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudHSMClientResourceModel
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

	clientOp, _ := cloudhsm.NewClientOp(client, chsm)
	created, err := clientOp.Create(ctx, cloudhsm.CloudHSMClientCreateParams{
		Name:        plan.Name.ValueString(),
		Certificate: plan.Certificate.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", err.Error())
		return
	}

	plan.updateState(created, zone, chsm.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudHSMClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudHSMClientResourceModel
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

	chsmClient := getCloudHSMClient(ctx, client, chsm, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsmClient == nil {
		return
	}

	state.updateState(chsmClient, zone, chsm.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudHSMClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cloudHSMClientResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	chsm := getCloudHSM(ctx, client, plan.CloudHSMID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	clientOp, _ := cloudhsm.NewClientOp(client, chsm)
	updated, err := clientOp.Update(ctx, plan.ID.ValueString(), cloudhsm.CloudHSMClientUpdateParams{Name: plan.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", err.Error())
		return
	}

	plan.updateState(updated, zone, chsm.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudHSMClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudHSMClientResourceModel
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

	clientOp, err := cloudhsm.NewClientOp(client, chsm)
	if err != nil {
		resp.Diagnostics.AddError("Delete: Client Error", err.Error())
		return
	}

	if err := clientOp.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", err.Error())
		return
	}
}

func getCloudHSMClient(ctx context.Context, client *v1.Client, chsm *v1.CloudHSM, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.CloudHSMClient {
	clientOp, _ := cloudhsm.NewClientOp(client, chsm)
	chsmClient, err := clientOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read CloudHSM Client[%s]: %s", id, err))
		return nil
	}
	return chsmClient
}
