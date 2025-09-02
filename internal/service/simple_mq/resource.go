// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package simple_mq

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/simplemq-api-go"
	"github.com/sacloud/simplemq-api-go/apis/v1/queue"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/validators"
)

type simpleMQResource struct {
	client *queue.Client
}

var (
	_ resource.Resource                = &simpleMQResource{}
	_ resource.ResourceWithConfigure   = &simpleMQResource{}
	_ resource.ResourceWithImportState = &simpleMQResource{}
)

func NewSimpleMQResource() resource.Resource {
	return &simpleMQResource{}
}

func (r *simpleMQResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_mq"
}

func (r *simpleMQResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SimpleMqClient
}

type simpleMQResourceModel struct {
	simpleMqBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *simpleMQResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("SimpleMQ"),
			"description": common.SchemaResourceDescription("SimpleMQ"),
			"tags":        common.SchemaResourceTags("SimpleMQ"),
			"icon_id":     common.SchemaResourceIconID("SimpleMQ"),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SimpleMQ.",
				Validators: []validator.String{
					validators.StringFuncValidator(func(v string) error {
						return queue.QueueName(v).Validate()
					}),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"visibility_timeout_seconds": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "The duration in seconds that a message is invisible to others after being read from a queue. Default is 30 seconds.",
				Validators: []validator.Int64{
					validators.Int64FuncValidator(func(v int64) error {
						return queue.VisibilityTimeoutSeconds(v).Validate()
					}),
				},
			},
			"expire_seconds": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(345600),
				Description: "The duration in seconds that a message is stored in a queue. Default is 345600 seconds (4 days).",
				Validators: []validator.Int64{
					validators.Int64FuncValidator(func(v int64) error {
						return queue.ExpireSeconds(v).Validate()
					}),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *simpleMQResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *simpleMQResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan simpleMQResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	queueOp := simplemq.NewQueueOp(r.client)
	mq, err := queueOp.Create(ctx, expandSimpleMQCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("create SimpleMQ queue failed: %s", err))
		return
	}
	qid := simplemq.GetQueueID(mq)

	// SDK v2ではUpdateを呼び出して更新していたが、Frameworkではアクション間での状態の共有が難しいためメソッドに括り出して処理を共通化
	err = r.callUpdateRequest(ctx, qid, &plan, mq)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	q := getMessageQueue(ctx, r.client, qid, &resp.State, &resp.Diagnostics)
	if q == nil {
		return
	}

	plan.updateState(q)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *simpleMQResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state simpleMQResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mq := getMessageQueue(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if mq == nil {
		return
	}

	state.updateState(mq)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *simpleMQResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan simpleMQResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	err := r.callUpdateRequest(ctx, plan.ID.ValueString(), &plan, nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}

	q := getMessageQueue(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if q == nil {
		return
	}

	plan.updateState(q)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *simpleMQResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state simpleMQResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	queueOp := simplemq.NewQueueOp(r.client)
	mq := getMessageQueue(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if mq == nil {
		return
	}

	if err := queueOp.Delete(ctx, simplemq.GetQueueID(mq)); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("delete SimpleMQ[%s] queue failed: %s", state.ID.ValueString(), err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *simpleMQResource) callUpdateRequest(ctx context.Context, id string, plan *simpleMQResourceModel, mq *queue.CommonServiceItem) error {
	var err error
	queueOp := simplemq.NewQueueOp(r.client)

	if mq == nil {
		mq, err = queueOp.Read(ctx, id)
		if err != nil {
			return fmt.Errorf("could not read SimpleMQ[%s] queue: %w", id, err)
		}
	}

	_, err = queueOp.Config(ctx, simplemq.GetQueueID(mq), expandSimpleMQUpdateRequest(plan, mq))
	if err != nil {
		return fmt.Errorf("update SimpleMQ[%s] queue config failed: %w", id, err)
	}

	return nil
}

func getMessageQueue(ctx context.Context, client *queue.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *queue.CommonServiceItem {
	queueOp := simplemq.NewQueueOp(client)
	mq, err := queueOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get Queue Error", fmt.Sprintf("could not read SimpleMQ[%s] queue: %s", id, err))
		return nil
	}

	return mq
}

func expandSimpleMQCreateRequest(d *simpleMQResourceModel) queue.CreateQueueRequest {
	req := queue.CreateQueueRequest{
		CommonServiceItem: queue.CreateQueueRequestCommonServiceItem{
			Name: queue.QueueName(d.Name.ValueString()),
			Tags: common.TsetToStrings(d.Tags),
			Icon: queue.NewOptIcon(queue.NewIcon1Icon(queue.Icon1{
				ID: queue.NewOptIcon1ID(queue.NewStringIcon1ID(common.ExpandSakuraCloudID(d.IconID).String())),
			})),
		},
	}

	if !d.Description.IsNull() && !d.Description.IsUnknown() {
		req.CommonServiceItem.Description = queue.NewOptString(d.Description.ValueString())
	}

	return req
}

func expandSimpleMQUpdateRequest(d *simpleMQResourceModel, before *queue.CommonServiceItem) queue.ConfigQueueRequest {
	req := queue.ConfigQueueRequest{
		CommonServiceItem: queue.ConfigQueueRequestCommonServiceItem{
			Settings: before.Settings,
			Tags:     common.TsetToStrings(d.Tags),
			Icon: queue.NewOptIcon(queue.NewIcon1Icon(queue.Icon1{
				ID: queue.NewOptIcon1ID(queue.NewStringIcon1ID(common.ExpandSakuraCloudID(d.IconID).String())),
			})),
		},
	}

	if !d.VisibilityTimeoutSeconds.IsNull() && !d.VisibilityTimeoutSeconds.IsUnknown() {
		req.CommonServiceItem.Settings.VisibilityTimeoutSeconds = queue.VisibilityTimeoutSeconds(d.VisibilityTimeoutSeconds.ValueInt64())
	}
	if !d.ExpireSeconds.IsNull() && !d.ExpireSeconds.IsUnknown() {
		req.CommonServiceItem.Settings.ExpireSeconds = queue.ExpireSeconds(d.ExpireSeconds.ValueInt64())
	}
	if !d.Description.IsNull() && !d.Description.IsUnknown() {
		req.CommonServiceItem.Description = queue.NewOptString(d.Description.ValueString())
	}

	return req
}
