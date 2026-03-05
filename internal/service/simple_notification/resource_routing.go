// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	simplenotification "github.com/sacloud/simple-notification-api-go"
	v1 "github.com/sacloud/simple-notification-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type routingResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &routingResource{}
	_ resource.ResourceWithConfigure   = &routingResource{}
	_ resource.ResourceWithImportState = &routingResource{}
)

func NewRoutingResource() resource.Resource {
	return &routingResource{}
}

func (r *routingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_notification_routing"
}

func (r *routingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SimpleNotificationClient
}

type routingResourceModel struct {
	routingBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *routingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	const resourceName = "SimpleNotification Routing"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			"tags":        common.SchemaResourceTags(resourceName),
			"icon_id":     common.SchemaResourceIconID(resourceName),
			"match_labels": schema.ListNestedAttribute{
				Required:    true,
				Description: desc.Sprintf("The type of the %s.", resourceName),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The key of the match label.",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "The value of the match label.",
						},
					},
				},
			},
			"source_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The value of the %s.", resourceName),
			},
			"target_group_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The value of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages SimpleNotification routing.",
	}
}

func (r *routingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *routingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	routingOp := simplenotification.NewRoutingOp(r.client)
	res, err := routingOp.Create(ctx, makeRoutingCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create SimpleNotification routing: %s", err))
		return
	}

	if err := plan.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update SimpleNotification routing[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	routingOp := simplenotification.NewRoutingOp(r.client)
	res, err := routingOp.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read SimpleNotification routing: %s", err))
		return
	}

	if err := state.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update SimpleNotification routing[%s] state: %s", state.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan routingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	routingOp := simplenotification.NewRoutingOp(r.client)
	res, err := routingOp.Update(ctx, plan.ID.ValueString(), makeRoutingUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update SimpleNotification routing[%s]: %s", plan.ID.String(), err))
		return
	}

	if err := plan.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update SimpleNotification routing[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	routingOp := simplenotification.NewRoutingOp(r.client)
	if err := routingOp.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete SimpleNotification routing[%s]: %s", state.ID.String(), err))
		return
	}
}

func makeRoutingCreateRequest(d *routingResourceModel) v1.PostCommonServiceItemRequest {
	req := v1.PostCommonServiceItemRequest{
		CommonServiceItem: v1.PostCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: d.Description.ValueString(),
			Icon: v1.NilCommonServiceItemIcon{
				Value: v1.CommonServiceItemIcon{
					ID: d.IconID.ValueString(),
				},
			},
			Settings: v1.CommonServiceItemSettings{
				RoutingSettings: v1.RoutingSettings{
					MatchLabels:   expandMatchLabels(d.MatchLabels),
					SourceID:      d.SourceID.ValueString(),
					TargetGroupID: d.TargetGroupID.ValueString(),
					PriorityRank:  1, // API requires this value, but it is not managed by this resource, so we set a init value.
				},
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}
	return req
}

func makeRoutingUpdateRequest(d *routingResourceModel) v1.PutCommonServiceItemRequest {
	req := v1.PutCommonServiceItemRequest{
		CommonServiceItem: v1.PutCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: d.Description.ValueString(),
			Icon: v1.NilCommonServiceItemIcon{
				Value: v1.CommonServiceItemIcon{
					ID: d.IconID.ValueString(),
				},
			},
			Settings: v1.OptCommonServiceItemSettings{
				Set: true,
				Value: v1.CommonServiceItemSettings{
					RoutingSettings: v1.RoutingSettings{
						MatchLabels:   expandMatchLabels(d.MatchLabels),
						SourceID:      d.SourceID.ValueString(),
						TargetGroupID: d.TargetGroupID.ValueString(),
						PriorityRank:  1, // API requires this value, but it is not managed by this resource, so we set a init value.
					},
				},
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}
	return req
}

func expandMatchLabels(models []matchLabelModel) []v1.RoutingSettingsMatchLabelsItem {
	result := make([]v1.RoutingSettingsMatchLabelsItem, len(models))
	for i, ml := range models {
		result[i] = v1.RoutingSettingsMatchLabelsItem{
			Name:  ml.Name.ValueString(),
			Value: ml.Value.ValueString(),
		}
	}
	return result
}
