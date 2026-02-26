// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                   = &workflowResource{}
	_ resource.ResourceWithConfigure      = &workflowResource{}
	_ resource.ResourceWithImportState    = &workflowResource{}
	_ resource.ResourceWithValidateConfig = &workflowResource{}
)

func NewWorkflowsResource() resource.Resource {
	return &workflowResource{}
}

func (r *workflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows"
}

func (r *workflowResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.WorkflowsClient
}

type workflowResourceModel struct {
	workflowBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *workflowResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceName := "Workflows"

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaResourceId(resourceName),
			"tags": common.SchemaResourceTags(resourceName),
			"subscription_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The subscription ID of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			// NOTE: 独自のバリデーションルールがあるので、common.SchemaResourceName() は使わない
			"name": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The name of the %s.", resourceName),
			},
			// NOTE: 独自のバリデーションルールがあるので、common.SchemaResourceDescription() は使わない
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The description of the %s.", resourceName),
			},
			"publish": schema.BoolAttribute{
				Required:    true,
				Description: desc.Sprintf("Whether the %s is published.", resourceName),
			},
			"logging": schema.BoolAttribute{
				Required:    true,
				Description: desc.Sprintf("Whether logging is enabled for the %s.", resourceName),
			},
			"service_principal_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The service principal id of the %s.", resourceName),
			},
			"concurrency_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The concurrency mode of the %s.", resourceName),
			},
			"created_at": common.SchemaResourceCreatedAt(resourceName),
			"updated_at": common.SchemaResourceUpdatedAt(resourceName),
			"latest_revision": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The ID of the revision.",
					},
					"runbook": schema.StringAttribute{
						Required:    true,
						Description: "The runbook definition of the revision.",
					},
					"created_at": common.SchemaResourceCreatedAt(resourceName),
					"updated_at": common.SchemaResourceUpdatedAt(resourceName),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Workflow.",
	}
}

func (r *workflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *workflowResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config workflowResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowReq := config.toCreateRequest()
	if err := workflowReq.Validate(); err != nil {
		resp.Diagnostics.AddError(
			"Invalid workflow configuration",
			fmt.Sprintf("Configuration validation failed: %s", err),
		)
		return
	}
}

func (r *workflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	workflowOp := workflows.NewWorkflowOp(r.client)
	workflowReq := plan.toCreateRequest()
	workflow, err := workflowOp.Create(ctx, workflowReq)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Workflow: %s", err))
		return
	}

	// NOTE: createWorkflowで作ったrevisionのIDはレスポンスに無いため、リビジョンの一覧取得にて確認しステートに保存
	revisionOp := workflows.NewRevisionOp(r.client)
	revisions, err := revisionOp.List(ctx, v1.ListWorkflowRevisionsParams{ID: workflow.ID})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to list created Workflow revisions: %s", err))
		return
	}

	plan.updateStateFromCreated(workflow)
	if err := plan.updateRevisionsState(revisions.Revisions); err != nil {
		resp.Diagnostics.AddError("Create: State Error", fmt.Sprintf("failed to update revisions state: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowOp := workflows.NewWorkflowOp(r.client)
	workflow, err := workflowOp.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Workflow: %s", err))
		return
	}

	// NOTE: revisionのIDはRead workflowのレスポンスに無いため、まとめて一覧取得しステートに保存
	revisionOp := workflows.NewRevisionOp(r.client)
	revisions, err := revisionOp.List(ctx, v1.ListWorkflowRevisionsParams{ID: workflow.ID})
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Workflow revisions: %s", err))
		return
	}

	state.updateState(workflow)
	if err := state.updateRevisionsState(revisions.Revisions); err != nil {
		resp.Diagnostics.AddError("Read: State Error", fmt.Sprintf("failed to update revisions state: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state workflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	workflowOp := workflows.NewWorkflowOp(r.client)
	revisionOp := workflows.NewRevisionOp(r.client)

	workflowReq, createRevisionReq := plan.toUpdateRequest(&state.workflowBaseModel)
	workflow, err := workflowOp.Update(ctx, plan.ID.ValueString(), workflowReq)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Workflow: %s", err))
		return
	}

	// create additional revisions if needed
	var latestRevision *v1.CreateWorkflowRevisionCreatedRevision
	if createRevisionReq != nil {
		rev, err := revisionOp.Create(ctx, workflow.ID, *createRevisionReq)
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to create Workflow revision: %s", err))
			return
		}
		latestRevision = rev
	}

	plan.updateStateFromUpdated(workflow)
	if latestRevision != nil {
		plan.LatestRevision.ID = types.StringValue(strconv.Itoa(latestRevision.RevisionId))
		plan.LatestRevision.Runbook = types.StringValue(latestRevision.Runbook)
		plan.LatestRevision.CreatedAt = types.StringValue(latestRevision.CreatedAt.String())
		plan.LatestRevision.UpdatedAt = types.StringValue(latestRevision.UpdatedAt.String())
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	workflowOp := workflows.NewWorkflowOp(r.client)
	err := workflowOp.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Workflow: %s", err))
		return
	}
}
