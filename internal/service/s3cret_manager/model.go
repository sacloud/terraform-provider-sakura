// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secret_manager

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
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
