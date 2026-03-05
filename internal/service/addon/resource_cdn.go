// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/addon-api-go"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type cdnResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &cdnResource{}
	_ resource.ResourceWithConfigure   = &cdnResource{}
	_ resource.ResourceWithImportState = &cdnResource{}
)

func NewCDNResource() resource.Resource {
	return &cdnResource{}
}

func (r *cdnResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_cdn"
}

func (r *cdnResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureAddonClient(req, resp)
}

type cdnResourceModel struct {
	cdnBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *cdnResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("Addon CDN"),
			"location": schema.StringAttribute{
				Required:    true,
				Description: "The location of the Addon CDN.",
			},
			"pricing_level": schemaResourceAddonPricingLevel("Addon CDN", common.MapTo(v1.PricingLevel1.AllValues(), common.ToInt32)),
			"patterns": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "The route patterns of the Addon CDN.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"origin": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The origin settings of the Addon CDN.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"hostname": schema.StringAttribute{
						Required:    true,
						Description: "The origin host name.",
					},
					"host_header": schema.StringAttribute{
						Required:    true,
						Description: "The origin host header.",
					},
				},
			},
			"deployment_name": schemaResourceAddonDeploymentName("Addon CDN"),
			"url":             schemaResourceAddonURL("Addon CDN"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Addon CDN.",
	}
}

func (r *cdnResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *cdnResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cdnResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	op := addon.NewCDNOp(r.client)
	result, err := op.Create(ctx, addon.CDNCreateParams{
		Location:     plan.Location.ValueString(),
		PricingLevel: v1.PricingLevel(plan.PricingLevel.ValueInt32()),
		Patterns:     common.TlistToStrings(plan.Patterns),
		Origin: v1.FrontDoorOrigin{
			HostName:   plan.Origin.Hostname.ValueString(),
			HostHeader: plan.Origin.HostHeader.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Addon CDN: %s", err))
		return
	}

	id, deploymentName, ok := getAddonIDsFromDeployment("CDN", result, &resp.Diagnostics)
	if !ok {
		return
	}

	if err := waitFrontDoorDeployment(ctx, op.Read, id); err != nil {
		resp.Diagnostics.AddError("Create: Resource Error", fmt.Sprintf("failed to wait for Addon FrontDoor[%s] deployment ready: %s", id, err))
		return
	}

	cdn := getAddon(ctx, "CDN", id, op.Read, &resp.State, &resp.Diagnostics)
	if cdn == nil {
		return
	}

	plan.updateState(id, deploymentName, cdn.URL.Value, &v1.NetworkRequestBody{
		Location: plan.Location.ValueString(),
		Profile: v1.FrontDoorProfile{
			Level: v1.PricingLevel(plan.PricingLevel.ValueInt32()),
		},
		Endpoint: v1.FrontDoorEndpoint{
			Route: v1.FrontDoorRoute{
				Patterns: common.TlistToStrings(plan.Patterns),
				OriginGroup: v1.FrontDoorOriginGroup{
					Origin: v1.FrontDoorOrigin{
						HostName:   plan.Origin.Hostname.ValueString(),
						HostHeader: plan.Origin.HostHeader.ValueString(),
					},
				},
			},
		},
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cdnResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cdnResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := addon.NewCDNOp(r.client)
	result := getAddon(ctx, "CDN", state.ID.ValueString(), op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	var body v1.NetworkRequestBody
	err := decodeFrontDoorFamilyResponse(result, &body)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon CDN[%s] response: %s", state.ID.ValueString(), err))
		return
	}
	state.updateState(state.ID.ValueString(), state.DeploymentName.ValueString(), result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cdnResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Addon CDN does not support updates")
}

func (r *cdnResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cdnResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	op := addon.NewCDNOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Addon CDN[%s]: %s", state.ID.ValueString(), err))
		return
	}
}
