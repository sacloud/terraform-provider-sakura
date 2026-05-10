// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/webaccel-api-go"
)

type webAccelCertificateResource struct {
	client *webaccel.Client
}

var (
	_ resource.Resource                = &webAccelCertificateResource{}
	_ resource.ResourceWithConfigure   = &webAccelCertificateResource{}
	_ resource.ResourceWithImportState = &webAccelCertificateResource{}
)

func NewWebAccelCertificateResource() resource.Resource {
	return &webAccelCertificateResource{}
}

func (r *webAccelCertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webaccel_certificate"
}

func (r *webAccelCertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.WebaccelClient
}

type webAccelCertificateResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	SiteID             types.String `tfsdk:"site_id"`
	CertificateChain   types.String `tfsdk:"certificate_chain"`
	PrivateKey         types.String `tfsdk:"private_key"`
	CertificateVersion types.Int32  `tfsdk:"certificate_version"`
	SerialNumber       types.String `tfsdk:"serial_number"`
	NotBefore          types.String `tfsdk:"not_before"`
	NotAfter           types.String `tfsdk:"not_after"`
	IssuerCommonName   types.String `tfsdk:"issuer_common_name"`
	SubjectCommonName  types.String `tfsdk:"subject_common_name"`
	DNSNames           types.List   `tfsdk:"dns_names"`
	SHA256Fingerprint  types.String `tfsdk:"sha256_fingerprint"`
}

func (r *webAccelCertificateResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("WebAccel Certificate"),
			"site_id": schema.StringAttribute{
				Required:    true,
				Description: "The site ID of WebAccel.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_chain": schema.StringAttribute{
				Required:    true,
				WriteOnly:   true,
				Description: "Certificate chain in PEM format.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("certificate_version")),
				},
			},
			"private_key": schema.StringAttribute{
				Required:    true,
				WriteOnly:   true,
				Description: "Private key in PEM format.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("certificate_version")),
				},
			},
			"certificate_version": schema.Int32Attribute{
				Optional:    true,
				Description: "The version of the certificate chain/key. This value must be greater than 0 when set. Increment this when changing certificate.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
					int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("certificate_chain")),
				},
			},
			"serial_number": schema.StringAttribute{
				Computed:    true,
				Description: "Certificate serial number.",
			},
			"not_before": schema.StringAttribute{
				Computed:    true,
				Description: "Certificate validity start time (RFC3339).",
			},
			"not_after": schema.StringAttribute{
				Computed:    true,
				Description: "Certificate validity end time (RFC3339).",
			},
			"issuer_common_name": schema.StringAttribute{
				Computed:    true,
				Description: "Issuer common name.",
			},
			"subject_common_name": schema.StringAttribute{
				Computed:    true,
				Description: "Subject common name.",
			},
			"dns_names": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "DNS names included in the certificate.",
			},
			"sha256_fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "SHA256 fingerprint of the certificate.",
			},
		},
		MarkdownDescription: "Manages a WebAccel certificate.",
	}
}

func (r *webAccelCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *webAccelCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config webAccelCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := webaccel.NewOp(r.client).CreateCertificate(ctx, plan.SiteID.ValueString(), &webaccel.CreateOrUpdateCertificateRequest{
		CertificateChain: config.CertificateChain.ValueString(),
		Key:              config.PrivateKey.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create WebAccel certificate[%s]: %s", plan.SiteID.ValueString(), err))
		return
	}

	plan.updateState(res.Current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webAccelCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webAccelCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.ID.ValueString()
	certs, err := webaccel.NewOp(r.client).ReadCertificate(ctx, siteID)
	if err != nil {
		if webaccel.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read WebAccel certificate[%s]: %s", siteID, err))
		return
	}

	if certs.Current == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.updateState(certs.Current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webAccelCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config webAccelCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.CertificateVersion.ValueInt32() > state.CertificateVersion.ValueInt32() {
		siteID := state.ID.ValueString()
		res, err := webaccel.NewOp(r.client).UpdateCertificate(ctx, siteID, &webaccel.CreateOrUpdateCertificateRequest{
			CertificateChain: config.CertificateChain.ValueString(),
			Key:              config.PrivateKey.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update WebAccel certificate[%s]: %s", siteID, err))
			return
		}
		plan.updateState(res.Current)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webAccelCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webAccelCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.ID.ValueString()
	if err := webaccel.NewOp(r.client).DeleteCertificate(ctx, siteID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete WebAccel certificate[%s]: %s", siteID, err))
		return
	}
}

func (m *webAccelCertificateResourceModel) updateState(data *webaccel.CurrentCertificate) {
	if data == nil {
		return
	}

	notBefore := time.Unix(data.NotBefore/1000, 0).Format(time.RFC3339)
	notAfter := time.Unix(data.NotAfter/1000, 0).Format(time.RFC3339)

	m.ID = types.StringValue(data.SiteID)
	m.SiteID = types.StringValue(data.SiteID)
	m.SerialNumber = types.StringValue(data.SerialNumber)
	m.NotBefore = types.StringValue(notBefore)
	m.NotAfter = types.StringValue(notAfter)
	m.IssuerCommonName = types.StringValue(data.Issuer.CommonName)
	m.SubjectCommonName = types.StringValue(data.Subject.CommonName)
	m.DNSNames = common.StringsToTlist(data.DNSNames)
	m.SHA256Fingerprint = types.StringValue(data.SHA256Fingerprint)
}
