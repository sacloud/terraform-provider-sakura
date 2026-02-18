// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowRevisionAliasBaseModel struct {
	WorkflowID types.String `tfsdk:"workflow_id"`
	RevisionID types.String `tfsdk:"revision_id"`
	Alias      types.String `tfsdk:"alias"`
}

func (model *workflowRevisionAliasBaseModel) updateStateFromCreated(data *v1.UpdateWorkflowRevisionAliasOKRevision) {
	if val, ok := data.RevisionAlias.Get(); ok {
		model.Alias = types.StringValue(val)
	}
}

func (model *workflowRevisionAliasBaseModel) updateStateFromRead(data *v1.GetWorkflowRevisionsOKRevision) {
	if val, ok := data.RevisionAlias.Get(); ok {
		model.Alias = types.StringValue(val)
	}
}

func (model *workflowRevisionAliasBaseModel) updateStateFromUpdated(data *v1.UpdateWorkflowRevisionAliasOKRevision) {
	if val, ok := data.RevisionAlias.Get(); ok {
		model.Alias = types.StringValue(val)
	}
}
