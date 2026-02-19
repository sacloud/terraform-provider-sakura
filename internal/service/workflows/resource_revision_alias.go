// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowRevisionAliasResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                   = &workflowRevisionAliasResource{}
	_ resource.ResourceWithConfigure      = &workflowRevisionAliasResource{}
	_ resource.ResourceWithImportState    = &workflowRevisionAliasResource{}
	_ resource.ResourceWithValidateConfig = &workflowRevisionAliasResource{}
)

func NewWorkflowsRevisionAliasResource() resource.Resource {
	return &workflowRevisionAliasResource{}
}

func (r *workflowRevisionAliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows_revision_alias"
}

func (r *workflowRevisionAliasResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.WorkflowsClient
}

type workflowRevisionAliasResourceModel struct {
	workflowRevisionAliasBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *workflowRevisionAliasResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceName := "Workflows RevisionAlias"

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"workflow_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The workflow ID of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"revision_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The revision ID of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d+$`), "needs to be a string representation of an integer."),
				},
			},
			"alias": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The alias name of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Workflows Revision Alias.",
	}
}

func (r *workflowRevisionAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *workflowRevisionAliasResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config workflowRevisionAliasResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasReq := config.toUpdateRequest()
	if err := aliasReq.Validate(); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("alias"),
			"Invalid workflows revision alias",
			fmt.Sprintf("Revision alias validation failed: %s", err),
		)
		return
	}
}

func (r *workflowRevisionAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workflowRevisionAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	revisionIDStr := plan.RevisionID.ValueString()
	revisionID, err := strconv.Atoi(revisionIDStr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("revision_id"),
			"Create: validation error",
			"Invalid revision_id format. Expected an integer value.")
		return
	}

	revisionOp := workflows.NewRevisionOp(r.client)
	rev, err := revisionOp.UpdateAlias(ctx, plan.WorkflowID.ValueString(), revisionID, plan.toUpdateRequest())
	if err != nil {
		resp.Diagnostics.AddError(
			"Create: API Error",
			fmt.Sprintf("failed to create workflow revision alias: %s", err),
		)
		return
	}

	plan.updateStateFromCreated(rev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowRevisionAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workflowRevisionAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionIDStr := state.RevisionID.ValueString()
	revisionID, err := strconv.Atoi(revisionIDStr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("revision_id"),
			"Read: Validation error",
			"Invalid revision_id format. Expected an integer value.")
		return
	}

	revisionOp := workflows.NewRevisionOp(r.client)
	rev, err := revisionOp.Read(ctx, state.WorkflowID.ValueString(), revisionID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Read: API Error",
			fmt.Sprintf("failed to read workflow revision: %s", err),
		)
		return
	}

	state.updateStateFromRead(rev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workflowRevisionAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state workflowRevisionAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	revisionIDStr := plan.RevisionID.ValueString()
	revisionID, err := strconv.Atoi(revisionIDStr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("revision_id"),
			"Update: validation error",
			"Invalid revision_id format. Expected an integer value.")
		return
	}

	revisionOp := workflows.NewRevisionOp(r.client)
	rev, err := revisionOp.UpdateAlias(ctx, plan.WorkflowID.ValueString(), revisionID, plan.toUpdateRequest())
	if err != nil {
		resp.Diagnostics.AddError(
			"Update: API Error",
			fmt.Sprintf("failed to update workflow revision alias: %s", err),
		)
		return
	}

	plan.updateStateFromUpdated(rev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowRevisionAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workflowRevisionAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	revisionIDStr := state.RevisionID.ValueString()
	revisionID, err := strconv.Atoi(revisionIDStr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("revision_id"),
			"Delete: validation error",
			"Invalid revision_id format. Expected an integer value.")
		return
	}

	revisionOp := workflows.NewRevisionOp(r.client)

	if err := revisionOp.DeleteAlias(ctx, state.WorkflowID.ValueString(), revisionID); err != nil {
		resp.Diagnostics.AddError(
			"Delete: API Error",
			fmt.Sprintf("failed to delete workflow revision alias: %s", err),
		)
		return
	}
}
