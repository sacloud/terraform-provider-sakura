// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	simplenotification "github.com/sacloud/simple-notification-api-go"
	v1 "github.com/sacloud/simple-notification-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type groupResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_notification_group"
}

func (r *groupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SimpleNotificationClient
}

type groupResourceModel struct {
	groupBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *groupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	const resourceName = "SimpleNotification group"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			"tags":        common.SchemaResourceTags(resourceName),
			"destinations": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: desc.Sprintf("The ProcessConfiguration ID of the %s.", resourceName),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages SimpleNotification group.",
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	groupOp := simplenotification.NewGroupOp(r.client)
	res, err := groupOp.Create(ctx, makegroupCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create SimpleNotification group: %s", err))
		return
	}

	if err := plan.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update SimpleNotification group[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupOp := simplenotification.NewGroupOp(r.client)
	res, err := groupOp.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read SimpleNotification group: %s", err))
		return
	}

	if err := state.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update SimpleNotification group[%s] state: %s", state.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	groupOp := simplenotification.NewGroupOp(r.client)

	res, err := groupOp.Update(ctx, plan.ID.ValueString(), makegroupUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update SimpleNotification group[%s]: %s", plan.ID.String(), err))
		return
	}

	if err := plan.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update SimpleNotification group[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	groupOp := simplenotification.NewGroupOp(r.client)

	if err := groupOp.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete SimpleNotification group[%s]: %s", state.ID.String(), err))
		return
	}
}

func makegroupCreateRequest(d *groupResourceModel) v1.PostCommonServiceItemRequest {

	destinations := common.TlistToStringsOrDefault(d.Destinations)

	req := v1.PostCommonServiceItemRequest{
		CommonServiceItem: v1.PostCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: d.Description.ValueString(),
			Icon: v1.NilCommonServiceItemIcon{
				Null: true,
			},
			Settings: v1.PostCommonServiceItemRequestCommonServiceItemSettings{
				CommonServiceItemGroupSettings: v1.CommonServiceItemGroupSettings{
					Destinations: destinations,
				},
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}
	return req
}

func makegroupUpdateRequest(d *groupResourceModel) v1.PutCommonServiceItemRequest {
	destinations := common.TlistToStringsOrDefault(d.Destinations)
	req := v1.PutCommonServiceItemRequest{
		CommonServiceItem: v1.PutCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: d.Description.ValueString(),
			Icon: v1.NilCommonServiceItemIcon{
				Null: true,
			},
			Settings: v1.OptPutCommonServiceItemRequestCommonServiceItemSettings{
				Set: true,
				Value: v1.PutCommonServiceItemRequestCommonServiceItemSettings{
					CommonServiceItemGroupSettings: v1.CommonServiceItemGroupSettings{
						Destinations: destinations,
					},
				},
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}
	return req
}
