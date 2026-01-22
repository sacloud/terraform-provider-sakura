// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type apigwDomainResource struct {
	client *v1.Client
}

func NewApigwDomainResource() resource.Resource {
	return &apigwDomainResource{}
}

var (
	_ resource.Resource                = &apigwDomainResource{}
	_ resource.ResourceWithConfigure   = &apigwDomainResource{}
	_ resource.ResourceWithImportState = &apigwDomainResource{}
)

func (r *apigwDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_domain"
}

func (r *apigwDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwDomainResourceModel struct {
	apigwDomainBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *apigwDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("API Gateway Domain"),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the API Gateway Domain",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\s*(([a-z\d]([a-z\d-]*[a-z\d])?)\.)+([a-z\d-]{2,})(\.)?\s*$`), "Valid domain name is required. IPv4 and wildcard are not allowed."),
				},
			},
			"created_at": schemaResourceAPIGWCreatedAt("API Gateway Domain"),
			"updated_at": schemaResourceAPIGWUpdatedAt("API Gateway Domain"),
			"certificate_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the API Gateway Certificate",
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(uuid.Validate),
				},
			},
			"certificate_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the API Gateway Certificate",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manage an API Gateway Domain.",
	}
}

func (r *apigwDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if err := uuid.Validate(req.ID); err != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	}
}

func (r *apigwDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apigwDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	domainOp := apigw.NewDomainOp(r.client)
	created, err := domainOp.Create(ctx, expandAPIGWDomainRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create API Gateway Domain: %s", err))
		return
	}

	domain := getAPIGWDomain(ctx, r.client, created.ID.Value.String(), created.DomainName, &resp.State, &resp.Diagnostics)
	if domain == nil {
		return
	}

	plan.updateState(domain)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apigwDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := getAPIGWDomain(ctx, r.client, data.ID.ValueString(), data.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if domain == nil {
		return
	}

	data.updateState(domain)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apigwDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apigwDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	domain := getAPIGWDomain(ctx, r.client, plan.ID.ValueString(), plan.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if domain == nil {
		return
	}

	updateReq := &v1.DomainPUT{}
	if utils.IsKnown(plan.CertificateId) {
		updateReq.CertificateId = v1.NewOptUUID(uuid.MustParse(plan.CertificateId.ValueString()))
	}

	domainOp := apigw.NewDomainOp(r.client)
	err := domainOp.Update(ctx, updateReq, domain.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update API Gateway Domain[%s]: %s", domain.ID.Value.String(), err))
		return
	}

	domain = getAPIGWDomain(ctx, r.client, plan.ID.ValueString(), plan.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if domain == nil {
		return
	}

	plan.updateState(domain)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apigwDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	domain := getAPIGWDomain(ctx, r.client, state.ID.ValueString(), state.Name.ValueString(), &resp.State, &resp.Diagnostics)
	if domain == nil {
		return
	}

	domainOp := apigw.NewDomainOp(r.client)
	err := domainOp.Delete(ctx, domain.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete API Gateway Domain[%s]: %s", domain.ID.Value.String(), err))
		return
	}
}

func getAPIGWDomain(ctx context.Context, client *v1.Client, id string, name string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Domain {
	domainOp := apigw.NewDomainOp(client)
	domains, err := domainOp.List(ctx)
	if err != nil {
		diags.AddError("API List Error", fmt.Sprintf("failed to list API Gateway Domains: %s", err))
		return nil
	}

	var domain *v1.Domain
	for _, d := range domains {
		if id != "" && d.ID.Value.String() == id {
			domain = &d
			break
		}
		if name != "" && d.DomainName == name {
			domain = &d
			break
		}
	}
	if domain == nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to get API Gateway Domain[%s]", id))
		return nil
	}

	return domain
}

func expandAPIGWDomainRequest(plan *apigwDomainResourceModel) *v1.Domain {
	d := &v1.Domain{
		DomainName: plan.Name.ValueString(),
	}
	if utils.IsKnown(plan.CertificateId) {
		d.CertificateId = v1.NewOptUUID(uuid.MustParse(plan.CertificateId.ValueString()))
	}
	return d
}
