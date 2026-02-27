// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
)

type etlBaseModel struct {
	ID             types.String `tfsdk:"id"`
	Location       types.String `tfsdk:"location"`
	DeploymentName types.String `tfsdk:"deployment_name"`
	URL            types.String `tfsdk:"url"`
}

func (model *etlBaseModel) updateState(id, deploymentName, url string, body *v1.EtlPostRequestBody) {
	model.ID = types.StringValue(id)
	model.Location = types.StringValue(body.Location)
	model.DeploymentName = deploymentNameValue(deploymentName)
	model.URL = types.StringValue(url)
}
