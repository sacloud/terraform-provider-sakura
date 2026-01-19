// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwServiceBaseModel struct {
	ID             types.String           `tfsdk:"id"`
	Name           types.String           `tfsdk:"name"`
	Tags           types.Set              `tfsdk:"tags"`
	CreatedAt      types.String           `tfsdk:"created_at"`
	UpdatedAt      types.String           `tfsdk:"updated_at"`
	SubscriptionID types.String           `tfsdk:"subscription_id"`
	Protocol       types.String           `tfsdk:"protocol"`
	Host           types.String           `tfsdk:"host"`
	Path           types.String           `tfsdk:"path"`
	Port           types.Int32            `tfsdk:"port"`
	Retries        types.Int32            `tfsdk:"retries"`
	ConnectTimeout types.Int32            `tfsdk:"connect_timeout"`
	WriteTimeout   types.Int32            `tfsdk:"write_timeout"`
	ReadTimeout    types.Int32            `tfsdk:"read_timeout"`
	Authentication types.String           `tfsdk:"authentication"`
	RouteHost      types.String           `tfsdk:"route_host"`
	OIDC           *apigwServiceOIDCModel `tfsdk:"oidc"`
	CORSConfig     *apigwServiceCORSModel `tfsdk:"cors_config"`
}

type apigwServiceOIDCModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type apigwServiceCORSModel struct {
	Credentials                 types.Bool   `tfsdk:"credentials"`
	AccessControlExposedHeaders types.String `tfsdk:"access_control_exposed_headers"`
	AccessControlAllowHeaders   types.String `tfsdk:"access_control_allow_headers"`
	MaxAge                      types.Int32  `tfsdk:"max_age"`
	AccessControlAllowMethods   types.Set    `tfsdk:"access_control_allow_methods"`
	AccessControlAllowOrigins   types.String `tfsdk:"access_control_allow_origins"`
	PreflightContinue           types.Bool   `tfsdk:"preflight_continue"`
	PrivateNetwork              types.Bool   `tfsdk:"private_network"`
}

type apigwServiceObjectStorageModel struct {
	Bucket           types.String `tfsdk:"bucket"`
	Folder           types.String `tfsdk:"folder"`
	Endpoint         types.String `tfsdk:"endpoint"`
	Region           types.String `tfsdk:"region"`
	UseDocumentIndex types.Bool   `tfsdk:"use_document_index"`
}

func (m *apigwServiceBaseModel) updateState(service *v1.ServiceDetailResponse) {
	m.ID = types.StringValue(service.ID.Value.String())
	m.Name = types.StringValue(string(service.Name))
	m.Tags = common.StringsToTset(service.Tags)
	m.CreatedAt = types.StringValue(service.CreatedAt.Value.String())
	m.UpdatedAt = types.StringValue(service.UpdatedAt.Value.String())
	m.SubscriptionID = types.StringValue(service.Subscription.ID.String())
	m.Protocol = types.StringValue(string(service.Protocol))
	m.Host = types.StringValue(service.Host)
	m.Path = types.StringValue(service.Path.Value)
	m.Port = types.Int32Value(int32(service.Port.Value))
	m.Retries = types.Int32Value(int32(service.Retries.Value))
	m.ConnectTimeout = types.Int32Value(int32(service.ConnectTimeout.Value))
	m.WriteTimeout = types.Int32Value(int32(service.WriteTimeout.Value))
	m.ReadTimeout = types.Int32Value(int32(service.ReadTimeout.Value))
	m.Authentication = types.StringValue(string(service.Authentication.Value))
	m.RouteHost = types.StringValue(service.RouteHost.Value)

	if service.Oidc.IsSet() {
		m.OIDC = &apigwServiceOIDCModel{
			ID:   types.StringValue(service.Oidc.Value.ID.Value.String()),
			Name: types.StringValue(service.Oidc.Value.Name.Value),
		}
	} else {
		m.OIDC = nil
	}

	if service.CorsConfig.IsSet() {
		cors := service.CorsConfig.Value
		m.CORSConfig = &apigwServiceCORSModel{
			Credentials:                 types.BoolValue(cors.Credentials.Value),
			AccessControlExposedHeaders: types.StringValue(cors.AccessControlExposedHeaders.Value),
			AccessControlAllowHeaders:   types.StringValue(cors.AccessControlAllowHeaders.Value),
			MaxAge:                      types.Int32Value(cors.MaxAge.Value),
			AccessControlAllowMethods:   common.StringsToTset(common.MapTo(cors.AccessControlAllowMethods, common.ToString)),
			AccessControlAllowOrigins:   types.StringValue(cors.AccessControlAllowOrigins.Value),
			PreflightContinue:           types.BoolValue(cors.PreflightContinue.Value),
			PrivateNetwork:              types.BoolValue(cors.PrivateNetwork.Value),
		}
	} else {
		m.CORSConfig = nil
	}
}

func (m *apigwServiceObjectStorageModel) updateState(obst *v1.ObjectStorageConfig) {
	m.Bucket = types.StringValue(obst.BucketName)
	m.Endpoint = types.StringValue(obst.Endpoint)
	m.Region = types.StringValue(obst.Region)
	m.UseDocumentIndex = types.BoolValue(obst.UseDocumentIndex)

	if obst.FolderName.IsSet() {
		m.Folder = types.StringValue(obst.FolderName.Value)
	}
}

func flattenAPIGWServiceObjectStorageConfig(conf v1.OptObjectStorageConfig) *apigwServiceObjectStorageModel {
	if conf.IsSet() {
		obst := &apigwServiceObjectStorageModel{}
		obst.updateState(&conf.Value)
		return obst
	} else {
		return nil
	}
}
