// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ipv4_ptr

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type ipv4PtrResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &ipv4PtrResource{}
	_ resource.ResourceWithConfigure   = &ipv4PtrResource{}
	_ resource.ResourceWithImportState = &ipv4PtrResource{}
)

func NewIPv4PtrResource() resource.Resource {
	return &ipv4PtrResource{}
}

func (r *ipv4PtrResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv4_ptr"
}

func (r *ipv4PtrResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type ipv4PtrResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	IPAddress     types.String   `tfsdk:"ip_address"`
	Hostname      types.String   `tfsdk:"hostname"`
	RetryMax      types.Int32    `tfsdk:"retry_max"`
	RetryInterval types.Int32    `tfsdk:"retry_interval"`
	Zone          types.String   `tfsdk:"zone"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

func (r *ipv4PtrResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("IPv4 PTR"),
			"ip_address": schema.StringAttribute{
				Required:    true,
				Description: "The IP address to which the PTR record is set",
				Validators: []validator.String{
					sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
				},
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "The value of the PTR record. This must be FQDN",
				Validators: []validator.String{
					sacloudvalidator.HostnameValidator(),
				},
			},
			"retry_max": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(30),
				Description: "The maximum number of API call retries used when SakuraCloud API returns any errors",
				Validators: []validator.Int32{
					int32validator.Between(1, 100),
				},
			},
			"retry_interval": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(10),
				Description: "The wait interval(in seconds) for retrying API call used when SakuraCloud API returns any errors",
				Validators: []validator.Int32{
					int32validator.Between(1, 600),
				},
			},
			"zone": common.SchemaResourceZone("IPv4 PTR"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IPv4 PTR.",
	}
}

func (r *ipv4PtrResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ipv4PtrResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipv4PtrResourceModel
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

	if err := updateIPv4Ptr(ctx, r.client, zone, &plan); err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update IPv4 PTR[IP:%s Host:%s]: %s", plan.IPAddress.ValueString(), plan.Hostname.ValueString(), err))
		return
	}

	ptr := getIPv4Ptr(ctx, r.client, zone, plan.IPAddress.ValueString(), &resp.State, &resp.Diagnostics)
	if ptr == nil {
		return
	}
	plan.updateState(ptr, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipv4PtrResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipv4PtrResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ptr := getIPv4Ptr(ctx, r.client, zone, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if ptr == nil {
		return
	}

	state.updateState(ptr, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipv4PtrResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ipv4PtrResourceModel
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

	if err := updateIPv4Ptr(ctx, r.client, zone, &plan); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IPv4 PTR[IP:%s Host:%s]: %s", plan.IPAddress.ValueString(), plan.Hostname.ValueString(), err))
		return
	}

	ptr := getIPv4Ptr(ctx, r.client, zone, plan.IPAddress.ValueString(), &resp.State, &resp.Diagnostics)
	if ptr == nil {
		return
	}
	plan.updateState(ptr, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipv4PtrResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipv4PtrResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ip := state.ID.ValueString()
	ipAddrOp := iaas.NewIPAddressOp(r.client)
	if _, err := ipAddrOp.Read(ctx, zone, ip); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if _, err := ipAddrOp.UpdateHostName(ctx, zone, ip, ""); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to update IPv4 PTR[%s]: %s", ip, err))
		return
	}
}

func getIPv4Ptr(ctx context.Context, client *common.APIClient, zone string, ip string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.IPAddress {
	op := iaas.NewIPAddressOp(client)
	ptr, err := op.Read(ctx, zone, ip)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IPv4 PTR[%s]: %s", ip, err))
		return nil
	}
	return ptr
}

func updateIPv4Ptr(ctx context.Context, client *common.APIClient, zone string, plan *ipv4PtrResourceModel) error {
	ipAddrOp := iaas.NewIPAddressOp(client)
	ip := plan.IPAddress.ValueString()
	hostName := plan.Hostname.ValueString()
	retryMax := int(plan.RetryMax.ValueInt32())
	interval := time.Duration(plan.RetryInterval.ValueInt32()) * time.Second

	if _, err := ipAddrOp.Read(ctx, zone, ip); err != nil {
		// includes 404 error
		return fmt.Errorf("failed to find IPv4 PTR[%s]: %s", ip, err)
	}

	var err error
	i := 0
	success := false
	for i < retryMax {
		// set
		if _, err = ipAddrOp.UpdateHostName(ctx, zone, ip, hostName); err == nil {
			success = true
			break
		}

		time.Sleep(interval)
		i++
	}

	if !success {
		return fmt.Errorf("failed to update IPv4 PTR[IP:%s Host:%s]: %s", ip, hostName, err)
	}

	return nil
}

func (model *ipv4PtrResourceModel) updateState(ptr *iaas.IPAddress, zone string) {
	model.ID = types.StringValue(ptr.IPAddress)
	model.IPAddress = types.StringValue(ptr.IPAddress)
	model.Hostname = types.StringValue(ptr.HostName)
	model.Zone = types.StringValue(zone)
}
