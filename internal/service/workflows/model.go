// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowBaseModel struct {
	common.SakuraBaseModel

	Publish            types.Bool   `tfsdk:"publish"`
	Logging            types.Bool   `tfsdk:"logging"`
	ServicePrincipalID types.String `tfsdk:"service_principal_id"`
	ConcurrencyMode    types.String `tfsdk:"concurrency_mode"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`

	Revisions []*workflowsRevisionModel `tfsdk:"revisions"`
}

type workflowsRevisionModel struct {
	ID        types.String `tfsdk:"id"`
	Alias     types.String `tfsdk:"alias"`
	Runbook   types.String `tfsdk:"runbook"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (model *workflowBaseModel) toCreateRequest() (v1.CreateWorkflowReq, []v1.CreateWorkflowRevisionReq, error) {
	workflowReq := v1.CreateWorkflowReq{
		Name:    model.Name.ValueString(),
		Publish: model.Publish.ValueBool(),
		Logging: model.Logging.ValueBool(),
		Tags:    workflowCreateTagsFromTf(model.Tags),
	}

	if desc := model.Description.ValueString(); desc != "" {
		workflowReq.Description = v1.NewOptString(desc)
	}

	if spID := model.ServicePrincipalID.ValueString(); spID != "" {
		workflowReq.ServicePrincipalId = v1.NewOptCreateWorkflowReqServicePrincipalId(v1.NewStringCreateWorkflowReqServicePrincipalId(spID))
	}

	if concurrencyMode := model.ConcurrencyMode.ValueString(); concurrencyMode != "" {
		workflowReq.ConcurrencyMode = v1.NewOptCreateWorkflowReqConcurrencyMode(v1.CreateWorkflowReqConcurrencyMode(concurrencyMode))
	}

	if len(model.Revisions) == 0 {
		return v1.CreateWorkflowReq{}, nil, errors.New("at least one revision is required")
	}
	// initial revision
	workflowReq.Runbook = model.Revisions[0].Runbook.ValueString()
	if revisionAlias := model.Revisions[0].Alias.ValueString(); revisionAlias != "" {
		workflowReq.RevisionAlias = v1.NewOptString(revisionAlias)
	}

	additionalRevisionsReq := make([]v1.CreateWorkflowRevisionReq, 0, len(model.Revisions)-1)
	for _, rev := range model.Revisions[1:] { // skip initial revision
		req := v1.CreateWorkflowRevisionReq{
			Runbook: rev.Runbook.ValueString(),
		}
		if revisionAlias := rev.Alias.ValueString(); revisionAlias != "" {
			req.RevisionAlias = v1.NewOptString(revisionAlias)
		}
		additionalRevisionsReq = append(additionalRevisionsReq, req)
	}

	return workflowReq, additionalRevisionsReq, nil
}

type updateRevisionAlias struct {
	ID  int
	Req v1.UpdateWorkflowRevisionAliasReq
}

func (model *workflowBaseModel) toUpdateRequest(state *workflowBaseModel) (v1.UpdateWorkflowReq, []updateRevisionAlias, []v1.CreateWorkflowRevisionReq, error) {
	req := v1.UpdateWorkflowReq{
		Name:    v1.NewOptString(model.Name.ValueString()),
		Publish: v1.NewOptBool(model.Publish.ValueBool()),
		Logging: v1.NewOptBool(model.Logging.ValueBool()),
		Tags:    workflowUpdateTagsFromTf(model.Tags),
	}

	if desc := model.Description.ValueString(); desc != "" {
		req.Description = v1.NewOptString(desc)
	}

	if concurrencyMode := model.ConcurrencyMode.ValueString(); concurrencyMode != "" {
		req.ConcurrencyMode = v1.NewOptUpdateWorkflowReqConcurrencyMode(v1.UpdateWorkflowReqConcurrencyMode(concurrencyMode))
	}

	updateRevisionsReq := make([]updateRevisionAlias, 0, len(model.Revisions))
	additionalRevisionsReq := make([]v1.CreateWorkflowRevisionReq, 0, len(model.Revisions))
	for _, rev := range model.Revisions {
		exists := false
		for _, stateRev := range state.Revisions {
			// NOTE: Runbookの重複はバリデーションで排除しているためここでは考慮しなくて良い。
			if rev.Runbook.ValueString() == stateRev.Runbook.ValueString() {
				exists = true
				// if there are changes in aliases, they need to be updated.
				if alias := rev.Alias.ValueString(); alias != stateRev.Alias.ValueString() {
					id, err := strconv.Atoi(stateRev.ID.ValueString())
					if err != nil {
						return v1.UpdateWorkflowReq{}, nil, nil, errors.New("invalid revision ID in state")
					}
					updateRevisionsReq = append(updateRevisionsReq, updateRevisionAlias{
						ID: id,
						Req: v1.UpdateWorkflowRevisionAliasReq{
							RevisionAlias: rev.Alias.ValueString(),
						},
					})
				}
				break
			}
		}
		if exists {
			continue
		}

		req := v1.CreateWorkflowRevisionReq{
			Runbook: rev.Runbook.ValueString(),
		}
		if revisionAlias := rev.Alias.ValueString(); revisionAlias != "" {
			req.RevisionAlias = v1.NewOptString(revisionAlias)
		}
		additionalRevisionsReq = append(additionalRevisionsReq, req)
	}

	return req, updateRevisionsReq, additionalRevisionsReq, nil
}

func (model *workflowBaseModel) updateStateFromCreated(data *v1.CreateWorkflowCreatedWorkflow) {
	model.UpdateBaseState(data.ID, data.Name, data.Description.Value, extractTagNames(data.Tags))
	model.Publish = types.BoolValue(data.Publish)
	model.Logging = types.BoolValue(data.Logging)
	if val, ok := data.ServicePrincipalId.Get(); ok {
		if strVal, strOk := val.GetString(); strOk {
			model.ServicePrincipalID = types.StringValue(strVal)
		} else if floatVal, floatOk := val.GetFloat64(); floatOk {
			model.ServicePrincipalID = types.StringValue(fmt.Sprintf("%.0f", floatVal))
		}
	} else {
		model.ServicePrincipalID = types.StringNull()
	}
	if val, ok := data.ConcurrencyMode.Get(); ok {
		model.ConcurrencyMode = types.StringValue(string(val))
	} else {
		model.ConcurrencyMode = types.StringNull()
	}
	model.CreatedAt = types.StringValue(data.CreatedAt.String())
	model.UpdatedAt = types.StringValue(data.UpdatedAt.String())
}

func (model *workflowBaseModel) updateRevisionsState(data []v1.ListWorkflowRevisionsOKRevisionsItem) error {
	for _, rev := range data {
		found := false
		for _, revState := range model.Revisions {
			// NOTE: 作成時stateにはrevisionのIDが未知のため、Runbookの内容でマッチングしてIDをセットするしかない。
			// 更新時はrevisionのIDがstateに存在するため、IDでマッチングする。
			if stateID := revState.ID.ValueString(); (stateID != "" && stateID == strconv.Itoa(rev.RevisionId)) || revState.Runbook.ValueString() == rev.Runbook {
				revState.ID = types.StringValue(strconv.Itoa(rev.RevisionId))
				revState.Runbook = types.StringValue(rev.Runbook)
				revState.CreatedAt = types.StringValue(rev.CreatedAt.String())
				revState.UpdatedAt = types.StringValue(rev.UpdatedAt.String())

				if alias := rev.RevisionAlias.Value; alias != "" {
					revState.Alias = types.StringValue(alias)
				} else {
					revState.Alias = types.StringNull()
				}

				found = true
			}
		}
		if !found {
			return fmt.Errorf("failed to find matching revision state for revisionID of: %d", rev.RevisionId)
		}
	}
	return nil
}

func (model *workflowBaseModel) updateStateFromUpdated(data *v1.UpdateWorkflowOKWorkflow) {
	model.UpdateBaseState(data.ID, data.Name, data.Description.Value, extractTagNames(data.Tags))
	model.Publish = types.BoolValue(data.Publish)
	model.Logging = types.BoolValue(data.Logging)
	if val, ok := data.ServicePrincipalId.Get(); ok {
		if strVal, strOk := val.GetString(); strOk {
			model.ServicePrincipalID = types.StringValue(strVal)
		} else if floatVal, floatOk := val.GetFloat64(); floatOk {
			model.ServicePrincipalID = types.StringValue(fmt.Sprintf("%.0f", floatVal))
		}
	} else {
		model.ServicePrincipalID = types.StringNull()
	}
	if val, ok := data.ConcurrencyMode.Get(); ok {
		model.ConcurrencyMode = types.StringValue(string(val))
	} else {
		model.ConcurrencyMode = types.StringNull()
	}
	model.CreatedAt = types.StringValue(data.CreatedAt.String())
	model.UpdatedAt = types.StringValue(data.UpdatedAt.String())
}

func (model *workflowBaseModel) updateState(data *v1.GetWorkflowOKWorkflow) {
	model.UpdateBaseState(data.ID, data.Name, data.Description.Value, extractTagNames(data.Tags))
	model.Publish = types.BoolValue(data.Publish)
	model.Logging = types.BoolValue(data.Logging)
	if val, ok := data.ServicePrincipalId.Get(); ok {
		if strVal, strOk := val.GetString(); strOk {
			model.ServicePrincipalID = types.StringValue(strVal)
		} else if floatVal, floatOk := val.GetFloat64(); floatOk {
			model.ServicePrincipalID = types.StringValue(fmt.Sprintf("%.0f", floatVal))
		}
	} else {
		model.ServicePrincipalID = types.StringNull()
	}
	if val, ok := data.ConcurrencyMode.Get(); ok {
		model.ConcurrencyMode = types.StringValue(string(val))
	} else {
		model.ConcurrencyMode = types.StringNull()
	}
	model.CreatedAt = types.StringValue(data.CreatedAt.String())
	model.UpdatedAt = types.StringValue(data.UpdatedAt.String())
}

func extractTagNames(tags any) []string {
	if tags == nil {
		return nil
	}
	switch t := tags.(type) {
	case []v1.CreateWorkflowCreatedWorkflowTagsItem:
		result := make([]string, 0, len(t))
		for _, tag := range t {
			result = append(result, tag.Name)
		}
		return result
	case []v1.UpdateWorkflowOKWorkflowTagsItem:
		result := make([]string, 0, len(t))
		for _, tag := range t {
			result = append(result, tag.Name)
		}
		return result
	case []v1.GetWorkflowOKWorkflowTagsItem:
		result := make([]string, 0, len(t))
		for _, tag := range t {
			result = append(result, tag.Name)
		}
		return result
	}
	return nil
}

func workflowCreateTagsFromTf(tags types.Set) []v1.CreateWorkflowReqTagsItem {
	if tags.IsNull() || tags.IsUnknown() {
		return nil
	}

	tagsSlice := common.TsetToStrings(tags)
	result := make([]v1.CreateWorkflowReqTagsItem, 0, len(tagsSlice))
	for _, tag := range tagsSlice {
		result = append(result, v1.CreateWorkflowReqTagsItem{
			Name: tag,
		})
	}
	return result
}

func workflowUpdateTagsFromTf(tags types.Set) []v1.UpdateWorkflowReqTagsItem {
	if tags.IsNull() || tags.IsUnknown() {
		return nil
	}

	tagsSlice := common.TsetToStrings(tags)
	result := make([]v1.UpdateWorkflowReqTagsItem, 0, len(tagsSlice))
	for _, tag := range tagsSlice {
		result = append(result, v1.UpdateWorkflowReqTagsItem{
			Name: tag,
		})
	}
	return result
}
