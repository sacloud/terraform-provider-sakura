// Copyright 2016-2025 The terraform-provider-sakura Authors
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

type destinationResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &destinationResource{}
	_ resource.ResourceWithConfigure   = &destinationResource{}
	_ resource.ResourceWithImportState = &destinationResource{}
)

func NewDestinationResource() resource.Resource {
	return &destinationResource{}
}

func (r *destinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_notification_destination"
}

func (r *destinationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SimpleNotificationClient
}

type destinationResourceModel struct {
	destinationBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *destinationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	const resourceName = "SimpleNotification Destination"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			"tags":        common.SchemaResourceTags(resourceName),
			"icon_id":     common.SchemaResourceIconID(resourceName),

			"type": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The ProcessConfiguration ID of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						if err := v1.CommonServiceItemDestinationSettingsType(v).Validate(); err != nil {
							return fmt.Errorf("invalid operator: %s", v)
						}
						return nil
					}),
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The source of the %s.", resourceName),
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages SimpleNotification Destination.",
	}
}

func (r *destinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *destinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan destinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	if err := v1.CommonServiceItemDestinationSettingsType(plan.Type.ValueString()).Validate(); err != nil {
		resp.Diagnostics.AddError("Create: Validation Error", fmt.Sprintf("invalid type for SimpleNotification Destination: %s", err))
		return
	}

	destinationOp := simplenotification.NewDestinationOp(r.client)
	res, err := destinationOp.Create(ctx, makeDestinationCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create SimpleNotification Destination: %s", err))
		return
	}

	if err := plan.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update SimpleNotification Destination[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *destinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state destinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	destinationOp := simplenotification.NewDestinationOp(r.client)
	res, err := destinationOp.Read(ctx, state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read SimpleNotification Destination: %s", err))
		return
	}

	if err := state.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update SimpleNotification Destination[%s] state: %s", state.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *destinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan destinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	destinationOp := simplenotification.NewDestinationOp(r.client)

	res, err := destinationOp.Update(ctx, plan.ID.ValueString(), makeDestinationUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update SimpleNotification Destination[%s]: %s", plan.ID.String(), err))
		return
	}

	if err := plan.updateState(&res.CommonServiceItem); err != nil {
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update SimpleNotification Destination[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *destinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state destinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	destinationOp := simplenotification.NewDestinationOp(r.client)
	id := common.SakuraCloudID(state.ID.ValueString())

	if err := destinationOp.Delete(ctx, id.String()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete SimpleNotification Destination[%s]: %s", state.ID.String(), err))
		return
	}
}

func makeDestinationCreateRequest(d *destinationResourceModel) v1.PostCommonServiceItemRequest {
	req := v1.PostCommonServiceItemRequest{
		CommonServiceItem: v1.PostCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: d.Description.ValueString(),
			Icon: v1.NilCommonServiceItemIcon{
				Value: v1.CommonServiceItemIcon{
					ID: d.IconID.ValueString(),
				},
			},
			Settings: v1.PostCommonServiceItemRequestCommonServiceItemSettings{
				CommonServiceItemDestinationSettings: v1.CommonServiceItemDestinationSettings{
					Type:  v1.CommonServiceItemDestinationSettingsType(d.Type.ValueString()),
					Value: d.Value.ValueString(),
				},
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}
	return req
}

func makeDestinationUpdateRequest(d *destinationResourceModel) v1.PutCommonServiceItemRequest {
	req := v1.PutCommonServiceItemRequest{
		CommonServiceItem: v1.PutCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: d.Description.ValueString(),
			Icon: v1.NilCommonServiceItemIcon{
				Value: v1.CommonServiceItemIcon{
					ID: d.IconID.ValueString(),
				},
			},
			Settings: v1.OptPutCommonServiceItemRequestCommonServiceItemSettings{
				Set: true,
				Value: v1.PutCommonServiceItemRequestCommonServiceItemSettings{
					CommonServiceItemDestinationSettings: v1.CommonServiceItemDestinationSettings{
						Type:  v1.CommonServiceItemDestinationSettingsType(d.Type.ValueString()),
						Value: d.Value.ValueString(),
					},
				},
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}
	return req
}
