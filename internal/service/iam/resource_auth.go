// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type authResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &authResource{}
	_ resource.ResourceWithConfigure   = &authResource{}
	_ resource.ResourceWithImportState = &authResource{}
)

func NewAuthResource() resource.Resource {
	return &authResource{}
}

func (r *authResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_auth"
}

func (r *authResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type authResourceModel struct {
	authBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *authResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"password_policy": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"min_length": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(8),
						Description: "The minimum length of the password.",
						Validators: []validator.Int32{
							int32validator.AtLeast(8),
						},
					},
					"require_uppercase": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether to require uppercase letters in the password.",
					},
					"require_lowercase": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether to require lowercase letters in the password.",
					},
					"require_symbols": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether to require symbols in the password.",
					},
				},
			},
			"conditions": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"ip_restriction": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"mode": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "The mode of IP restriction.",
								Validators: []validator.String{
									stringvalidator.OneOf("allow_all", "allow_list"),
								},
							},
							"source_network": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
								Description: "The source networks for IP restriction.",
							},
						},
					},
					"require_two_factor_auth": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether to require two-factor authentication.",
					},
					"datetime_restriction": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"after": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "The start time for datetime restriction.",
								Validators: []validator.String{
									sacloudvalidator.StringFuncValidator(func(v string) error {
										if _, err := time.Parse(time.RFC3339Nano, v); err != nil {
											return fmt.Errorf("invalid datetime format of after: %s", err)
										}
										return nil
									}),
								},
							},
							"before": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "The end time for datetime restriction.",
								Validators: []validator.String{
									sacloudvalidator.StringFuncValidator(func(v string) error {
										if _, err := time.Parse(time.RFC3339Nano, v); err != nil {
											return fmt.Errorf("invalid datetime format of before: %s", err)
										}
										return nil
									}),
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
		MarkdownDescription: "Manages an IAM Auth.",
	}
}

func (r *authResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// APIがシングルトンなためパラメータは必要ないが、respを触らないとエラーになるので意味のない処理を入れておく
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("password_policy"), types.ObjectNull(authPasswordModel{}.AttributeTypes()))...)
}

func (r *authResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan authResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	var err error
	authOp := iam.NewAuthOp(r.client)
	if utils.IsKnown(plan.PasswordPolicy) {
		_, err = authOp.UpdatePasswordPolicy(ctx, *expandPasswordPolicyRequest(&plan))
		if err != nil {
			resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update IAM Auth Password Policy: %s", err))
			return
		}
	}
	if utils.IsKnown(plan.Conditions) {
		_, err = authOp.UpdateAuthConditions(ctx, expandAuthConditionsRequest(&plan))
		if err != nil {
			resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update IAM Auth Conditions: %s", err))
			return
		}
	}

	ppRes, acRes := getAuth(ctx, r.client, &resp.State, &resp.Diagnostics)
	if ppRes == nil && acRes == nil {
		return
	}

	plan.updateState(ppRes, acRes)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *authResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state authResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppRes, acRes := getAuth(ctx, r.client, &resp.State, &resp.Diagnostics)
	if ppRes == nil && acRes == nil {
		return
	}

	state.updateState(ppRes, acRes)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *authResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan authResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	var err error
	authOp := iam.NewAuthOp(r.client)
	if utils.IsKnown(plan.PasswordPolicy) {
		_, err = authOp.UpdatePasswordPolicy(ctx, *expandPasswordPolicyRequest(&plan))
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM Auth Password Policy: %s", err))
			return
		}
	}
	if utils.IsKnown(plan.Conditions) {
		_, err = authOp.UpdateAuthConditions(ctx, expandAuthConditionsRequest(&plan))
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM Auth Conditions: %s", err))
			return
		}
	}

	ppRes, acRes := getAuth(ctx, r.client, &resp.State, &resp.Diagnostics)
	if ppRes == nil && acRes == nil {
		return
	}

	plan.updateState(ppRes, acRes)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *authResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// do nothing
}

