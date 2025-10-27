// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ssh_key

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
)

type sshKeyBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	PublicKey   types.String `tfsdk:"public_key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

func (model *sshKeyBaseModel) updateState(key *iaas.SSHKey) {
	model.ID = types.StringValue(key.ID.String())
	model.Name = types.StringValue(key.Name)
	model.Description = types.StringValue(key.Description)
	model.PublicKey = types.StringValue(key.PublicKey)
	model.Fingerprint = types.StringValue(key.Fingerprint)
}
