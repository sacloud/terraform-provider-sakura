// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type processConfigurationResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                   = &processConfigurationResource{}
	_ resource.ResourceWithConfigure      = &processConfigurationResource{}
	_ resource.ResourceWithImportState    = &processConfigurationResource{}
	_ resource.ResourceWithValidateConfig = &processConfigurationResource{}
)

func NewEventBusProcessConfigurationResource() resource.Resource {
	return &processConfigurationResource{}
}

func (r *processConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eventbus_process_configuration"
}

func (r *processConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.EventBusClient
}

func requiredAttributeMissing(resp *resource.ValidateConfigResponse, rootAttributeName, destination string) {
	resp.Diagnostics.AddAttributeError(
		path.Root(rootAttributeName),
		"Missing attribute",
		fmt.Sprintf("Expected %q to be configured when destination is %q", rootAttributeName, destination),
	)
}

func (r *processConfigurationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config processConfigurationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch destination := config.Destination.ValueString(); destination {
	case destinationSimpleMQ:
		if config.SimpleMQAPIKey.ValueString() == "" {
			requiredAttributeMissing(resp, "simplemq_api_key_wo", destinationSimpleMQ)
		}
		version := config.CredentialsVersion
		if version.IsNull() || version.IsUnknown() {
			requiredAttributeMissing(resp, "credentials_wo_version", destinationSimpleMQ)
		}

	case destinationSimpleNotification, destinationAutoScale:
		if config.SimpleNotificationAccessToken.ValueString() == "" {
			requiredAttributeMissing(resp, "sakura_access_token_wo", destination)
		}
		if config.SimpleNotificationAccessTokenSecret.ValueString() == "" {
			requiredAttributeMissing(resp, "sakura_access_token_secret_wo", destination)
		}
		version := config.CredentialsVersion
		if version.IsNull() || version.IsUnknown() {
			requiredAttributeMissing(resp, "credentials_wo_version", destination)
		}
	default:
		resp.Diagnostics.AddAttributeError(
			path.Root("destination"),
			"Unknown destination",
			fmt.Sprintf("Destination should be either %q, %q or %q", destinationSimpleNotification, destinationSimpleMQ, destinationAutoScale),
		)
	}
}