func getAuth(ctx context.Context, client *v1.Client, state *tfsdk.State, diags *diag.Diagnostics) (*v1.PasswordPolicy, *v1.AuthConditions) {
	authOp := iam.NewAuthOp(client)
	ppRes, err := authOp.ReadPasswordPolicy(ctx)
	if err != nil {
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM Auth Password Policy: %s", err.Error()))
		return nil, nil
	}
	acRes, err := authOp.ReadAuthConditions(ctx)
	if err != nil {
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM Auth Conditions: %s", err.Error()))
		return nil, nil
	}
	return ppRes, acRes
}

func expandPasswordPolicyRequest(model *authResourceModel) *v1.PasswordPolicy {
	var passwordPolicy authPasswordModel
	_ = model.PasswordPolicy.As(context.Background(), &passwordPolicy, basetypes.ObjectAsOptions{})

	return &v1.PasswordPolicy{
		MinLength:        int(passwordPolicy.MinLength.ValueInt32()),
		RequireUppercase: passwordPolicy.RequireUppercase.ValueBool(),
		RequireLowercase: passwordPolicy.RequireLowercase.ValueBool(),
		RequireSymbols:   passwordPolicy.RequireSymbols.ValueBool(),
	}
}

func expandAuthConditionsRequest(model *authResourceModel) *v1.AuthConditions {
	var conditions authConditionsModel
	_ = model.Conditions.As(context.Background(), &conditions, basetypes.ObjectAsOptions{})

	params := v1.AuthConditions{
		RequireTwoFactorAuth: v1.AuthConditionsRequireTwoFactorAuth{Enabled: conditions.RequireTwoFactorAuth.ValueBool()},
	}

	if utils.IsKnown(conditions.IPRestriction) {
		var ipr authIPRestrictionModel
		_ = conditions.IPRestriction.As(context.Background(), &ipr, basetypes.ObjectAsOptions{})

		ipRes := v1.AuthConditionsIPRestriction{}
		switch ipr.Mode.ValueString() {
		case "allow_all":
			ipRes.OneOf = v1.AuthConditionsIPRestrictionSum{
				Type: v1.AuthConditionsIPRestrictionSum0AuthConditionsIPRestrictionSum,
				AuthConditionsIPRestrictionSum0: v1.AuthConditionsIPRestrictionSum0{
					Mode: v1.NewOptAuthConditionsIPRestrictionSum0Mode(v1.AuthConditionsIPRestrictionSum0Mode(ipr.Mode.ValueString())),
					//Mode: v1.NewOptAuthConditionsIPRestrictionSum0Mode(v1.AuthConditionsIPRestrictionSum0Mode("allow_all")),
				},
			}
		case "allow_list":
			ipRes.OneOf = v1.AuthConditionsIPRestrictionSum{
				Type: v1.AuthConditionsIPRestrictionSum1AuthConditionsIPRestrictionSum,
				AuthConditionsIPRestrictionSum1: v1.AuthConditionsIPRestrictionSum1{
					Mode:          v1.NewOptAuthConditionsIPRestrictionSum1Mode(v1.AuthConditionsIPRestrictionSum1Mode(ipr.Mode.ValueString())),
					SourceNetwork: common.TlistToStrings(ipr.SourceNetwork),
				},
			}
		}
		params.IPRestriction = ipRes
	}
	if utils.IsKnown(conditions.DatetimeRestriction) {
		var dtr authDatetimeRestrictionModel
		_ = conditions.DatetimeRestriction.As(context.Background(), &dtr, basetypes.ObjectAsOptions{})

		dt := v1.AuthConditionsDatetimeRestriction{}
		if utils.IsKnown(dtr.After) {
			v, _ := time.Parse(time.RFC3339Nano, dtr.After.ValueString())
			dt.After = v1.NewNilDateTime(v)
		}
		if utils.IsKnown(dtr.Before) {
			v, _ := time.Parse(time.RFC3339Nano, dtr.Before.ValueString())
			dt.Before = v1.NewNilDateTime(v)
		}
		params.DatetimeRestriction = dt
	}

	return &params
}
