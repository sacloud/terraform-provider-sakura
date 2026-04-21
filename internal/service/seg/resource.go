// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package seg

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/saclient-go"
	seg "github.com/sacloud/service-endpoint-gateway-api-go"
	v1 "github.com/sacloud/service-endpoint-gateway-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloud_validator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type segResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &segResource{}
	_ resource.ResourceWithConfigure   = &segResource{}
	_ resource.ResourceWithImportState = &segResource{}
)

func NewSEGResource() resource.Resource {
	return &segResource{}
}

func (r *segResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_seg"
}

func (r *segResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type segResourceModel struct {
	segBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *segResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	const resourceName = "Service Endpoint Gateway"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId(resourceName),
			"zone":       common.SchemaResourceZone(resourceName),
			"vswitch_id": common.SchemaResourceVSwitchID(resourceName),
			"server_ip_addresses": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The IP server addresslist to connect the Service Endpoint Gateway",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						sacloud_validator.IPAddressValidator(sacloud_validator.IPv4),
					),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"netmask": schema.Int32Attribute{
				Required:    true,
				Description: desc.Sprintf("The bit length of the subnet to assign to the Service Endpoint Gateway. %s", desc.Range(8, 29)),
				Validators: []validator.Int32{
					int32validator.Between(8, 29),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"endpoint_setting": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The endpoint settings of the Service Endpoint Gateway",
				Attributes: map[string]schema.Attribute{
					"object_storage_endpoints": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "The list of sakura object storage endpoints to connect to the Service Endpoint Gateway",
						// validator is not added here because api is not required validation and endpoint is validated by server side,
					},
					"monitoring_suite_endpoints": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "The list of monitoring suite endpoints to connect to the Service Endpoint Gateway",
						// validator is not added here because api is not required validation and endpoint is validated by server side,
					},
					"container_registry_endpoints": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "The list of sakura container registry endpoints to connect to the Service Endpoint Gateway",
						// validator is not added here because api is not required validation and endpoint is validated by server side,
					},
					"ai_engine_endpoints": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "The list of AI engine endpoints to connect to the Service Endpoint Gateway",
						// validator is not added here because api is not required validation and endpoint is validated by server side,
					},
					"apprun_dedicated_control_enabled": schema.BoolAttribute{
						Optional:    true,
						Description: "The flag to enable AppRun Dedicated Control Plane endpoint on the Service Endpoint Gateway",
					},
				},
			},
			"monitoring_suite_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to enable monitoring suite endpoint on the Service Endpoint Gateway",
			},
			"dns_forwarding": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The DNS forwarding settings of the Service Endpoint Gateway",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Required:    true,
						Description: "The flag to enable DNS forwarding on the Service Endpoint Gateway",
					},
					"private_hosted_zone": schema.StringAttribute{
						Required:    true,
						Description: "The private hosted zone name for DNS forwarding",
					},
					"upstream_dns_1": schema.StringAttribute{
						Required:    true,
						Description: "The IP address of the first upstream DNS server for DNS forwarding",
					},
					"upstream_dns_2": schema.StringAttribute{
						Required:    true,
						Description: "The IP address of the second upstream DNS server for DNS forwarding",
					},
				},
				MarkdownDescription: "The DNS forwarding settings of the Service Endpoint Gateway. This block is required when `dns_forwarding.enabled` is `true`.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Service Endpoint Gateway.",
	}
}

