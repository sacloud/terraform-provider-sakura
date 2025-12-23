// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package gslb

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"

	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"

	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type gslbResource struct {
	client *common.APIClient
}

func NewGSLBResource() resource.Resource {
	return &gslbResource{}
}

var (
	_ resource.Resource                = &gslbResource{}
	_ resource.ResourceWithConfigure   = &gslbResource{}
	_ resource.ResourceWithImportState = &gslbResource{}
)

func (r *gslbResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gslb"
}

func (r *gslbResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type gslbResourceModel struct {
	gslbBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *gslbResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("GSLB"),
			"name":        common.SchemaResourceName("GSLB"),
			"description": common.SchemaResourceDescription("GSLB"),
			"tags":        common.SchemaResourceTags("GSLB"),
			"icon_id":     common.SchemaResourceIconID("GSLB"),
			"fqdn": schema.StringAttribute{
				Computed:    true,
				Description: "The FQDN for accessing to the GSLB. This is typically used as value of CNAME record",
			},
			"health_check": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Health check configuration",
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Required:    true,
						Description: desc.Sprintf("The protocol used for health checks. This must be one of [%s]", iaastypes.GSLBHealthCheckProtocolStrings),
						Validators: []validator.String{
							stringvalidator.OneOf(iaastypes.GSLBHealthCheckProtocolStrings...),
						},
					},
					"delay_loop": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(10),
						Description: desc.Sprintf("The interval in seconds between checks. %s", desc.Range(10, 60)),
						Validators: []validator.Int32{
							int32validator.Between(10, 60),
						},
					},
					"host_header": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The value of host header send when checking by HTTP/HTTPS",
					},
					"path": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The path used when checking by HTTP/HTTPS",
					},
					"status": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The response-code to expect when checking by HTTP/HTTPS",
					},
					"port": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The port number used when checking by TCP/HTTP/HTTPS",
					},
				},
			},
			"weighted": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to enable weighted load-balancing",
			},
			"sorry_server": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The IP address of the SorryServer. This will be used when all servers are down",
			},
			"server": schema.ListNestedAttribute{
				Optional: true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(12),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Required:    true,
							Description: "The IP address of the server",
							Validators: []validator.String{
								sacloudvalidator.IPAddressValidator(sacloudvalidator.IPv4),
							},
						},
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "The flag to enable as destination of load balancing",
						},
						"weight": schema.Int32Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int32default.StaticInt32(1),
							Description: desc.Sprintf("The weight used when weighted load balancing is enabled. %s", desc.Range(1, 10000)),
							Validators: []validator.Int32{
								int32validator.Between(1, 10000),
							},
						},
					},
				},
			},
			"monitoring_suite": common.SchemaResourceMonitoringSuite("GSLB"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manage a GSLB.",
	}
}

func (r *gslbResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *gslbResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan gslbResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gslbOp := iaas.NewGSLBOp(r.client)
	created, err := gslbOp.Create(ctx, expandGSLBCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create GSLB: %s", err.Error()))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gslbResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gslbResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gslb := getGSLB(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if gslb == nil {
		return
	}

	state.updateState(gslb)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gslbResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gslbResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gslb := getGSLB(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if gslb == nil {
		return
	}

	gslbOp := iaas.NewGSLBOp(r.client)
	if _, err := gslbOp.Update(ctx, gslb.ID, expandGSLBUpdateRequest(&plan, gslb)); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update GSLB[%s]: %s", gslb.ID.String(), err.Error()))
		return
	}

	updated := getGSLB(ctx, r.client, gslb.ID.String(), &resp.State, &resp.Diagnostics)
	if updated == nil {
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gslbResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gslbResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gslb := getGSLB(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if gslb == nil {
		return
	}

	if err := iaas.NewGSLBOp(r.client).Delete(ctx, gslb.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete GSLB[%s]: %s", gslb.ID.String(), err.Error()))
		return
	}
}

func getGSLB(ctx context.Context, client *common.APIClient, id string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.GSLB {
	gslbOp := iaas.NewGSLBOp(client)
	gslb, err := gslbOp.Read(ctx, common.SakuraCloudID(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read GSLB[%s]: %s", id, err.Error()))
		return nil
	}
	return gslb
}

func expandGSLBCreateRequest(model *gslbResourceModel) *iaas.GSLBCreateRequest {
	return &iaas.GSLBCreateRequest{
		Name:               model.Name.ValueString(),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		HealthCheck:        expandGSLBHealthCheck(model),
		DelayLoop:          int(model.HealthCheck.DelayLoop.ValueInt32()),
		Weighted:           iaastypes.StringFlag(model.Weighted.ValueBool()),
		SorryServer:        model.SorryServer.ValueString(),
		DestinationServers: expandGSLBServers(model),
		MonitoringSuiteLog: common.ExpandMonitoringSuiteLog(model.MonitoringSuite),
	}
}

func expandGSLBUpdateRequest(model *gslbResourceModel, gslb *iaas.GSLB) *iaas.GSLBUpdateRequest {
	return &iaas.GSLBUpdateRequest{
		Name:               model.Name.ValueString(),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		HealthCheck:        expandGSLBHealthCheck(model),
		DelayLoop:          int(model.HealthCheck.DelayLoop.ValueInt32()),
		Weighted:           iaastypes.StringFlag(model.Weighted.ValueBool()),
		SorryServer:        model.SorryServer.ValueString(),
		DestinationServers: expandGSLBServers(model),
		MonitoringSuiteLog: common.ExpandMonitoringSuiteLog(model.MonitoringSuite),
		SettingsHash:       gslb.SettingsHash,
	}
}

func expandGSLBHealthCheck(model *gslbResourceModel) *iaas.GSLBHealthCheck {
	p := model.HealthCheck.Protocol.ValueString()
	switch p {
	case "http", "https":
		return &iaas.GSLBHealthCheck{
			Protocol:     iaastypes.EGSLBHealthCheckProtocol(p),
			HostHeader:   model.HealthCheck.HostHeader.ValueString(),
			Port:         iaastypes.StringNumber(int(model.HealthCheck.Port.ValueInt32())),
			Path:         model.HealthCheck.Path.ValueString(),
			ResponseCode: iaastypes.StringNumber(utils.MustAtoI(model.HealthCheck.Status.ValueString())),
		}
	case "tcp":
		return &iaas.GSLBHealthCheck{
			Protocol: iaastypes.EGSLBHealthCheckProtocol(p),
			Port:     iaastypes.StringNumber(int(model.HealthCheck.Port.ValueInt32())),
		}
	case "ping":
		return &iaas.GSLBHealthCheck{Protocol: iaastypes.EGSLBHealthCheckProtocol(p)}
	}
	return nil
}

func expandGSLBServers(model *gslbResourceModel) []*iaas.GSLBServer {
	var servers []*iaas.GSLBServer
	if len(model.Server) == 0 {
		return servers
	}

	for _, server := range model.Server {
		servers = append(servers, &iaas.GSLBServer{
			IPAddress: server.IPAddress.ValueString(),
			Enabled:   iaastypes.StringFlag(server.Enabled.ValueBool()),
			Weight:    iaastypes.StringNumber(int(server.Weight.ValueInt32())),
		})
	}
	return servers
}
