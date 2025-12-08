// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cloudhsm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type cloudHSMBaseModel struct {
	common.SakuraBaseModel
	Zone               types.String `tfsdk:"zone"`
	CreatedAt          types.String `tfsdk:"created_at"`
	ModifiedAt         types.String `tfsdk:"modified_at"`
	Availability       types.String `tfsdk:"availability"`
	IPv4NetworkAddress types.String `tfsdk:"ipv4_network_address"`
	IPv4Netmask        types.Int32  `tfsdk:"ipv4_netmask"`
	IPv4Address        types.String `tfsdk:"ipv4_address"`
	LocalRouter        types.Object `tfsdk:"local_router"`
}

type cloudHSMLocalRouterModel struct {
	ID        types.String `tfsdk:"id"`
	SecretKey types.String `tfsdk:"secret_key"`
}

func (m cloudHSMLocalRouterModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":         types.StringType,
		"secret_key": types.StringType,
	}
}

func (model *cloudHSMBaseModel) updateState(cloudhsm *v1.CloudHSM, zone string) {
	model.UpdateBaseState(cloudhsm.ID, cloudhsm.Name, cloudhsm.Description.Value, cloudhsm.Tags)
	model.Zone = types.StringValue(zone)
	model.CreatedAt = types.StringValue(string(cloudhsm.CreatedAt))
	model.ModifiedAt = types.StringValue(string(cloudhsm.ModifiedAt))
	model.Availability = types.StringValue(string(cloudhsm.Availability))
	model.IPv4NetworkAddress = types.StringValue(cloudhsm.Ipv4NetworkAddress)
	model.IPv4Netmask = types.Int32Value(int32(cloudhsm.Ipv4PrefixLength))
	model.IPv4Address = types.StringValue(cloudhsm.Ipv4Address)

	if router, ok := cloudhsm.LocalRouter.Get(); ok {
		m := cloudHSMLocalRouterModel{
			ID:        types.StringValue(router.ResourceID.Value),
			SecretKey: types.StringValue(router.SecretKey.Value),
		}
		value, _ := types.ObjectValueFrom(context.Background(), m.AttributeTypes(), m)
		model.LocalRouter = value
	} else {
		model.LocalRouter = types.ObjectNull(cloudHSMLocalRouterModel{}.AttributeTypes())
	}
}

type cloudHSMClientBaseModel struct {
	ID           types.String `tfsdk:"id"`
	Zone         types.String `tfsdk:"zone"`
	Name         types.String `tfsdk:"name"`
	CloudHSMID   types.String `tfsdk:"cloudhsm_id"`
	Certificate  types.String `tfsdk:"certificate"`
	CreatedAt    types.String `tfsdk:"created_at"`
	ModifiedAt   types.String `tfsdk:"modified_at"`
	Availability types.String `tfsdk:"availability"`
}

func (model *cloudHSMClientBaseModel) updateState(cloudhsmClient *v1.CloudHSMClient, zone, cloudhsmId string) {
	model.ID = types.StringValue(cloudhsmClient.ID)
	model.Zone = types.StringValue(zone)
	model.Name = types.StringValue(cloudhsmClient.Name)
	model.CloudHSMID = types.StringValue(cloudhsmId)
	model.Certificate = types.StringValue(cloudhsmClient.Certificate)
	model.CreatedAt = types.StringValue(string(cloudhsmClient.CreatedAt))
	model.ModifiedAt = types.StringValue(string(cloudhsmClient.ModifiedAt))
	model.Availability = types.StringValue(string(cloudhsmClient.Availability))
}

type cloudHSMPeerBaseModel struct {
	ID         types.String `tfsdk:"id"`
	CloudHSMID types.String `tfsdk:"cloudhsm_id"`
	Zone       types.String `tfsdk:"zone"`
	Index      types.Int64  `tfsdk:"index"`
	Status     types.String `tfsdk:"status"`
	Routes     types.List   `tfsdk:"routes"`
}

func (model *cloudHSMPeerBaseModel) updateState(peer *v1.CloudHSMPeer, zone, cloudhsmId string) {
	model.ID = types.StringValue(peer.ID)
	model.Zone = types.StringValue(zone)
	model.Index = types.Int64Value(int64(peer.Index.Value))
	model.Status = types.StringValue(string(peer.Status.Value))
	model.Routes = common.StringsToTlist(peer.Routes)
	model.CloudHSMID = types.StringValue(cloudhsmId)
}

type cloudHSMLicenseBaseModel struct {
	common.SakuraBaseModel
	Zone       types.String `tfsdk:"zone"`
	CreatedAt  types.String `tfsdk:"created_at"`
	ModifiedAt types.String `tfsdk:"modified_at"`
}

func (model *cloudHSMLicenseBaseModel) updateState(license *v1.CloudHSMSoftwareLicense, zone string) {
	model.UpdateBaseState(license.ID, license.Name, license.Description, license.Tags)
	model.Zone = types.StringValue(zone)
	model.CreatedAt = types.StringValue(string(license.CreatedAt))
	model.ModifiedAt = types.StringValue(string(license.ModifiedAt))
}
