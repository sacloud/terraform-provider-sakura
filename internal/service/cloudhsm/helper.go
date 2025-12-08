// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cloudhsm

import (
	"fmt"

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/cloudhsm-api-go"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

var (
	defaultZone    = "is1b"
	supportedZones = []string{"is1b", "tk1a"}
)

func getZone(zone types.String, client *common.APIClient, diags *diag.Diagnostics) string {
	z := common.GetZone(zone, client, diags)
	if err := common.StringInSlice(supportedZones, "zone", z, false); err != nil {
		diags.AddWarning("Zone Validation Warning", fmt.Sprintf("Default zone is not valid with CloudHSM. Use is1b instead: err = %s", err))
		z = defaultZone
	}
	return z
}

func createClient(zone string, apiClient *common.APIClient) *v1.Client {
	cloudhsmClient, err := cloudhsm.NewClient(client.WithOptions(apiClient.CallerOptions), cloudhsm.WithZone(zone))
	if err != nil {
		return nil
	}
	return cloudhsmClient
}

func schemaDataSourceZone(name string) dschema.Attribute {
	return dschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc.Sprintf("The zone of the %s. This must be one of [%s]", name, supportedZones),
		Validators: []validator.String{
			stringvalidator.OneOf(supportedZones...),
		},
	}
}

func schemaResourceZone(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Default:     stringdefault.StaticString(defaultZone),
		Description: desc.Sprintf("The zone of the %s. This must be one of [%s]", name, supportedZones),
		Validators: []validator.String{
			stringvalidator.OneOf(supportedZones...),
		},
	}
}
