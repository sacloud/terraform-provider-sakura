// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"encoding/json"
	"errors"
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

type ddosResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &ddosResource{}
	_ resource.ResourceWithConfigure   = &ddosResource{}
	_ resource.ResourceWithImportState = &ddosResource{}
)

func NewDDoSResource() resource.Resource {
	return &ddosResource{}
}

func (r *ddosResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_ddos"
}

func (r *ddosResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureAddonClient(req, resp)
}

type ddosResourceModel struct {
	ddosBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *ddosResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":            common.SchemaResourceId("Addon DDoS"),
			"location":      schemaResourceAddonLocation("Addon DDoS"),
			"pricing_level": schemaResourceAddonPricingLevel("Addon DDoS", common.MapTo(v1.PricingLevel1.AllValues(), common.ToInt32)),
			"patterns": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "The route patterns of the Addon DDoS.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"origin": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The origin settings of the Addon DDoS.",
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
			"deployment_name": schemaResourceAddonDeploymentName("Addon DDoS"),
			"url":             schemaResourceAddonURL("Addon DDoS"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Addon DDoS.",
	}
}

func (r *ddosResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ddosResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ddosResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	op := addon.NewDDoSOp(r.client)
	result, err := op.Create(ctx, addon.DDoSCreateParams{
		Location:     plan.Location.ValueString(),
		PricingLevel: v1.PricingLevel(plan.PricingLevel.ValueInt32()),
		Patterns:     common.TlistToStrings(plan.Patterns),
		Origin: v1.FrontDoorOrigin{
			HostName:   plan.Origin.Hostname.ValueString(),
			HostHeader: plan.Origin.HostHeader.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Addon DDoS: %s", err))
		return
	}

	id, deploymentName, ok := getAddonIDsFromDeployment("DDoS", result, &resp.Diagnostics)
	if !ok {
		return
	}

	if err := waitCDNRouteDeployment(ctx, op.Read, id); err != nil {
		resp.Diagnostics.AddError("Create: Resource Error", fmt.Sprintf("failed to wait for Addon DDoS[%s] deployment ready: %s", id, err))
		return
	}

	ddos := getAddon(ctx, "DDoS", id, op.Read, &resp.State, &resp.Diagnostics)
	if ddos == nil {
		return
	}

	plan.updateState(id, deploymentName, ddos.URL.Value, &v1.DdosRequestBody{
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

func (r *ddosResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ddosResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := addon.NewDDoSOp(r.client)
	result := getAddon(ctx, "DDoS", state.ID.ValueString(), op.Read, &resp.State, &resp.Diagnostics)
	if result == nil {
		return
	}

	var body v1.DdosRequestBody
	err := decodeCDNFamilyResponse(result, &body)
	if err != nil {
		resp.Diagnostics.AddError("Read: Decode Error", fmt.Sprintf("failed to decode Addon DDoS[%s] response: %s", state.ID.ValueString(), err))
		return
	}
	state.updateState(state.ID.ValueString(), state.DeploymentName.ValueString(), result.URL.Value, &body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ddosResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update: Not Implemented Error", "Addon DDoS does not support updates")
}

func (r *ddosResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ddosResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	op := addon.NewDDoSOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		if saclient.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Addon DDoS[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func decodeDDoSResponse(resp *v1.GetResourceResponse) (v1.DdosRequestBody, error) {
	var result v1.DdosRequestBody
	if resp == nil || len(resp.Data) == 0 {
		return result, errors.New("got invalid response from Addon CDN API")
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return result, err
	}

	location, profile, err := getCDNLocationAndProfile(data)
	if err != nil {
		return result, err
	}
	endpoint, err := getFrontDoorEndpoint(data)
	if err != nil {
		return result, err
	}
	result.Location = location
	result.Profile = profile
	result.Endpoint = endpoint

	return result, nil
}
