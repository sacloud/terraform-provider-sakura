// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/path"

	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwCertResource struct {
	client *v1.Client
}

func NewApigwCertResource() resource.Resource {
	return &apigwCertResource{}
}

var (
	_ resource.Resource                = &apigwCertResource{}
	_ resource.ResourceWithConfigure   = &apigwCertResource{}
	_ resource.ResourceWithImportState = &apigwCertResource{}
)

func (r *apigwCertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_cert"
}

func (r *apigwCertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwCertResourceModel struct {
	apigwCertBaseModel
	RSA      *apigwCertCertResourceModel `tfsdk:"rsa"`
	ECDSA    *apigwCertCertResourceModel `tfsdk:"ecdsa"`
	Timeouts timeouts.Value              `tfsdk:"timeouts"`
}

type apigwCertCertResourceModel struct {
	CertWO        types.String `tfsdk:"cert_wo"`
	KeyWO         types.String `tfsdk:"key_wo"`
	CertWOVersion types.Int32  `tfsdk:"cert_wo_version"`
	ExpiredAt     types.String `tfsdk:"expired_at"`
}

func (r *apigwCertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("API Gateway Certificate"),
			"name":       schemaResourceAPIGWName("API Gateway Certificate"),
			"created_at": schemaResourceAPIGWCreatedAt("API Gateway Certificate"),
			"updated_at": schemaResourceAPIGWUpdatedAt("API Gateway Certificate"),
			"rsa":        schemaResourceAPIGWCert("RSA setting for API Gateway Certificate", true),
			"ecdsa":      schemaResourceAPIGWCert("ECDSA setting for API Gateway Certificate", false),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manage an API Gateway Certificate.",
	}
}

func (r *apigwCertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if err := uuid.Validate(req.ID); err != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	}
}

func (r *apigwCertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config apigwCertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	certOp := apigw.NewCertificateOp(r.client)
	created, err := certOp.Create(ctx, expandAPIGWCertRequest(&plan, &config))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create API Gateway Certificate: %s", err))
		return
	}

	cert := getAPIGWCert(ctx, r.client, created.ID.Value.String(), string(created.Name.Value), &resp.State, &resp.Diagnostics)
	if cert == nil {
		return
	}

	updateModel(&plan, cert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwCertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apigwCertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert := getAPIGWCert(ctx, r.client, data.ID.ValueString(), data.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if cert == nil {
		return
	}

	updateModel(&data, cert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apigwCertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config apigwCertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	cert := getAPIGWCert(ctx, r.client, plan.ID.ValueString(), plan.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if cert == nil {
		return
	}

	certOp := apigw.NewCertificateOp(r.client)
	err := certOp.Update(ctx, expandAPIGWCertRequest(&plan, &config), cert.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update API Gateway Certificate[%s]: %s", cert.ID.Value.String(), err))
		return
	}

	cert = getAPIGWCert(ctx, r.client, plan.ID.ValueString(), plan.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if cert == nil {
		return
	}

	updateModel(&plan, cert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwCertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apigwCertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	cert := getAPIGWCert(ctx, r.client, state.ID.ValueString(), state.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if cert == nil {
		return
	}

	err := apigw.NewCertificateOp(r.client).Delete(ctx, cert.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete API Gateway Certificate[%s]: %s", cert.ID.Value.String(), err))
		return
	}
}

func getAPIGWCert(ctx context.Context, client *v1.Client, id string, name string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Certificate {
	certOp := apigw.NewCertificateOp(client)
	certs, err := certOp.List(ctx)
	if err != nil {
		diags.AddError("API List Error", fmt.Sprintf("failed to list API Gateway Certificates: %s", err))
		return nil
	}

	var cert *v1.Certificate
	for _, c := range certs {
		if id != "" && c.ID.Value.String() == id {
			cert = &c
			break
		}
		if name != "" && string(c.Name.Value) == name {
			cert = &c
			break
		}
	}
	if cert == nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to get API Gateway certificate[%s]", id))
		return nil
	}

	return cert
}

func expandAPIGWCertRequest(plan, config *apigwCertResourceModel) *v1.Certificate {
	c := &v1.Certificate{
		Name: v1.NewOptName(v1.Name(plan.Name.ValueString())),
		Rsa: v1.NewOptCertificateDetails(v1.CertificateDetails{
			Cert: v1.NewOptString(config.RSA.CertWO.ValueString()),
			Key:  v1.NewOptString(config.RSA.KeyWO.ValueString()),
		}),
	}
	if plan.ECDSA != nil {
		c.Ecdsa = v1.NewOptCertificateDetails(v1.CertificateDetails{
			Cert: v1.NewOptString(config.ECDSA.CertWO.ValueString()),
			Key:  v1.NewOptString(config.ECDSA.KeyWO.ValueString()),
		})
	}
	return c
}

func updateModel(model *apigwCertResourceModel, cert *v1.Certificate) {
	model.updateState(cert)
	if cert.Rsa.IsSet() {
		// import時にExpiredAtをセットするため
		if model.RSA == nil {
			model.RSA = &apigwCertCertResourceModel{}
		}
		model.RSA.ExpiredAt = types.StringValue(cert.Rsa.Value.ExpiredAt.Value.String())
	}
	if cert.Ecdsa.IsSet() {
		// ECDSAはnilではなく{}が返ってくることがありIsSetでは厳密に判定できないため、ExpiredAtで判定
		if !cert.Ecdsa.Value.ExpiredAt.Value.IsZero() {
			// import時にExpiredAtをセットするため
			if model.ECDSA == nil {
				model.ECDSA = &apigwCertCertResourceModel{}
			}
			model.ECDSA.ExpiredAt = types.StringValue(cert.Ecdsa.Value.ExpiredAt.Value.String())
		}
	}
}
