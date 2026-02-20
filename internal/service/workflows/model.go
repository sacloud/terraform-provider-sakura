// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowBaseModel struct {
	common.SakuraBaseModel

	SubscriptionID     types.String `tfsdk:"subscription_id"`
	Publish            types.Bool   `tfsdk:"publish"`
	Logging            types.Bool   `tfsdk:"logging"`
	ServicePrincipalID types.String `tfsdk:"service_principal_id"`
	ConcurrencyMode    types.String `tfsdk:"concurrency_mode"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`

	LatestRevision *workflowsRevisionModel `tfsdk:"latest_revision"`
}

type workflowsRevisionModel struct {
	ID        types.String `tfsdk:"id"`
	Runbook   types.String `tfsdk:"runbook"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (model *workflowBaseModel) toCreateRequest() v1.CreateWorkflowReq {
	workflowReq := v1.CreateWorkflowReq{
		Name:    model.Name.ValueString(),
		Publish: model.Publish.ValueBool(),
		Logging: model.Logging.ValueBool(),
		Tags:    workflowCreateTagsFromTf(model.Tags),
		Runbook: model.LatestRevision.Runbook.ValueString(),
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

	return workflowReq
}

func (model *workflowBaseModel) toUpdateRequest(state *workflowBaseModel) (v1.UpdateWorkflowReq, *v1.CreateWorkflowRevisionReq) {
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

	var revisionReq *v1.CreateWorkflowRevisionReq
	if model.LatestRevision.Runbook.ValueString() != state.LatestRevision.Runbook.ValueString() {
		revisionReq = &v1.CreateWorkflowRevisionReq{
			Runbook: model.LatestRevision.Runbook.ValueString(),
			// NOTE: aliasは別resourceで管理するので指定不要
		}
	}

	return req, revisionReq
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
	if len(data) == 0 {
		return fmt.Errorf("no revisions found for workflow ID: %s", model.ID.ValueString())
	}

	// NOTE: 作成順でソートしておく。最初の値がlatest
	sort.Slice(data, func(i, j int) bool {
		return data[i].CreatedAt.After(data[j].CreatedAt)
	})
	latestRevision := data[0]

	model.LatestRevision = &workflowsRevisionModel{
		ID:        types.StringValue(strconv.Itoa(latestRevision.RevisionId)),
		Runbook:   types.StringValue(latestRevision.Runbook),
		CreatedAt: types.StringValue(latestRevision.CreatedAt.String()),
		UpdatedAt: types.StringValue(latestRevision.UpdatedAt.String()),
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
