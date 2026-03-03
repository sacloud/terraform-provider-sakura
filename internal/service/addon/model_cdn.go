// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type cdnBaseModel struct {
	ID             types.String          `tfsdk:"id"`
	Location       types.String          `tfsdk:"location"`
	PricingLevel   types.Int32           `tfsdk:"pricing_level"`
	Patterns       types.List            `tfsdk:"patterns"`
	Origin         *frontDoorOriginModel `tfsdk:"origin"`
	DeploymentName types.String          `tfsdk:"deployment_name"`
	URL            types.String          `tfsdk:"url"`
}

func (model *cdnBaseModel) updateState(id, deploymentName, url string, body *v1.NetworkRequestBody) {
	model.ID = types.StringValue(id)
	// Locationを指定しても、Frontdoor関連サービスのAPIレスポンスでは"Global"になるため、Locationは固定値で設定する
	// 現状はjapaneastのみなのでそれを設定。将来的にはリソースグループのLocationをAPIレスポンスから取得して設定するように変更する
	// ddos/wafも同様
	model.Location = types.StringValue("japaneast")
	model.PricingLevel = types.Int32Value(int32(body.Profile.Level))
	model.Patterns = common.StringsToTlist(body.Endpoint.Route.Patterns)
	model.Origin = &frontDoorOriginModel{
		Hostname:   types.StringValue(body.Endpoint.Route.OriginGroup.Origin.HostName),
		HostHeader: types.StringValue(body.Endpoint.Route.OriginGroup.Origin.HostHeader),
	}
	model.DeploymentName = deploymentNameValue(deploymentName)
	model.URL = types.StringValue(url)
}
