// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dns

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type dnsResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &dnsResource{}
	_ resource.ResourceWithConfigure   = &dnsResource{}
	_ resource.ResourceWithImportState = &dnsResource{}
)

func NewDNSResource() resource.Resource {
	return &dnsResource{}
}

func (r *dnsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns"
}

func (r *dnsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type dnsResourceModel struct {
	dnsBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *dnsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("DNS"),
			"description": common.SchemaResourceDescription("DNS"),
			"tags":        common.SchemaResourceTags("DNS"),
			"icon_id":     common.SchemaResourceIconID("DNS"),
			"zone": schema.StringAttribute{
				Required:    true,
				Description: "The target zone. (e.g. `example.com`)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dns_servers": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of IP address of DNS server that manage this zone",
			},
			"record": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A list of DNS records.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(2000),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": common.SchemaResourceName("DNS"),
						"type": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The type of DNS Record. This must be one of [%s]", iaastypes.DNSRecordTypeStrings),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.DNSRecordTypeStrings...),
							},
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "The value of the DNS Record.",
						},
						"ttl": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The number of the TTL.",
							Default:     int64default.StaticInt64(defaultTTL),
						},
						"priority": schema.Int32Attribute{
							Optional:    true,
							Description: desc.Sprintf("The priority of target DNS Record. %s", desc.Range(0, 65535)),
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
						"weight": schema.Int32Attribute{
							Optional:    true,
							Description: desc.Sprintf("The weight of target DNS Record. %s", desc.Range(0, 65535)),
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
						"port": schema.Int32Attribute{
							Optional:    true,
							Description: desc.Sprintf("The number of port. %s", desc.Range(1, 65535)),
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
							},
						},
					},
				},
			},
			"monitoring_suite": common.SchemaResourceMonitoringSuite("DNS"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a DNS",
	}
}

func (r *dnsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	dnsOp := iaas.NewDNSOp(r.client)
	dns, err := dnsOp.Create(ctx, expandDNSCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud DNS is failed: %s", err))
		return
	}

	plan.updateState(dns)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dns := getDNS(ctx, r.client, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if dns == nil {
		return
	}

	state.updateState(dns)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dnsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	dnsOp := iaas.NewDNSOp(r.client)
	dns, err := dnsOp.Read(ctx, common.ExpandSakuraCloudID(state.ID))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("reading SakuraCloud DNS[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	updateReq := expandDNSUpdateRequest(&plan, &state, dns)
	_, err = dnsOp.Update(ctx, dns.ID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud DNS[%s] is failed: %s", dns.ID.String(), err))
		return
	}

	updated, err := dnsOp.Read(ctx, dns.ID)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("reading updated SakuraCloud DNS[%s] is failed: %s", dns.ID.String(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	dns := getDNS(ctx, r.client, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if dns == nil {
		return
	}

	dnsOp := iaas.NewDNSOp(r.client)
	if err := dnsOp.Delete(ctx, dns.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud DNS[%s] is failed: %s", dns.ID.String(), err))
		return
	}
}

func getDNS(ctx context.Context, client *common.APIClient, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.DNS {
	dnsOp := iaas.NewDNSOp(client)
	dns, err := dnsOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get DNS Error", fmt.Sprintf("could not read SakuraCloud DNS[%s]: %s", id.String(), err))
		return nil
	}
	return dns
}

func expandDNSCreateRequest(model *dnsResourceModel) *iaas.DNSCreateRequest {
	return &iaas.DNSCreateRequest{
		Name:               model.Zone.ValueString(),
		Description:        model.Description.ValueString(),
		Tags:               common.TsetToStrings(model.Tags),
		IconID:             common.ExpandSakuraCloudID(model.IconID),
		Records:            expandDNSRecords(model),
		MonitoringSuiteLog: common.ExpandMonitoringSuiteLog(model.MonitoringSuite),
	}
}

func expandDNSUpdateRequest(plan, state *dnsResourceModel, dns *iaas.DNS) *iaas.DNSUpdateRequest {
	records := dns.Records
	if !plan.Records.Equal(state.Records) {
		records = expandDNSRecords(plan)
	}
	return &iaas.DNSUpdateRequest{
		Description:        plan.Description.ValueString(),
		Tags:               common.TsetToStrings(plan.Tags),
		IconID:             common.ExpandSakuraCloudID(plan.IconID),
		Records:            records,
		MonitoringSuiteLog: common.ExpandMonitoringSuiteLog(plan.MonitoringSuite),
	}
}

func expandDNSRecords(model *dnsResourceModel) []*iaas.DNSRecord {
	modelRecords := make([]dnsRecordModel, 0, len(model.Records.Elements()))
	_ = model.Records.ElementsAs(context.Background(), &modelRecords, false)

	var records []*iaas.DNSRecord
	for _, rawRecord := range modelRecords {
		records = append(records, expandDNSRecord(&rawRecord))
	}
	return records
}

func expandDNSRecord(model *dnsRecordModel) *iaas.DNSRecord {
	ttl := model.TTL.ValueInt64()
	name := model.Name.ValueString()
	value := model.Value.ValueString()
	recordType := model.Type.ValueString()

	switch recordType {
	case "MX":
		pr := 10
		if !model.Priority.IsNull() && !model.Priority.IsUnknown() {
			pr = int(model.Priority.ValueInt32())
		}
		rdata := value
		if rdata != "" && !strings.HasSuffix(rdata, ".") {
			rdata += "."
		}
		return &iaas.DNSRecord{
			Name:  name,
			Type:  iaastypes.EDNSRecordType(recordType),
			RData: fmt.Sprintf("%d %s", pr, rdata),
			TTL:   int(ttl),
		}
	case "SRV":
		pr := 0
		if !model.Priority.IsNull() && !model.Priority.IsUnknown() {
			pr = int(model.Priority.ValueInt32())
		}
		weight := 0
		if !model.Weight.IsNull() && !model.Weight.IsUnknown() {
			weight = int(model.Weight.ValueInt32())
		}
		port := 1
		if !model.Port.IsNull() && !model.Port.IsUnknown() {
			port = int(model.Port.ValueInt32())
		}
		rdata := value
		if rdata != "" && !strings.HasSuffix(rdata, ".") {
			rdata += "."
		}
		return &iaas.DNSRecord{
			Name:  name,
			Type:  iaastypes.EDNSRecordType(recordType),
			RData: fmt.Sprintf("%d %d %d %s", pr, weight, port, rdata),
			TTL:   int(ttl),
		}
	default:
		return &iaas.DNSRecord{
			Name:  name,
			Type:  iaastypes.EDNSRecordType(recordType),
			RData: value,
			TTL:   int(ttl),
		}
	}
}
