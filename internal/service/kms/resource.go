// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package kms

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/kms-api-go"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type kmsResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &kmsResource{}
	_ resource.ResourceWithConfigure   = &kmsResource{}
	_ resource.ResourceWithImportState = &kmsResource{}
)

func NewKMSResource() resource.Resource {
	return &kmsResource{}
}

func (r *kmsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms"
}

func (r *kmsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.KmsClient
}

type kmsResourceModel struct {
	kmsBaseModel
	PlainKey                types.String   `tfsdk:"plain_key"`
	ScheduleDestructionDays types.Int32    `tfsdk:"schedule_destruction_days"`
	RotateVersion           types.Int64    `tfsdk:"rotate_version"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

func (r *kmsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("KMS key"),
			"name":        common.SchemaResourceName("KMS key"),
			"description": common.SchemaResourceDescription("KMS key"),
			"tags":        common.SchemaResourceTags("KMS key"),
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The status of the KMS key.",
				Default:     stringdefault.StaticString(string(v1.ChangeKeyStatusStatusActive)),
				Validators: []validator.String{
					stringvalidator.OneOf(common.MapTo(v1.ChangeKeyStatusStatusActive.AllValues(), common.ToString)...),
				},
			},
			"key_origin": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("generated"),
				Description: "Key origin of the KMS key. 'generated' or 'imported'. Default is 'generated'.",
				Validators: []validator.String{
					stringvalidator.OneOf("generated", "imported"),
				},
			},
			"plain_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Plain key for imported KMS key. Required when `key_origin` is 'imported'.",
			},
			"schedule_destruction_days": schema.Int32Attribute{
				Optional:    true,
				Description: "The number of days to schedule the destruction of the KMS key. If set, the KMS key will be scheduled for destruction after the specified number of days instead of immediate destruction in 'terraform destroy'.",
				Validators: []validator.Int32{
					int32validator.Between(7, 90),
				},
			},
			"rotate_version": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "The rotateion version. This number is incremented when you want rotate KMS key.",
			},
			"latest_version": schema.Int64Attribute{
				Computed:    true,
				Description: "The latest material version of the KMS key.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation time of the KMS key.",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "The last modification time of the KMS key.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a KMS.",
	}
}

func (r *kmsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *kmsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan kmsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	keyReq, err := expandKMSCreateKey(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Create: Expand Error", fmt.Sprintf("failed to expand create key request: %s", err))
		return
	}

	keyOp := kms.NewKeyOp(r.client)
	createdKey, err := keyOp.Create(ctx, keyReq)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create KMS key: %s", err))
		return
	}

	err = updateKMS(ctx, r.client, &plan, nil, createdKey.ID, "active")
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update KMS key status: %s", err))
		return
	}

	key := getKMS(ctx, r.client, createdKey.ID, &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	plan.updateState(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kmsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data kmsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := getKMS(ctx, r.client, data.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	data.updateState(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kmsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state kmsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	keyOp := kms.NewKeyOp(r.client)
	key := getKMS(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	_, err := keyOp.Update(ctx, key.ID, expandKMSUpdateKey(&plan, key))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update KMS key: %s", err))
		return
	}

	err = updateKMS(ctx, r.client, &plan, &state, key.ID, string(key.Status))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update KMS key status: %s", err))
		return
	}

	key = getKMS(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	plan.updateState(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kmsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state kmsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	keyOp := kms.NewKeyOp(r.client)
	key := getKMS(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	if !state.ScheduleDestructionDays.IsNull() {
		err := keyOp.ScheduleDestruction(ctx, state.ID.ValueString(), int(state.ScheduleDestructionDays.ValueInt32()))
		if err != nil {
			resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to set schedule destruction of KMS key[%s]: %s", state.ID.ValueString(), err))
			return
		}
	} else {
		if err := keyOp.Delete(ctx, key.ID); err != nil {
			resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete KMS key: %s", err))
			return
		}
	}
}

func getKMS(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Key {
	keyOp := kms.NewKeyOp(client)
	key, err := keyOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read SakuraCloud KMS key[%s]: %s", id, err))
		return nil
	}

	return key
}

func expandKMSCreateKey(model *kmsResourceModel) (v1.CreateKey, error) {
	keyOrig := model.KeyOrigin.ValueString()
	var req v1.CreateKey
	if keyOrig == "generated" {
		req = v1.CreateKey{
			Name:      model.Name.ValueString(),
			KeyOrigin: v1.KeyOriginEnumGenerated,
		}
	} else {
		plainKey := model.PlainKey.ValueString()
		if plainKey == "" {
			return v1.CreateKey{}, errors.New("plain_key is required when key_origin is 'imported'")
		}
		req = v1.CreateKey{
			Name:      model.Name.ValueString(),
			KeyOrigin: v1.KeyOriginEnumImported,
			PlainKey:  v1.NewOptString(plainKey),
		}
	}

	if !model.Tags.IsNull() {
		req.Tags = common.TsetToStrings(model.Tags)
	}
	if !model.Description.IsNull() {
		req.Description = v1.NewOptString(model.Description.ValueString())
	}

	return req, nil
}

func expandKMSUpdateKey(model *kmsResourceModel, before *v1.Key) v1.Key {
	req := v1.Key{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		KeyOrigin:   before.KeyOrigin,
	}

	if !model.Tags.IsNull() {
		req.Tags = common.TsetToStrings(model.Tags)
	}

	return req
}

func updateKMS(ctx context.Context, client *v1.Client, model *kmsResourceModel, before *kmsResourceModel, id string, beforeStatus string) error {
	keyOp := kms.NewKeyOp(client)

	status := model.Status.ValueString()
	if before != nil && status != beforeStatus {
		err := keyOp.ChangeStatus(ctx, id, v1.ChangeKeyStatusStatus(status))
		if err != nil {
			return fmt.Errorf("failed to change status of KMS key[%s]: %s", id, err)
		}
	}

	if before != nil && !model.RotateVersion.Equal(before.RotateVersion) {
		if beforeStatus == string(v1.ChangeKeyStatusStatusActive) {
			tflog.Info(ctx, fmt.Sprintf("Rotating KMS key[%s]", id))
			_, err := keyOp.Rotate(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to rotate KMS key[%s]: %s", id, err)
			}
		} else {
			tflog.Warn(ctx, fmt.Sprintf("Can't rotate KMS key[%s] when status is not 'active'", id))
		}
	}

	return nil
}
