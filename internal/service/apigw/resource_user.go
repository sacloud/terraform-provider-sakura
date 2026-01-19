// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/path"

	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type apigwUserResource struct {
	client *v1.Client
}

func NewApigwUserResource() resource.Resource {
	return &apigwUserResource{}
}

var (
	_ resource.Resource                = &apigwUserResource{}
	_ resource.ResourceWithConfigure   = &apigwUserResource{}
	_ resource.ResourceWithImportState = &apigwUserResource{}
)

func (r *apigwUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_user"
}

func (r *apigwUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwUserResourceModel struct {
	apigwUserBaseModel
	Authentication *apigwUserAuthenticationResourceModel `tfsdk:"authentication"`
	Timeouts       timeouts.Value                        `tfsdk:"timeouts"`
}

type apigwUserAuthenticationResourceModel struct {
	BasicAuth *apigwUserAuthenticationBasicAuthResourceModel `tfsdk:"basic_auth"`
	JWT       *apigwUserAuthenticationJWTResourceModel       `tfsdk:"jwt"`
	HMACAuth  *apigwUserAuthenticationHMACAuthResourceModel  `tfsdk:"hmac_auth"`
}

type apigwUserAuthenticationBasicAuthResourceModel struct {
	Username          types.String `tfsdk:"username"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int32  `tfsdk:"password_wo_version"`
}

type apigwUserAuthenticationJWTResourceModel struct {
	Key             types.String `tfsdk:"key"`
	SecretWO        types.String `tfsdk:"secret_wo"`
	SecretWOVersion types.Int32  `tfsdk:"secret_wo_version"`
	Algorithm       types.String `tfsdk:"algorithm"`
}

type apigwUserAuthenticationHMACAuthResourceModel struct {
	Username        types.String `tfsdk:"username"`
	SecretWO        types.String `tfsdk:"secret_wo"`
	SecretWOVersion types.Int32  `tfsdk:"secret_wo_version"`
}

func (r *apigwUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("API Gateway User"),
			"name":       common.SchemaResourceName("API Gateway User"),
			"tags":       common.SchemaResourceTags("API Gateway User"),
			"created_at": schemaResourceAPIGWCreatedAt("API Gateway User"),
			"updated_at": schemaResourceAPIGWUpdatedAt("API Gateway User"),
			"custom_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The custom ID of the API Gateway User",
			},
			"ip_restriction": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "IP restriction configuration for the user",
				Attributes: map[string]schema.Attribute{
					"protocols": schema.StringAttribute{
						Required:    true,
						Description: "The protocols to restrict.",
					},
					"restricted_by": schema.StringAttribute{
						Required:    true,
						Description: "The category to restrict by.",
					},
					"ips": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The IPv4 addresses to be restricted.",
					},
				},
			},
			"groups": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Groups associated with the user",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "ID of the API Gateway Group",
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Name of the API Gateway Group",
						},
					},
				},
			},
			"authentication": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Authentication information of the API Gateway User",
				Attributes: map[string]schema.Attribute{
					"basic_auth": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "The BASIC auth",
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Required:    true,
								Description: "The basic auth username",
							},
							"password_wo": schema.StringAttribute{
								Required:    true,
								WriteOnly:   true,
								Description: "The basic auth password",
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
								},
							},
							"password_wo_version": schema.Int32Attribute{
								Required:    true,
								Description: "The version of the password_wo. This value must be greater than 0 when set. Increment this when changing password.",
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
									int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo")),
								},
							},
						},
					},
					"jwt": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "The JWT auth",
						Attributes: map[string]schema.Attribute{
							"key": schema.StringAttribute{
								Required:    true,
								Description: "The JWT key",
							},
							"secret_wo": schema.StringAttribute{
								Required:    true,
								WriteOnly:   true,
								Description: "The JWT secret",
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_wo_version")),
								},
							},
							"secret_wo_version": schema.Int32Attribute{
								Required:    true,
								Description: "The version of the secret_wo. This value must be greater than 0 when set. Increment this when changing secret.",
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
									int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_wo")),
								},
							},
							"algorithm": schema.StringAttribute{
								Required:    true,
								Description: "The JWT algorithm",
							},
						},
					},
					"hmac_auth": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "The HMAC auth",
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Required:    true,
								Description: "The HMAC auth username",
							},
							"secret_wo": schema.StringAttribute{
								Required:    true,
								WriteOnly:   true,
								Description: "The HMAC auth secret",
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_wo_version")),
								},
							},
							"secret_wo_version": schema.Int32Attribute{
								Required:    true,
								Description: "The version of the secret_wo. This value must be greater than 0 when set. Increment this when changing secret.",
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
									int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("secret_wo")),
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
		MarkdownDescription: "Manage an API Gateway User.",
	}
}

func (r *apigwUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *apigwUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config apigwUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	userOp := apigw.NewUserOp(r.client)
	created, err := userOp.Create(ctx, expandAPIGWUserRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create API Gateway User: %s", err))
		return
	}
	err = updateUserExtra(ctx, r.client, created.ID.Value, &plan, &config)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to set extra settings for API Gateway User[%s]: %s. Remove the resource and try again", created.ID.Value.String(), err))
		return
	}

	user := getAPIGWUser(ctx, r.client, created.ID.Value.String(), &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}

	plan.updateState(user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apigwUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := getAPIGWUser(ctx, r.client, data.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}
	auth, err := apigw.NewUserExtraOp(r.client, user.ID.Value).ReadAuth(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read authentication settings for API Gateway User[%s]: %s", user.ID.Value.String(), err))
		return
	}

	data.updateState(user)
	data.Authentication = flattenAPIGWUserAuthenticationResource(data.Authentication, auth)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apigwUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config apigwUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	user := getAPIGWUser(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}

	userOp := apigw.NewUserOp(r.client)
	err := userOp.Update(ctx, expandAPIGWUserRequest(&plan), user.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update API Gateway service: %s", err))
		return
	}
	err = updateUserExtra(ctx, r.client, user.ID.Value, &plan, &config)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update extra settings for API Gateway User[%s]: %s.", user.ID.Value.String(), err))
		return
	}
	err = removeGroups(ctx, r.client, user.ID.Value, plan.Groups, user.Groups)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", err.Error())
		return
	}

	user = getAPIGWUser(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}

	plan.updateState(user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apigwUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	user := getAPIGWUser(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}

	err := apigw.NewUserOp(r.client).Delete(ctx, user.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete API Gateway User[%s]: %s", user.ID.Value.String(), err))
		return
	}
}

func getAPIGWUser(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.UserDetail {
	userOp := apigw.NewUserOp(client)
	user, err := userOp.Read(ctx, uuid.MustParse(id))
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read APIGW User[%s]: %s", id, err))
		return nil
	}

	return user
}

func updateUserExtra(ctx context.Context, client *v1.Client, userID uuid.UUID, plan, config *apigwUserResourceModel) error {
	ueOp := apigw.NewUserExtraOp(client, userID)

	if len(plan.Groups) > 0 {
		for _, g := range plan.Groups {
			idOrName := g.ID.ValueString()
			if idOrName == "" {
				idOrName = g.Name.ValueString()
			}
			err := ueOp.UpdateGroup(ctx, idOrName, true)
			if err != nil {
				return fmt.Errorf("failed to add user[%s] to group[%s]: %s", userID.String(), g.ID.ValueString(), err)
			}
		}
	}

	if plan.Authentication != nil {
		auth := v1.UserAuthentication{}
		if plan.Authentication.BasicAuth != nil {
			auth.BasicAuth = v1.NewOptBasicAuth(v1.BasicAuth{
				UserName: plan.Authentication.BasicAuth.Username.ValueString(),
				Password: config.Authentication.BasicAuth.PasswordWO.ValueString(),
			})
		}
		if plan.Authentication.JWT != nil {
			auth.Jwt = v1.NewOptJwt(v1.Jwt{
				Key:       plan.Authentication.JWT.Key.ValueString(),
				Secret:    config.Authentication.JWT.SecretWO.ValueString(),
				Algorithm: v1.JwtAlgorithm(plan.Authentication.JWT.Algorithm.ValueString()),
			})
		}
		if plan.Authentication.HMACAuth != nil {
			auth.HmacAuth = v1.NewOptHmacAuth(v1.HmacAuth{
				UserName: plan.Authentication.HMACAuth.Username.ValueString(),
				Secret:   config.Authentication.HMACAuth.SecretWO.ValueString(),
			})
		}
		err := ueOp.UpdateAuth(ctx, auth)
		if err != nil {
			return fmt.Errorf("failed to update authentication for user[%s]: %s", userID.String(), err)
		}
	}

	return nil
}

func removeGroups(ctx context.Context, client *v1.Client, userID uuid.UUID, cur []apigwGroupModel, pre []v1.Group) error {
	ueOp := apigw.NewUserExtraOp(client, userID)

	var groups []string
	if len(cur) == 0 {
		for _, g := range pre {
			groups = append(groups, g.ID.Value.String())
		}
	} else {
		for _, g := range pre {
			found := false
			for _, c := range cur {
				if g.ID.Value.String() == c.ID.ValueString() || string(g.Name.Value) == c.Name.ValueString() {
					found = true
					break
				}
			}
			if !found {
				groups = append(groups, g.ID.Value.String())
			}
		}
	}

	if len(groups) > 0 {
		for _, g := range groups {
			err := ueOp.UpdateGroup(ctx, g, false)
			if err != nil {
				return fmt.Errorf("failed to remove user[%s] from group[%s]: %s", userID.String(), g, err)
			}
		}
	}

	return nil
}

func expandAPIGWUserRequest(plan *apigwUserResourceModel) *v1.UserDetail {
	res := &v1.UserDetail{
		Name:     v1.Name(plan.Name.ValueString()),
		Tags:     common.TsetToStrings(plan.Tags),
		CustomID: v1.NewOptString(plan.CustomID.ValueString()),
	}
	if plan.IPRestriction != nil {
		res.IpRestrictionConfig = v1.NewOptIpRestrictionConfig(
			v1.IpRestrictionConfig{
				Protocols:    v1.IpRestrictionConfigProtocols(plan.IPRestriction.Protocols.ValueString()),
				RestrictedBy: v1.IpRestrictionConfigRestrictedBy(plan.IPRestriction.RestrictedBy.ValueString()),
				Ips:          common.TsetToStrings(plan.IPRestriction.Ips),
			},
		)
	}

	return res
}

func flattenAPIGWUserAuthenticationResource(model *apigwUserAuthenticationResourceModel, auth *v1.UserAuthentication) *apigwUserAuthenticationResourceModel {
	if auth == nil {
		return nil
	}

	authentication := &apigwUserAuthenticationResourceModel{}
	if auth.BasicAuth.IsSet() {
		basic := auth.BasicAuth.Value
		bm := &apigwUserAuthenticationBasicAuthResourceModel{
			Username: types.StringValue(basic.UserName),
		}
		// import時はmodelがnilの可能性があるためチェック
		if model != nil && model.BasicAuth != nil && utils.IsKnown(model.BasicAuth.PasswordWOVersion) {
			bm.PasswordWOVersion = types.Int32Value(model.BasicAuth.PasswordWOVersion.ValueInt32())
		}
		authentication.BasicAuth = bm
	}
	if auth.Jwt.IsSet() {
		jwt := auth.Jwt.Value
		jm := &apigwUserAuthenticationJWTResourceModel{
			Key:       types.StringValue(jwt.Key),
			Algorithm: types.StringValue(string(jwt.Algorithm)),
		}
		// ditto
		if model != nil && model.JWT != nil && utils.IsKnown(model.JWT.SecretWOVersion) {
			jm.SecretWOVersion = types.Int32Value(model.JWT.SecretWOVersion.ValueInt32())
		}
		authentication.JWT = jm
	}
	if auth.HmacAuth.IsSet() {
		hmac := auth.HmacAuth.Value
		hm := &apigwUserAuthenticationHMACAuthResourceModel{
			Username: types.StringValue(hmac.UserName),
		}
		// ditto
		if model != nil && model.HMACAuth != nil && utils.IsKnown(model.HMACAuth.SecretWOVersion) {
			hm.SecretWOVersion = types.Int32Value(model.HMACAuth.SecretWOVersion.ValueInt32())
		}
		authentication.HMACAuth = hm
	}

	return authentication
}
