// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type policyResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &policyResource{}
	_ resource.ResourceWithConfigure   = &policyResource{}
	_ resource.ResourceWithImportState = &policyResource{}
)

func NewPolicyResource() resource.Resource {
	return &policyResource{}
}

func (r *policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_policy"
}

func (r *policyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type policyResourceModel struct {
	policyBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

const (
	targetProject = "project"
	targetFolder  = "folder"
	targetOrg     = "organization"
)

func (r *policyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"target": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The target of the IAM Policy. This must be one of %s.", []string{targetProject, targetFolder, targetOrg}),
				Validators: []validator.String{
					stringvalidator.OneOf(targetProject, targetFolder, targetOrg),
				},
			},
			"target_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the target. Required for Folder or Project",
			},
			"bindings": schema.ListNestedAttribute{
				Required:    true,
				Description: "The bindings of the IAM Policy",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.SingleNestedAttribute{
							Required:    true,
							Description: "The role of the IAM Policy",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the role",
								},
								"id": schema.StringAttribute{
									Required:    true,
									Description: "The ID of the IAM Policy",
								},
							},
						},
						"principals": schema.ListNestedAttribute{
							Required:    true,
							Description: "The principals of the IAM Policy",
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
		MarkdownDescription: "Manages an IAM Policy.",
	}
}

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 2)

	if len(parts) > 2 {
		resp.Diagnostics.AddError("Import Error",
			fmt.Sprintf("invalid import ID format. Please specify the import ID in the format of {target}_{id} or {target}: %s", req.ID))
		return
	}

	switch parts[0] {
	case targetOrg:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target"), parts[0])...)
	case targetFolder, targetProject:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_id"), parts[1])...)
	default:
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("invalid target '%s'. The target must be one of 'organization', 'folder', or 'project': %s", parts[0], req.ID))
	}
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	err := updateIAMPolicy(ctx, r.client, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", err.Error())
		return
	}

	res := getIAMPolicy(ctx, r.client, plan.Target.ValueString(), plan.TargetID.ValueString(), &resp.Diagnostics)
	if res == nil {
		return
	}

	plan.updateState(plan.Target.ValueString(), res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := getIAMPolicy(ctx, r.client, state.Target.ValueString(), state.TargetID.ValueString(), &resp.Diagnostics)
	if res == nil {
		return
	}

	state.updateState(state.Target.ValueString(), res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan policyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	err := updateIAMPolicy(ctx, r.client, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", err.Error())
		return
	}

	res := getIAMPolicy(ctx, r.client, plan.Target.ValueString(), plan.TargetID.ValueString(), &resp.Diagnostics)
	if res == nil {
		return
	}

	plan.updateState(plan.Target.ValueString(), res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// do nothing
}

func getIAMPolicy(ctx context.Context, client *v1.Client, target, targetId string, diags *diag.Diagnostics) []v1.IamPolicy {
	var res []v1.IamPolicy
	var err error
	op := iam.NewIAMPolicyOp(client)
	switch target {
	case targetProject:
		res, err = op.ReadProjectPolicy(ctx, utils.MustAtoI(targetId))
		if err != nil {
			diags.AddError("API Error", fmt.Sprintf("failed to read IAM Project Policy: %s", err))
			return nil
		}
	case targetFolder:
		res, err = op.ReadFolderPolicy(ctx, utils.MustAtoI(targetId))
		if err != nil {
			diags.AddError("API Error", fmt.Sprintf("failed to read IAM Folder Policy: %s", err))
			return nil
		}
	case targetOrg:
		res, err = op.ReadOrganizationPolicy(ctx)
		if err != nil {
			diags.AddError("API Error", fmt.Sprintf("failed to read IAM Organization Policy: %s", err))
			return nil
		}
	}
	return res
}

func updateIAMPolicy(ctx context.Context, client *v1.Client, model *policyResourceModel) error {
	op := iam.NewIAMPolicyOp(client)
	switch model.Target.ValueString() {
	case targetOrg:
		_, err := op.UpdateOrganizationPolicy(ctx, expandIAMPolicyCreateRequest(model))
		if err != nil {
			return fmt.Errorf("failed to update IAM Organization Policy: %s", err)
		}
	case targetFolder:
		_, err := op.UpdateFolderPolicy(ctx, utils.MustAtoI(model.TargetID.ValueString()), expandIAMPolicyCreateRequest(model))
		if err != nil {
			return fmt.Errorf("failed to update IAM Folder Policy: %s", err)
		}
	case targetProject:
		_, err := op.UpdateProjectPolicy(ctx, utils.MustAtoI(model.TargetID.ValueString()), expandIAMPolicyCreateRequest(model))
		if err != nil {
			return fmt.Errorf("failed to update IAM Project Policy: %s", err)
		}
	}
	return nil
}

func expandIAMPolicyCreateRequest(model *policyResourceModel) []v1.IamPolicy {
	var policies []v1.IamPolicy
	for _, b := range model.Bindings {
		idp := v1.IamPolicy{
			Role: v1.NewOptIamPolicyRole(v1.IamPolicyRole{
				Type: v1.NewOptIamPolicyRoleType(v1.IamPolicyRoleType(b.Role.Type.ValueString())),
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
