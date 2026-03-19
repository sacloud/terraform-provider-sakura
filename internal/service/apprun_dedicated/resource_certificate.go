// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cert "github.com/sacloud/apprun-dedicated-api-go/apis/certificate"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type certResource struct{ resourceClient }

type certResourceModel struct {
	certModel

	CertificatePEM             types.String   `tfsdk:"certificate_pem"`
	PrivateKeyPEM              types.String   `tfsdk:"private_key_pem"`
	IntermediateCertificatePEM types.String   `tfsdk:"intermediate_certificate_pem"`
	Timeouts                   timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ resource.Resource                = &certResource{}
	_ resource.ResourceWithConfigure   = &certResource{}
	_ resource.ResourceWithImportState = &certResource{}
)

func NewCertResource() resource.Resource { return &certResource{resourceNamed("certificate")} }

func (r *certResource) Schema(ctx context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	id := r.schemaID()

	name := r.schemaName(stringvalidator.RegexMatches(
		regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`),
		"no special characters allowed; alphanumeric and/or hyphens, dots and underscores",
	))

	clusterID := common.SchemaResourceId("cluster").(schema.StringAttribute)
	clusterID.Computed = false
	clusterID.Required = true
	clusterID.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}

	certPEM := schema.StringAttribute{
		Required:    true,
		Sensitive:   false,
		Description: "The PEM-encoded certificate",
		Validators: []validator.String{
			stringvalidator.LengthAtMost(1000000),
			stringvalidator.AlsoRequires(path.MatchRoot("private_key_pem")),
		},
	}

	keyPEM := schema.StringAttribute{
		Required:    true,
		Sensitive:   true,
		WriteOnly:   true,
		Description: "The PEM-encoded private key",
		Validators: []validator.String{
			stringvalidator.LengthAtMost(1000000),
			stringvalidator.AlsoRequires(path.MatchRoot("certificate_pem")),
		},
	}

	intermediateCertPEM := schema.StringAttribute{
		Optional:    true,
		Sensitive:   true,
		Description: "The PEM-encoded intermediate certificate",
		Validators: []validator.String{
			stringvalidator.LengthAtMost(1000000),
		},
	}

	commonName := schema.StringAttribute{
		Computed:    true,
		Description: "The common name of the certificate",
	}

	// SANはAPI上はリストで表現されている
	// が、X.509とRFC6125によると順序はない
	// Terraform上はSetであると考えるべきだろう
	subjectAlternativeNames := schema.SetAttribute{
		Computed:    true,
		ElementType: types.StringType,
		Description: "The subject alternative names of the certificate",
	}

	notBefore := schema.StringAttribute{
		Computed:    true,
		Description: "The certificate validity start time (Unix timestamp)",
	}

	notAfter := schema.StringAttribute{
		Computed:    true,
		Description: "The certificate validity end time (Unix timestamp)",
	}

	createdAt := common.SchemaResourceCreatedAt("certificate")

	updatedAt := common.SchemaResourceUpdatedAt("certificate")

	to := timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true})

	res.Schema = schema.Schema{
		Description: "Manages an AppRun dedicated certificate",
		Attributes: map[string]schema.Attribute{
			"id":                           id,
			"cluster_id":                   clusterID,
			"name":                         name,
			"certificate_pem":              certPEM,
			"private_key_pem":              keyPEM,
			"intermediate_certificate_pem": intermediateCertPEM,
			"common_name":                  commonName,
			"subject_alternative_names":    subjectAlternativeNames,
			"not_before":                   notBefore,
			"not_after":                    notAfter,
			"created_at":                   createdAt,
			"updated_at":                   updatedAt,
			"timeouts":                     to,
		},
	}
}

func (r *certResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan, config certResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	res.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if res.Diagnostics.HasError() {
		return
	}

	clusterID, err := plan.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Create: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	api := r.api(clusterID)
	created, err := api.Create(ctx, cert.CreateParams{
		Name:                       plan.Name.ValueString(),
		CertificatePEM:             plan.CertificatePEM.ValueString(),
		PrivateKeyPEM:              config.PrivateKeyPEM.ValueString(),
		IntermediateCertificatePEM: plan.IntermediateCertificatePEM.ValueStringPointer(),
	})

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create AppRun Dedicated certificate: %s", err))
		return
	}

	detail, err := api.Read(ctx, created.CertificateID)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read created AppRun Dedicated certificate: %s", err))
		return
	}

	plan.updateState(ctx, detail, clusterID)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *certResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state certResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	clusterID, err := state.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	certID, err := state.certID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Certificate ID", fmt.Sprintf("failed to parse certificate ID: %s", err))
		return
	}

	api := r.api(clusterID)

	detail, err := api.Read(ctx, certID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated certificate: %s", err))
		return
	}

	state.updateState(ctx, detail, clusterID)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (r *certResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	var plan, config certResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	res.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if res.Diagnostics.HasError() {
		return
	}

	clusterID, err := plan.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Update: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	certID, err := plan.certID()

	if err != nil {
		res.Diagnostics.AddError("Update: Invalid Certificate ID", fmt.Sprintf("failed to parse certificate ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	api := r.api(clusterID)
	err = api.Update(ctx, certID, cert.UpdateParams{
		Name:                       plan.Name.ValueString(),
		CertificatePEM:             plan.CertificatePEM.ValueString(),
		PrivateKeyPEM:              config.PrivateKeyPEM.ValueString(),
		IntermediateCertificatePEM: plan.IntermediateCertificatePEM.ValueStringPointer(),
	})

	if err != nil {
		res.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update AppRun Dedicated certificate: %s", err))
		return
	}

	detail, err := api.Read(ctx, certID)

	if err != nil {
		res.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read updated AppRun Dedicated certificate: %s", err))
		return
	}

	plan.updateState(ctx, detail, clusterID)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *certResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state certResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	clusterID, err := state.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Delete: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	certID, err := state.certID()

	if err != nil {
		res.Diagnostics.AddError("Delete: Invalid Certificate ID", fmt.Sprintf("failed to parse certificate ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	api := r.api(clusterID)

	err = api.Delete(ctx, certID)

	if err != nil {
		if saclient.IsNotFoundError(err) {
			res.State.RemoveResource(ctx)
			return
		}
		res.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete AppRun Dedicated certificate: %s", err))
		return
	}
}

func (c *certResource) api(id v1.ClusterID) cert.CertificateAPI {
	return cert.NewCertificateOp(c.client, id)
}
