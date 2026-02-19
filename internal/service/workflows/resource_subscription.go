// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowsSubscriptionResource struct {
	client *v1.Client
}

func NewSubscriptionResource() resource.Resource {
	return &workflowsSubscriptionResource{}
}

var (
	_ resource.Resource                = &workflowsSubscriptionResource{}
	_ resource.ResourceWithConfigure   = &workflowsSubscriptionResource{}
	_ resource.ResourceWithImportState = &workflowsSubscriptionResource{}
)

func (r *workflowsSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows_subscription"
}

func (r *workflowsSubscriptionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.WorkflowsClient
}

type workflowsSubscriptionResourceModel struct {
	workflowsSubscriptionBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *workflowsSubscriptionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceName := "Workflows Subscription"

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId(resourceName),
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The account ID of the %s.", resourceName),
			},
			"contract_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The contract ID of the %s.", resourceName),
			},
			"plan_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The plan ID of the %s.", resourceName),
				// NOTE: updateでもIDが変わる(課金設定毎にIDが払い出される)ようなので、plan_idの変更でRequiresReplace
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan_name": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The plan name of the %s.", resourceName),
			},
			"activate_from": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The activate from timestamp of the %s.", resourceName),
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The creation timestamp of the %s.", resourceName),
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The last update timestamp of the %s.", resourceName),
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a current Workflows Subscription. Only one subscription can exist at a time. If a subscription already configured, it will be overwritten by the configuration in terraform.",
	}
}

func (r *workflowsSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *workflowsSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workflowsSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	if err := r.setSubscription(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Create: API Error",
			fmt.Sprintf("failed to create Workflows Subscription: %s", err))
		return
	}

	subscriptionOp := workflows.NewSubscriptionOp(r.client)
	data, err := subscriptionOp.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read Workflows Subscription: %s", err))
		return
	}

	plan.updateState(data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowsSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workflowsSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subscriptionOp := workflows.NewSubscriptionOp(r.client)
	data, err := subscriptionOp.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Workflows Subscription: %s", err))
		return
	}

	state.updateState(data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workflowsSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workflowsSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	if err := r.setSubscription(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Update: API Error",
			fmt.Sprintf("failed to update Workflows Subscription: %s", err))
		return
	}

	subscriptionOp := workflows.NewSubscriptionOp(r.client)
	data, err := subscriptionOp.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read Workflows Subscription: %s", err))
		return
	}

	plan.updateState(data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowsSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workflowsSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	subscriptionOp := workflows.NewSubscriptionOp(r.client)
	if err := subscriptionOp.Delete(ctx); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Workflows Subscription: %s", err))
		return
	}
}

func (r *workflowsSubscriptionResource) setSubscription(ctx context.Context, plan workflowsSubscriptionResourceModel) error {
	planID := plan.PlanID.ValueString()
	id, err := strconv.ParseFloat(planID, 64)
	if err != nil {
		return err
	}

	subscriptionOp := workflows.NewSubscriptionOp(r.client)
	if err := subscriptionOp.Create(ctx, v1.CreateSubscriptionReq{PlanId: id}); err != nil {
		return err
	}
	return nil
}
