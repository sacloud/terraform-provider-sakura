// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SakuraBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
}

func (model *SakuraBaseModel) UpdateBaseState(id string, name string, desc string, tags []string) {
	model.ID = types.StringValue(id)
	model.Name = types.StringValue(name)
	model.Description = types.StringValue(desc)
	model.Tags = FlattenTags(tags)
}

type SakuraMonitoringSuiteModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

func (m SakuraMonitoringSuiteModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
	}
}

type SakuraEncryptionDiskModel struct {
	EncryptionAlgorithm types.String `tfsdk:"encryption_algorithm"`
	KMSKeyID            types.String `tfsdk:"kms_key_id"`
}

func (m SakuraEncryptionDiskModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"encryption_algorithm": types.StringType,
		"kms_key_id":           types.StringType,
	}
}