type processConfigurationResourceModel struct {
	processConfigurationBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`

	SimpleNotificationAccessToken       types.String `tfsdk:"sakura_access_token_wo"`
	SimpleNotificationAccessTokenSecret types.String `tfsdk:"sakura_access_token_secret_wo"`
	SimpleMQAPIKey                      types.String `tfsdk:"simplemq_api_key_wo"`
	CredentialsVersion                  types.Int32  `tfsdk:"credentials_wo_version"`
}

func (r *processConfigurationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	const resourceName = "EventBus ProcessConfiguration"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			"tags":        common.SchemaResourceTags(resourceName),
			"icon_id":     common.SchemaResourceIconID(resourceName),

			"destination": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The destination of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						return v1.ProcessConfigurationSettingsDestination(v).Validate()
					}),
				},
			},
			"parameters": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The parameter of the %s.", resourceName),
			},

			"simplemq_api_key_wo": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: desc.Sprintf("The SimpleMQ API key for %s.", resourceName),
			},
			"sakura_access_token_wo": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: desc.Sprintf("The SimpleNotification/AutoScale access token for %s.", resourceName),
			},
			"sakura_access_token_secret_wo": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: desc.Sprintf("The SimpleNotification/AutoScale access token secret for %s.", resourceName),
			},
			"credentials_wo_version": schema.Int32Attribute{
				Optional:    true,
				Description: desc.Sprintf("Version number for credentials. Change this when changing credentials."),
			},

			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a EventBus Process Configuration.",
	}
}

func (r *processConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *processConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config processConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)
	pc, err := processConfigurationOp.Create(ctx, expandProcessConfigurationCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("create EventBus ProcessConfiguration failed: %s", err))
		return
	}
	pcID := pc.ID

	// SDK v2ではUpdateを呼び出して更新していたが、Frameworkではアクション間での状態の共有が難しいためメソッドに括り出して処理を共通化
	err = r.callProcessConfigurationUpdateSecretRequest(ctx, pcID, &config, pc)
	if err != nil {
		resp.Diagnostics.AddError("UpdateSecret Error", err.Error())
		return
	}

	gotPC := getProcessConfiguration(ctx, r.client, pcID, &resp.State, &resp.Diagnostics)
	if gotPC == nil {
		return
	}

	if err := plan.updateState(gotPC); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to update EventBus ProcessConfiguration[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *processConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state processConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pc := getProcessConfiguration(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if pc == nil {
		return
	}

	if err := state.updateState(pc); err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update EventBus ProcessConfiguration[%s] state: %s", state.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *processConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state processConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)

	if _, err := processConfigurationOp.Update(ctx, plan.ID.ValueString(), expandProcessConfigurationUpdateRequest(&plan)); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("update on EventBus ProcessConfiguration[%s] failed: %s", plan.ID.ValueString(), err))
		return
	}

	if !plan.CredentialsVersion.Equal(state.CredentialsVersion) {
		// credentialsの更新があるときだけ実行
		if err := r.callProcessConfigurationUpdateSecretRequest(ctx, plan.ID.ValueString(), &config, nil); err != nil {
			resp.Diagnostics.AddError("UpdateSecret Error", err.Error())
			return
		}
	}

	pc := getProcessConfiguration(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if pc == nil {
		return
	}

	if err := plan.updateState(pc); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update EventBus ProcessConfiguration[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *processConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state processConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)
	pc := getProcessConfiguration(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if pc == nil {
		return
	}

	if err := processConfigurationOp.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("delete EventBus ProcessConfiguration[%s] failed: %s", state.ID.ValueString(), err))
		return
	}
}

func (r *processConfigurationResource) callProcessConfigurationUpdateSecretRequest(ctx context.Context, id string, config *processConfigurationResourceModel, pc *v1.CommonServiceItem) error {
	var err error
	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)

	if pc == nil {
		_, err = processConfigurationOp.Read(ctx, id)
		if err != nil {
			return fmt.Errorf("could not read EventBus ProcessConfiguration[%s]: %w", id, err)
		}
	}

	err = processConfigurationOp.UpdateSecret(ctx, id, expandProcessConfigurationUpdateSecretRequest(config))
	if err != nil {
		return fmt.Errorf("update secret on EventBus ProcessConfiguration[%s] failed: %w", id, err)
	}

	return nil
}

func getProcessConfiguration(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.CommonServiceItem {
	processConfigurationOp := eventbus.NewProcessConfigurationOp(client)
	pc, err := processConfigurationOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get ProcessConfiguration Error", fmt.Sprintf("could not read EventBus ProcessConfiguration[%s]: %s", id, err))
		return nil
	}

	return pc
}

func expandProcessConfigurationCreateRequest(d *processConfigurationResourceModel) v1.CreateCommonServiceItemRequest {
	req := v1.CreateCommonServiceItemRequest{
		CommonServiceItem: v1.CreateCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: v1.NewOptNilString(d.Description.ValueString()),
			Settings: v1.NewProcessConfigurationSettingsSettings(v1.ProcessConfigurationSettings{
				Destination: v1.ProcessConfigurationSettingsDestination(d.Destination.ValueString()),
				Parameters:  d.Parameters.ValueString(),
			}),
			Provider: v1.Provider{
				Class: v1.ProviderClassEventbusprocessconfiguration,
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}

	if !d.IconID.IsNull() && !d.IconID.IsUnknown() {
		req.CommonServiceItem.Icon = v1.NewOptNilIcon(v1.Icon{
			ID: v1.NewOptString(d.IconID.ValueString()),
		})
	}

	return req
}

func expandProcessConfigurationUpdateRequest(d *processConfigurationResourceModel) v1.UpdateCommonServiceItemRequest {
	req := v1.UpdateCommonServiceItemRequest{
		CommonServiceItem: v1.UpdateCommonServiceItemRequestCommonServiceItem{
			Name:        v1.NewOptString(d.Name.ValueString()),
			Description: v1.NewOptNilString(d.Description.ValueString()),
			Settings: v1.NewOptSettings(
				v1.NewProcessConfigurationSettingsSettings(v1.ProcessConfigurationSettings{
					Destination: v1.ProcessConfigurationSettingsDestination(d.Destination.ValueString()),
					Parameters:  d.Parameters.ValueString(),
				}),
			),
			Provider: v1.NewOptProvider(
				v1.Provider{
					Class: v1.ProviderClassEventbusprocessconfiguration,
				},
			),
			Tags: common.TsetToStrings(d.Tags),
		},
	}

	if !d.IconID.IsNull() && !d.IconID.IsUnknown() {
		req.CommonServiceItem.Icon = v1.NewOptNilIcon(v1.Icon{
			ID: v1.NewOptString(d.IconID.ValueString()),
		})
	}

	return req
}

func expandProcessConfigurationUpdateSecretRequest(d *processConfigurationResourceModel) v1.SetSecretRequest {
	req := v1.SetSecretRequest{}

	switch destination := d.Destination.ValueString(); destination {
	case destinationSimpleMQ:
		req.Secret = v1.NewSimpleMQSecretSetSecretRequestSecret(v1.SimpleMQSecret{
			APIKey: d.SimpleMQAPIKey.ValueString(),
		})
	case destinationSimpleNotification, destinationAutoScale:
		req.Secret = v1.NewSacloudAPISecretSetSecretRequestSecret(v1.SacloudAPISecret{
			AccessToken:       d.SimpleNotificationAccessToken.ValueString(),
			AccessTokenSecret: d.SimpleNotificationAccessTokenSecret.ValueString(),
		})
	}

	return req
}
