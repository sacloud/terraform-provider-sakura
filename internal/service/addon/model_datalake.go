// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
)

type dataLakeBaseModel struct {
	ID             types.String `tfsdk:"id"`
	Location       types.String `tfsdk:"location"`
	Performance    types.Int32  `tfsdk:"performance"`
	Redundancy     types.Int32  `tfsdk:"redundancy"`
	DeploymentName types.String `tfsdk:"deployment_name"`
	URL            types.String `tfsdk:"url"`
}

func (model *dataLakeBaseModel) updateState(id, deploymentName, url string, body *v1.DatalakePostRequestBody) {
	model.ID = types.StringValue(id)
	model.Location = types.StringValue(body.Location)
	model.Performance = types.Int32Value(int32(body.Performance))
	model.Redundancy = types.Int32Value(int32(body.Redundancy))
	model.DeploymentName = deploymentNameValue(deploymentName)
	model.URL = types.StringValue(url)
}
