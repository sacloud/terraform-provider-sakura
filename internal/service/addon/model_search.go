// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
)

type searchBaseModel struct {
	ID             types.String `tfsdk:"id"`
	Location       types.String `tfsdk:"location"`
	PartitionCount types.Int32  `tfsdk:"partition_count"`
	ReplicaCount   types.Int32  `tfsdk:"replica_count"`
	Sku            types.Int32  `tfsdk:"sku"`
	DeploymentName types.String `tfsdk:"deployment_name"`
	URL            types.String `tfsdk:"url"`
}

func (model *searchBaseModel) updateState(id, deploymentName, url string, body *v1.SearchPostRequestBody) {
	model.ID = types.StringValue(id)
	model.Location = types.StringValue(body.Location)
	model.PartitionCount = types.Int32Value(body.PartitionCount)
	model.ReplicaCount = types.Int32Value(body.ReplicaCount)
	model.Sku = types.Int32Value(int32(body.Sku))
	model.URL = types.StringValue(url)
	model.DeploymentName = deploymentNameValue(deploymentName)
}