func (r *segResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *segResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan segResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// check zone switch compatibility
	err := isAvailableZoneForVSwitch(ctx, r.client, zone, plan.VSwitchID, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Create: Error", fmt.Sprintf("failed to get VSwitch for Service Endpoint Gateway in zone %s: %s", plan.Zone.ValueString(), err))
		return
	}

	apiClient, err := getServiceEndpointGatewayAPIClient(r.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Client Error", fmt.Sprintf("failed to create API client for Service Endpoint Gateway in zone %s: %s", plan.Zone.ValueString(), err))
		return
	}

	segOp := seg.NewServiceEndpointGatewayOp(apiClient)
	appliance, err := createSEGAppliance(ctx, segOp, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to set up seg: %s", err))
		return
	}

	err = plan.updateState(appliance, zone)
	if err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update state for seg resource: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *segResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state segResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := getServiceEndpointGatewayAPIClient(r.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Client Error",
			fmt.Sprintf("failed to create API client for Service Endpoint Gateway in zone %s: %s", state.Zone.ValueString(), err))
		return
	}

	segOp := seg.NewServiceEndpointGatewayOp(apiClient)
	appliance, err := getSEGAppliance(ctx, segOp, state.ID, &resp.State)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read seg resource: %s", err))
		return
	}

	err = state.updateState(appliance, zone)
	if err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update state for seg resource: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *segResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan segResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := getServiceEndpointGatewayAPIClient(r.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Client Error", fmt.Sprintf("failed to create API client for Service Endpoint Gateway in zone %s: %s", plan.Zone.ValueString(), err))
		return
	}

	segOp := seg.NewServiceEndpointGatewayOp(apiClient)
	appliance, err := updateSEGAppliance(ctx, segOp, plan.ID, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update seg resource: %s", err))
		return
	}

	if err := plan.updateState(appliance, zone); err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update state for seg resource: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *segResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state segResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout60min)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := getServiceEndpointGatewayAPIClient(r.client, zone)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error",
			fmt.Sprintf("failed to create API client for Service Endpoint Gateway in zone %s: %s", state.Zone.ValueString(), err))
		return
	}

	segAPI := seg.NewServiceEndpointGatewayOp(apiClient)
	err = deleteSEGAppliance(ctx, segAPI, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to shutdown seg[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func createSEGAppliance(ctx context.Context, segAPI seg.ServiceEndpointGatewayAPI, d *segResourceModel) (*v1.ModelsApplianceAppliance, error) {
	createRequest := expandSEGCreateRequest(d)

	created, err := segAPI.Create(ctx, createRequest)
	if err != nil {
		return nil, err
	}

	instanceID := created.Appliance.ID
	err = waitForInstanceStatus(ctx, segAPI, instanceID, v1.ModelsInstanceInstanceStatusUp)
	if err != nil {
		return nil, err
	}

	return updateSEGAppliance(ctx, segAPI, types.StringValue(instanceID), d)
}

func getSEGAppliance(ctx context.Context, segAPI seg.ServiceEndpointGatewayAPI, id types.String, state *tfsdk.State) (*v1.ModelsApplianceAppliance, error) {
	seg, err := segAPI.Read(ctx, id.ValueString())

	if err != nil {
		if saclient.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil, err
		}
		return nil, err
	}

	return &seg.Appliance, nil
}

func updateSEGAppliance(ctx context.Context, segAPI seg.ServiceEndpointGatewayAPI, id types.String, d *segResourceModel) (*v1.ModelsApplianceAppliance, error) {
	updateRequest := expandSEGUpdateRequest(d)

	_, err := segAPI.Update(ctx, id.ValueString(), updateRequest)
	if err != nil {
		return nil, err
	}

	if err = segAPI.Apply(ctx, id.ValueString()); err != nil {
		return nil, err
	}

	err = waitForInstanceStatus(ctx, segAPI, id.ValueString(), v1.ModelsInstanceInstanceStatusUp)
	if err != nil {
		return nil, err
	}

	res, err := segAPI.Read(ctx, id.ValueString())
	if err != nil {
		return nil, err
	}

	return &res.Appliance, nil
}

func deleteSEGAppliance(ctx context.Context, segAPI seg.ServiceEndpointGatewayAPI, id string) error {
	err := segAPI.Shutdown(ctx, id)
	if err != nil {
		return err
	}

	err = waitForInstanceStatus(ctx, segAPI, id, v1.ModelsInstanceInstanceStatusDown)
	if err != nil {
		return err
	}

	return segAPI.Delete(ctx, id)
}
