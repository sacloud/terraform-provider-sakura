// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowRevisionAliasBaseModel struct {
	WorkflowID types.String `tfsdk:"workflow_id"`
	RevisionID types.String `tfsdk:"revision_id"`
	Alias      types.String `tfsdk:"alias"`
}

func (model *workflowRevisionAliasBaseModel) toUpdateRequest() v1.UpdateWorkflowRevisionAliasReq {
	return v1.UpdateWorkflowRevisionAliasReq{
		RevisionAlias: model.Alias.ValueString(),
	}
}

type revisionData interface {
	GetWorkflowId() string
	GetRevisionId() int
	GetRevisionAlias() v1.OptString
}

func updateRevisionAliasState[T revisionData](model *workflowRevisionAliasBaseModel, data T) {
	model.WorkflowID = types.StringValue(data.GetWorkflowId())
	model.RevisionID = types.StringValue(strconv.Itoa(data.GetRevisionId()))
	if val, ok := data.GetRevisionAlias().Get(); ok {
		model.Alias = types.StringValue(val)
	}
}
