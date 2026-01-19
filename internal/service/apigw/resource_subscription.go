// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type apigwSubscriptionResource struct {
	client *v1.Client
}

func NewApigwSubscriptionResource() resource.Resource {
	return &apigwSubscriptionResource{}
}

var (
	_ resource.Resource                = &apigwSubscriptionResource{}
	_ resource.ResourceWithConfigure   = &apigwSubscriptionResource{}
	_ resource.ResourceWithImportState = &apigwSubscriptionResource{}
)

func (r *apigwSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_subscription"
}

func (r *apigwSubscriptionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwSubscriptionResourceModel struct {
	apigwSubscriptionBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *apigwSubscriptionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("API Gateway Subscription"),
			"name":       schemaResourceAPIGWName("API Gateway Subscription"),
			"created_at": schemaResourceAPIGWCreatedAt("API Gateway Subscription"),
			"updated_at": schemaResourceAPIGWUpdatedAt("API Gateway Subscription"),
			"plan_id": schema.StringAttribute{
				Required:    true,
				Description: "Plan ID of the API Gateway Subscription",
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						return uuid.Validate(v)
					}),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Resource ID of the API Gateway Subscription",
			},
			"monthly_request": schema.Int64Attribute{
				Computed:    true,
				Description: "Monthly request count of the API Gateway Subscription",
			},
			"service": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Service information of the API Gateway Subscription",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "ID of the API Gateway Service associated with the API Gateway Subscription",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "Name of the API Gateway Service associated with the API Gateway Subscription",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manage an API Gateway Subscription.",
	}
}

func (r *apigwSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *apigwSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apigwSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	subOp := apigw.NewSubscriptionOp(r.client)
	err := subOp.Create(ctx, uuid.MustParse(plan.PlanID.ValueString()), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create API Gateway Subscription: %s", err))
		return
	}

	id := getAPIGWSubscriptionId(ctx, r.client, plan.Name.ValueString(), &resp.Diagnostics)
	if id == "" {
		return
	}

	sub := getAPIGWSubscriptionFromList(ctx, r.client, id, &resp.State, &resp.Diagnostics)
	if sub == nil {
		return
	}

	plan.updateState(sub)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apigwSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub := getAPIGWSubscriptionFromList(ctx, r.client, data.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sub == nil {
		return
	}

	data.updateState(sub)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apigwSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apigwSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	sub := getAPIGWSubscriptionFromList(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sub == nil {
		return
	}

	subOp := apigw.NewSubscriptionOp(r.client)
	err := subOp.Update(ctx, sub.ID.Value, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update API Gateway Subscription[%s]: %s", sub.ID.Value.String(), err))
		return
	}

	sub = getAPIGWSubscriptionFromList(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sub == nil {
		return
	}

	plan.updateState(sub)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apigwSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	sub := getAPIGWSubscriptionFromList(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sub == nil {
		return
	}

	subOp := apigw.NewSubscriptionOp(r.client)
	err := subOp.Delete(ctx, sub.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete API Gateway Subscription[%s]: %s", sub.ID.Value.String(), err))
		return
	}
}

// 現状Readが壊れているためListを使う
func getAPIGWSubscriptionFromList(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Subscription {
	subOp := apigw.NewSubscriptionOp(client)
	subs, err := subOp.List(ctx)
	if err != nil {
		diags.AddError("API List Error", fmt.Sprintf("failed to list API Gateway subscriptions: %s", err))
		return nil
	}

	for _, s := range subs {
		if string(s.ID.Value.String()) == id {
			return &s
		}
	}

	diags.AddError("Search Error", fmt.Sprintf("failed to find API Gateway subscription by ID: %s", id))
	return nil
}

/* Readが壊れているため一旦コメントアウト
func getAPIGWSubscription(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.SubscriptionDetailResponse {
	subOp := apigw.NewSubscriptionOp(client)
	sub, err := subOp.Read(ctx, uuid.MustParse(id))
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read API Gateway subscription[%s]: %s", id, err))
		return nil
	}

	return sub
}
*/

func getAPIGWSubscriptionId(ctx context.Context, client *v1.Client, name string, diags *diag.Diagnostics) string {
	subOp := apigw.NewSubscriptionOp(client)
	subs, err := subOp.List(ctx)
	if err != nil {
		diags.AddError("API List Error", fmt.Sprintf("failed to list API Gateway subscriptions: %s", err))
		return ""
	}

	for _, s := range subs {
		if string(s.Name.Value) == name {
			return s.ID.Value.String()
		}
	}

	diags.AddError("Search Error", fmt.Sprintf("failed to find API Gateway subscription by name: %s", name))
	return ""
}
