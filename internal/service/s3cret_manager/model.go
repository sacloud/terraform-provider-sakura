// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package secret_manager

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type secretManagerBaseModel struct {
	common.SakuraBaseModel
	KmsKeyID types.String `tfsdk:"kms_key_id"`
}

func (model *secretManagerBaseModel) updateState(vault *v1.Vault) {
	model.UpdateBaseState(vault.ID, vault.Name, vault.Description.Value, vault.Tags)
	model.KmsKeyID = types.StringValue(vault.KmsKeyID)
}

type secretManagerSecretBaseModel struct {
	Name    types.String `tfsdk:"name"`
	VaultID types.String `tfsdk:"vault_id"`
	Version types.Int64  `tfsdk:"version"`
	Value   types.String `tfsdk:"value"`
}
