// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cloudhsm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/cloudhsm-api-go"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type cloudHSMResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &cloudHSMResource{}
	_ resource.ResourceWithConfigure   = &cloudHSMResource{}
	_ resource.ResourceWithImportState = &cloudHSMResource{}
)

func NewCloudHSMResource() resource.Resource {
	return &cloudHSMResource{}
}

func (r *cloudHSMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudhsm"
}

func (r *cloudHSMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type cloudHSMResourceModel struct {
	cloudHSMBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *cloudHSMResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("CloudHSM"),
			"name":        common.SchemaResourceName("CloudHSM"),
			"description": common.SchemaResourceDescription("CloudHSM"),
			"tags":        common.SchemaResourceTags("CloudHSM"),
			"zone":        schemaResourceZone("CloudHSM"),
			"ipv4_network_address": schema.StringAttribute{
				Required:    true,
				Description: "The IPv4 network address of the CloudHSM",
			},
			"ipv4_netmask": schema.Int32Attribute{
				Required:    true,
				Description: "The IPv4 netmask of the CloudHSM",
			},
			"ipv4_address": schema.StringAttribute{
				Computed:    true,
				Description: "The IPv4 address of the CloudHSM",
			},
			"local_router": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The local router information of the CloudHSM",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The ID of the local router",
					},
					"secret_key": schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "The secret key of the local router",
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation date of the CloudHSM",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "The modification date of the CloudHSM",
			},
			"availability": schema.StringAttribute{
				Computed:    true,
				Description: "The availability status of the CloudHSM",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a CloudHSM.",
	}
}

func (r *cloudHSMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *cloudHSMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudHSMResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	cloudhsmOp := cloudhsm.NewCloudHSMOp(client)
	created, err := cloudhsmOp.Create(ctx, expandCloudHSMCreateParams(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create CloudHSM: %s", err))
		return
	}

	chsm := &v1.CloudHSM{
		ID:                 created.ID,
		Name:               created.Name,
		Description:        created.Description,
		Tags:               created.Tags,
		Ipv4NetworkAddress: created.Ipv4NetworkAddress,
		Ipv4PrefixLength:   created.Ipv4PrefixLength,
		Ipv4Address:        created.Ipv4Address,
		CreatedAt:          created.CreatedAt,
		ModifiedAt:         created.ModifiedAt,
		Availability:       created.Availability,
	}
	plan.updateState(chsm, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudHSMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudHSMResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	chsm := getCloudHSM(ctx, client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	state.updateState(chsm, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudHSMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cloudHSMResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	id := plan.ID.ValueString()
	_, err := cloudhsm.NewCloudHSMOp(client).Update(ctx, id, expandCloudHSMUpdateParams(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update CloudHSM[%s]: %s", id, err))
		return
	}

	chsm := getCloudHSM(ctx, client, id, &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	plan.updateState(chsm, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudHSMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudHSMResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	client := createClient(zone, r.client)
	chsm := getCloudHSM(ctx, client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	if err := cloudhsm.NewCloudHSMOp(client).Delete(ctx, chsm.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete CloudHSM[%s]: %s", chsm.ID, err))
		return
	}
}

func getCloudHSM(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.CloudHSM {
	chsm, err := cloudhsm.NewCloudHSMOp(client).Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read CloudHSM[%s]: %s", id, err))
		return nil
	}
	return chsm
}

func expandCloudHSMCreateParams(model *cloudHSMResourceModel) cloudhsm.CloudHSMCreateParams {
	return cloudhsm.CloudHSMCreateParams{
		Name:               model.Name.ValueString(),
		Description:        common.Ptr(model.Description.ValueString()),
		Tags:               common.TsetToStrings(model.Tags),
		Ipv4NetworkAddress: model.IPv4NetworkAddress.ValueString(),
		Ipv4PrefixLength:   int(model.IPv4Netmask.ValueInt32()),
	}
}

func expandCloudHSMUpdateParams(model *cloudHSMResourceModel) cloudhsm.CloudHSMUpdateParams {
	return cloudhsm.CloudHSMUpdateParams{
		Name:               model.Name.ValueString(),
		Description:        common.Ptr(model.Description.ValueString()),
		Tags:               common.TsetToStrings(model.Tags),
		Ipv4NetworkAddress: model.IPv4NetworkAddress.ValueString(),
		Ipv4PrefixLength:   int(model.IPv4Netmask.ValueInt32()),
	}
}
