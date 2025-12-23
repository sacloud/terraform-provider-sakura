// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package secret_manager

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type secretManagerSecretResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &secretManagerSecretResource{}
	_ resource.ResourceWithConfigure   = &secretManagerSecretResource{}
	_ resource.ResourceWithImportState = &secretManagerSecretResource{}
)

func NewSecretManagerSecretResource() resource.Resource {
	return &secretManagerSecretResource{}
}

func (r *secretManagerSecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_manager_secret"
}

func (r *secretManagerSecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SecretManagerClient
}

type secretManagerSecretResourceModel struct {
	secretManagerSecretBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *secretManagerSecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": common.SchemaResourceName("Secret Manager's secret"),
			"vault_id": schema.StringAttribute{
				Required:    true,
				Description: "The Secret Manager's vault id.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Version of secret value. This value is incremented by create/update.",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Secret value.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Secret Manager's secret.",
	}
}

func (r *secretManagerSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("name"), path.Root("name"), req, resp)
}

func (r *secretManagerSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretManagerSecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	secretOp := sm.NewSecretOp(r.client, plan.VaultID.ValueString())
	createdSec, err := secretOp.Create(ctx, v1.CreateSecret{
		Name:  plan.Name.ValueString(),
		Value: plan.Value.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Secret Manager's secret: %s", err))
		return
	}

	plan.Name = types.StringValue(createdSec.Name)
	plan.Version = types.Int64Value(int64(createdSec.LatestVersion))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretManagerSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretManagerSecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret := getSecretManagerSecret(ctx, r.client, &state, &resp.State, &resp.Diagnostics)
	if secret == nil {
		return
	}

	state.Name = types.StringValue(secret.Name)
	state.Version = types.Int64Value(int64(secret.LatestVersion))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *secretManagerSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan secretManagerSecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	secretOp := sm.NewSecretOp(r.client, plan.VaultID.ValueString())
	createdSec, err := secretOp.Create(ctx, v1.CreateSecret{
		Name:  plan.Name.ValueString(),
		Value: plan.Value.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Secret Manager's secret: %s", err))
		return
	}

	plan.Name = types.StringValue(createdSec.Name)
	plan.Version = types.Int64Value(int64(createdSec.LatestVersion))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretManagerSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretManagerSecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	sec := getSecretManagerSecret(ctx, r.client, &state, &resp.State, &resp.Diagnostics)
	if sec == nil {
		return
	}

	secretOp := sm.NewSecretOp(r.client, state.VaultID.ValueString())
	err := secretOp.Delete(ctx, v1.DeleteSecret{Name: state.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Secret Manager's secret: %s", err))
		return
	}
}

func getSecretManagerSecret(ctx context.Context, client *v1.Client, model *secretManagerSecretResourceModel, state *tfsdk.State, diags *diag.Diagnostics) *v1.Secret {
	secretOp := sm.NewSecretOp(client, model.VaultID.ValueString())
	secret, err := FilterSecretManagerSecretByName(ctx, secretOp, model.Name.ValueString())
	if err != nil {
		if err == common.ErrFilterNoResult {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Secret Manager's secret: %s", err))
		return nil
	}

	return secret
}

func FilterSecretManagerSecretByName(ctx context.Context, secretOp sm.SecretAPI, name string) (*v1.Secret, error) {
	secrets, err := secretOp.List(ctx)
	if err != nil {
		return nil, err
	}

	match := slices.Collect(func(yield func(v1.Secret) bool) {
		for _, v := range secrets {
			if name != v.Name {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})

	if len(match) == 0 {
		return nil, common.ErrFilterNoResult
	}

	return &match[0], nil
}
