// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type orgIDPolicyResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &orgIDPolicyResource{}
	_ resource.ResourceWithConfigure   = &orgIDPolicyResource{}
	_ resource.ResourceWithImportState = &orgIDPolicyResource{}
)

func NewOrgIDPolicyResource() resource.Resource {
	return &orgIDPolicyResource{}
}

func (r *orgIDPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_organization_id_policy"
}

func (r *orgIDPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type orgIDPolicyResourceModel struct {
	idPolicyBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *orgIDPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bindings": schema.ListNestedAttribute{
				Required:    true,
				Description: "The bindings of the IAM Organization ID Policy",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.SingleNestedAttribute{
							Required:    true,
							Description: "The role of the IAM Organization ID Policy",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the role",
								},
								"id": schema.StringAttribute{
									Required:    true,
									Description: "The ID of the IAM Organization ID Policy",
								},
							},
						},
						"principals": schema.ListNestedAttribute{
							Required:    true,
							Description: "The principals of the IAM Organization ID Policy",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required:    true,
										Description: "The type of the principal",
									},
									"id": schema.StringAttribute{
										Required:    true,
										Description: "The ID of the principal",
									},
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IAM Organization ID Policy.",
	}
}

func (r *orgIDPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// APIがシングルトンなためパラメータは必要ないが、respを触らないとエラーになるので意味のない処理を入れておく
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bindings"), []idPolicyBindingModel{})...)
}

func (r *orgIDPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan orgIDPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	idPolicyOp := iam.NewIDPolicyOp(r.client)
	res, err := idPolicyOp.UpdateOrganizationIdPolicy(ctx, expandIdPolicyCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update IAM Organization ID Policy: %s", err))
		return
	}

	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *orgIDPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state orgIDPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := getOrgIDPolicy(ctx, r.client, &resp.Diagnostics)
	if res == nil {
		return
	}

	state.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *orgIDPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan orgIDPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	idPolicyOp := iam.NewIDPolicyOp(r.client)
	_, err := idPolicyOp.UpdateOrganizationIdPolicy(ctx, expandIdPolicyCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM Organization ID Policy: %s", err))
		return
	}

	res := getOrgIDPolicy(ctx, r.client, &resp.Diagnostics)
	if res == nil {
		return
	}

	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *orgIDPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// do nothing
}

func getOrgIDPolicy(ctx context.Context, client *v1.Client, diags *diag.Diagnostics) []v1.IdPolicy {
	op := iam.NewIDPolicyOp(client)
	res, err := op.ReadOrganizationIdPolicy(ctx)
	if err != nil {
		diags.AddError("API Error", fmt.Sprintf("failed to read IAM Organization ID Policy: %s", err))
		return nil
	}
	return res
}

func expandIdPolicyCreateRequest(model *orgIDPolicyResourceModel) []v1.IdPolicy {
	var policies []v1.IdPolicy
	for _, b := range model.Bindings {
		idp := v1.IdPolicy{
			Role: v1.NewOptIdPolicyRole(v1.IdPolicyRole{
				Type: v1.NewOptIdPolicyRoleType(v1.IdPolicyRoleType(b.Role.Type.ValueString())),
				ID:   v1.NewOptString(b.Role.ID.ValueString()),
			}),
		}
		for _, p := range b.Principals {
			idp.Principals = append(idp.Principals, v1.Principal{
				Type: v1.NewOptString(p.Type.ValueString()),
				ID:   v1.NewOptInt(utils.MustAtoI(p.ID.ValueString())),
			})
		}
		policies = append(policies, idp)
	}
	return policies
}
