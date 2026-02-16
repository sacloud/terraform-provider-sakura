// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	_ resource.ResourceWithModifyPlan     = &workflowResource{}
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
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The creation timestamp of the %s.", resourceName),
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The last update timestamp of the %s.", resourceName),
			},
			"revisions": schema.ListNestedAttribute{
				Required: true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the revision.",
						},
						"alias": schema.StringAttribute{
							Optional:    true,
							Description: "The alias for the revision.",
						},
						"runbook": schema.StringAttribute{
							Required:    true,
							Description: "The runbook definition of the revision.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "The creation timestamp of the revision.",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "The last update timestamp of the revision.",
						},
					},
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

	// check for duplicate runbooks
	for i, rev1 := range config.Revisions {
		for j, rev2 := range config.Revisions[i+1:] {
			if rev1.Runbook == rev2.Runbook {
				resp.Diagnostics.AddAttributeError(
					path.Root("revisions"),
					"Duplicate runbook",
					fmt.Sprintf("Duplicated runbooks not allowed. Runbook at revisions[%d] is duplicate of revisions[%d]", i, j),
				)
			}
		}
	}

	workflowReq, revisionsReq, err := config.toCreateRequest()
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid workflow",
			fmt.Sprintf("Failed to prepare workflow creation request: %s", err),
		)
		return
	}

	if err := workflowReq.Validate(); err != nil {
		resp.Diagnostics.AddError(
			"Invalid workflow configuration",
			fmt.Sprintf("Configuration validation failed: %s", err),
		)
		return
	}

	for revIndex, revReq := range revisionsReq {
		if err := revReq.Validate(); err != nil {
			index := revIndex + 1 // NOTE: 2つめ以降のrevisionsがrevisionsReqに入っている。1つめはcreateWorkflowに含まれるため。
			resp.Diagnostics.AddAttributeError(
				path.Root("revisions").AtListIndex(index),
				"Invalid revision",
				"Validation failed: "+err.Error())
			return
		}
	}
}

// NOTE: stateと比較が必要なバリデーションロジックをModifyPlanで実装している(ValidateConfigではstateが取れないため)
func (r *workflowResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// 作成時などstateが存在しない場合は問題ない。
	if req.State.Raw.IsNull() {
		return
	}
	// 削除時などplanが存在しない場合は問題ない。
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan, state workflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// リビジョン削除の検知
	for i, stateRev := range state.Revisions {
		found := false
		for _, planRev := range plan.Revisions {
			// NOTE: config側のIDが未知なのでRunbookで比較
			if stateRev.Runbook == planRev.Runbook {
				found = true
				break
			}
		}
		if !found {
			resp.Diagnostics.AddAttributeError(path.Root("revisions"),
				"Deleting revision is not supported",
				fmt.Sprintf("Deletion of existing revision is not supported, but revisions[%d] is deleted.", i))
			return
		}
	}

	// リビジョンの順序変更の検知。許容しても良いがdiffになると無駄なapplyが走るため、それを節約するために弾いている。
	for i, stateRev := range state.Revisions {
		if i >= len(plan.Revisions) {
			break
		}
		if stateRev.Runbook != plan.Revisions[i].Runbook {
			resp.Diagnostics.AddAttributeError(path.Root("revisions"),
				"Reordering revisions is not supported",
				"Reorder of revisions is not supported. The order of revisions cannot be changed after creation.")
			return
		}
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

	workflowReq, revisionReqs, err := plan.toCreateRequest()
	if err != nil {
		resp.Diagnostics.AddError("Create: Validation Error", fmt.Sprintf("failed to create Workflow request: %s", err))
		return
	}

	// create workflow with initial revision
	workflowOp := workflows.NewWorkflowOp(r.client)
	workflow, err := workflowOp.Create(ctx, workflowReq)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Workflow: %s", err))
		return
	}

	// create additional revisions if any
	revisionOp := workflows.NewRevisionOp(r.client)
	for _, revReq := range revisionReqs {
		_, err := revisionOp.Create(ctx, workflow.ID, revReq)
		if err != nil {
			resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Workflow revision: %s", err))
			return
		}
	}

	// NOTE: createWorkflowで作ったrevisionのIDはレスポンスに無いため、まとめて一覧取得しステートに保存
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

	workflowReq, updateRevisionReqs, createRevisionReqs, err := plan.toUpdateRequest(&state.workflowBaseModel)
	if err != nil {
		resp.Diagnostics.AddError("Update: Validation Error", fmt.Sprintf("failed to create Workflow update request: %s", err))
		return
	}

	workflow, err := workflowOp.Update(ctx, plan.ID.ValueString(), workflowReq)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Workflow: %s", err))
		return
	}

	// update aliases if any
	for _, revReq := range updateRevisionReqs {
		_, err := revisionOp.UpdateAlias(ctx, workflow.ID, revReq.ID, revReq.Req)
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update alias for Workflow revision: %s", err))
			return
		}
	}

	// create additional revisions if any
	for _, revReq := range createRevisionReqs {
		_, err := revisionOp.Create(ctx, workflow.ID, revReq)
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to create Workflow revision: %s", err))
			return
		}
	}

	// NOTE: createWorkflowで作ったrevisionのIDはレスポンスに無いため、まとめて一覧取得しステートに保存
	revisions, err := revisionOp.List(ctx, v1.ListWorkflowRevisionsParams{ID: workflow.ID})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to list created Workflow revisions: %s", err))
		return
	}

	plan.updateStateFromUpdated(workflow)
	if err := plan.updateRevisionsState(revisions.Revisions); err != nil {
		resp.Diagnostics.AddError("Update: State Error", fmt.Sprintf("failed to update revisions state: %s", err))
		return
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

	resp.State.RemoveResource(ctx)
}
