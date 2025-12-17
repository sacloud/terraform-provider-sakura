// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dns

import (
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type dnsRecordResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &dnsRecordResource{}
	_ resource.ResourceWithConfigure   = &dnsRecordResource{}
	_ resource.ResourceWithImportState = &dnsRecordResource{}
	_ resource.ResourceWithIdentity    = &dnsRecordResource{}
)

func NewDNSRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *dnsRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type dnsRecordResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	DNSID    types.String   `tfsdk:"dns_id"`
	Name     types.String   `tfsdk:"name"`
	Type     types.String   `tfsdk:"type"`
	Value    types.String   `tfsdk:"value"`
	TTL      types.Int64    `tfsdk:"ttl"`
	Priority types.Int32    `tfsdk:"priority"`
	Weight   types.Int32    `tfsdk:"weight"`
	Port     types.Int32    `tfsdk:"port"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *dnsRecordResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("DNS Record"),
			"dns_id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the DNS resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the DNS Record resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The type of DNS Record. This must be one of [%s]", iaastypes.DNSRecordTypeStrings),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.DNSRecordTypeStrings...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "The value of the DNS Record.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The number of the TTL.",
				Default:     int64default.StaticInt64(defaultTTL),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"priority": schema.Int32Attribute{
				Optional:    true,
				Description: desc.Sprintf("The priority of target DNS Record. %s", desc.Range(0, 65535)),
				Validators: []validator.Int32{
					int32validator.Between(0, 65535),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"weight": schema.Int32Attribute{
				Optional:    true,
				Description: desc.Sprintf("The weight of target DNS Record. %s", desc.Range(0, 65535)),
				Validators: []validator.Int32{
					int32validator.Between(0, 65535),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int32Attribute{
				Optional:    true,
				Description: desc.Sprintf("The number of port. %s", desc.Range(1, 65535)),
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Disk's record",
	}
}

type dnsRecordResourceIdentityModel struct {
	DNSID    types.String `tfsdk:"dns_id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int32  `tfsdk:"priority"`
	Weight   types.Int32  `tfsdk:"weight"`
	Port     types.Int32  `tfsdk:"port"`
}

func (r dnsRecordResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"dns_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"name": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"type": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"value": identityschema.StringAttribute{
				OptionalForImport: true,
			},
			"ttl": identityschema.Int64Attribute{
				OptionalForImport: true,
			},
			"priority": identityschema.Int32Attribute{
				OptionalForImport: true,
			},
			"weight": identityschema.Int32Attribute{
				OptionalForImport: true,
			},
			"port": identityschema.Int32Attribute{
				OptionalForImport: true,
			},
		},
	}
}

func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var identityData dnsRecordResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// To avoide timeouts related error, set each attribute directly. Don't use 'resp.State.Set'.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dns_id"), identityData.DNSID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), identityData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), identityData.Type)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("value"), identityData.Value)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ttl"), identityData.TTL)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("priority"), identityData.Priority)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("weight"), identityData.Weight)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), identityData.Port)...)
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	dnsID := plan.DNSID.ValueString()

	common.SakuraMutexKV.Lock(dnsID)
	defer common.SakuraMutexKV.Unlock(dnsID)

	dnsOp := iaas.NewDNSOp(r.client)
	dns, err := dnsOp.Read(ctx, common.SakuraCloudID(dnsID))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("could not read SakuraCloud DNS[%s]: %s", dnsID, err))
		return
	}

	record, reqSetting := expandDNSRecordCreateRequest(&plan, dns)
	_, err = dnsOp.UpdateSettings(ctx, dns.ID, reqSetting)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating Record for SakuraCloud DNS[%s] is failed: %s", dnsID, err))
		return
	}

	model := flattenDNSRecord(record)
	plan.updateState(dnsRecordIDHash(dnsID, record), dnsID, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	identity := dnsRecordResourceIdentityModel{
		DNSID:    plan.DNSID,
		Name:     plan.Name,
		Type:     plan.Type,
		Value:    plan.Value,
		TTL:      plan.TTL,
		Priority: plan.Priority,
		Weight:   plan.Weight,
		Port:     plan.Port,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dnsID := state.DNSID.ValueString()
	dns := getDNS(ctx, r.client, common.SakuraCloudID(dnsID), &resp.State, &resp.Diagnostics)
	if dns == nil {
		return
	}

	record := convertToDNSRecordModel(&state)
	if r := findRecordMatch(dns.Records, record); r == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		model := flattenDNSRecord(r)
		state.updateState(dnsRecordIDHash(dnsID, r), dnsID, &model)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

		identity := dnsRecordResourceIdentityModel{
			DNSID:    state.DNSID,
			Name:     state.Name,
			Type:     state.Type,
			Value:    state.Value,
			TTL:      state.TTL,
			Priority: state.Priority,
			Weight:   state.Weight,
			Port:     state.Port,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
	}
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Error", "updating DNS Record is not supported. To change the attributes, please delete and recreate the resource.")
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	dnsOp := iaas.NewDNSOp(r.client)
	dnsID := state.DNSID.ValueString()

	common.SakuraMutexKV.Lock(dnsID)
	defer common.SakuraMutexKV.Unlock(dnsID)

	dns := getDNS(ctx, r.client, common.SakuraCloudID(dnsID), &resp.State, &resp.Diagnostics)
	if dns == nil {
		return
	}

	_, err := dnsOp.UpdateSettings(ctx, common.SakuraCloudID(dnsID), expandDNSRecordDeleteRequest(&state, dns))
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud DNSRecord[%s] is failed: %s", dnsID, err))
		return
	}
}

func (d *dnsRecordResourceModel) updateState(id, dnsID string, model *dnsRecordModel) {
	d.ID = types.StringValue(id)
	d.DNSID = types.StringValue(dnsID)
	d.Name = model.Name
	d.Type = model.Type
	d.Value = model.Value
	d.TTL = model.TTL
	d.Priority = model.Priority
	d.Weight = model.Weight
	d.Port = model.Port
}

func expandDNSRecordCreateRequest(model *dnsRecordResourceModel, dns *iaas.DNS) (*iaas.DNSRecord, *iaas.DNSUpdateSettingsRequest) {
	record := expandDNSRecord(convertToDNSRecordModel(model))
	records := append(dns.Records, record) //nolint:gocritic

	return record, &iaas.DNSUpdateSettingsRequest{
		Records:      records,
		SettingsHash: dns.SettingsHash,
	}
}

func expandDNSRecordDeleteRequest(model *dnsRecordResourceModel, dns *iaas.DNS) *iaas.DNSUpdateSettingsRequest {
	record := expandDNSRecord(convertToDNSRecordModel(model))
	var records []*iaas.DNSRecord

	for _, r := range dns.Records {
		if !IsSameDNSRecord(r, record) {
			records = append(records, r)
		}
	}

	return &iaas.DNSUpdateSettingsRequest{
		Records:      records,
		SettingsHash: dns.SettingsHash,
	}
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

func convertToDNSRecordModel(model *dnsRecordResourceModel) *dnsRecordModel {
	return &dnsRecordModel{
		Name:     model.Name,
		Type:     model.Type,
		Value:    model.Value,
		TTL:      model.TTL,
		Priority: model.Priority,
		Weight:   model.Weight,
		Port:     model.Port,
	}
}

func findRecordMatch(records []*iaas.DNSRecord, target *dnsRecordModel) *iaas.DNSRecord {
	// TTLを指定してのImportは稀なので一番優先度を落とす。設定ファイル経由では値が指定されているのでちゃんと比較される
	for _, r := range records {
		record := expandDNSRecord(target)
		if r.Name == record.Name && r.Type == record.Type {
			switch record.Type {
			case "MX", "SRV":
				// MX, SRVはRData内にValue以外の必要なフィールドが含まれるため、変換後の値を直接比較
				if r.RData == record.RData {
					if target.TTL.IsNull() {
						return r
					} else if r.TTL == record.TTL {
						return r
					}
				}
			default:
				if target.Value.IsNull() {
					return r
				} else if r.RData == record.RData {
					if target.TTL.IsNull() {
						return r
					} else if r.TTL == record.TTL {
						return r
					}
				}
			}
		}
	}
	return nil
}

func IsSameDNSRecord(r1, r2 *iaas.DNSRecord) bool {
	return r1.Name == r2.Name && r1.RData == r2.RData && r1.TTL == r2.TTL && r1.Type == r2.Type
}

func dnsRecordIDHash(dnsID string, r *iaas.DNSRecord) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", dnsID))
	buf.WriteString(fmt.Sprintf("%s-", r.Type))
	buf.WriteString(fmt.Sprintf("%s-", r.RData))
	buf.WriteString(fmt.Sprintf("%d-", r.TTL))
	buf.WriteString(fmt.Sprintf("%s-", r.Name))
	return fmt.Sprintf("dnsrecord-%d", hashcode(buf.String()))
}

// Simulate SDK v2 hashcode
func hashcode(s string) int64 {
	v := int64(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return 0
}
