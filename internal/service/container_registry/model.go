// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package container_registry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	registryBuilder "github.com/sacloud/iaas-service-go/containerregistry/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type containerRegistryBaseModel struct {
	common.SakuraBaseModel
	AccessLevel    types.String `tfsdk:"access_level"`
	VirtualDomain  types.String `tfsdk:"virtual_domain"`
	SubDomainLabel types.String `tfsdk:"subdomain_label"`
	FQDN           types.String `tfsdk:"fqdn"`
	IconID         types.String `tfsdk:"icon_id"`
}

func (model *containerRegistryBaseModel) updateState(reg *iaas.ContainerRegistry) {
	model.UpdateBaseState(reg.ID.String(), reg.Name, reg.Description, reg.Tags)
	model.AccessLevel = types.StringValue(string(reg.AccessLevel))
	model.VirtualDomain = types.StringValue(reg.VirtualDomain)
	model.SubDomainLabel = types.StringValue(reg.SubDomainLabel)
	model.FQDN = types.StringValue(reg.FQDN)
	if reg.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(reg.IconID.String())
	}
}

func getContainerRegistryUsers(ctx context.Context, client *common.APIClient, user *iaas.ContainerRegistry) []*iaas.ContainerRegistryUser {
	regOp := iaas.NewContainerRegistryOp(client)
	users, err := regOp.ListUsers(ctx, user.ID)
	if err != nil {
		return nil
	}
	return users.Users
}

func expandContainerRegistryUsers(users, configUsers []*containerRegistryUserModel) []*registryBuilder.User {
	if len(users) == 0 {
		return nil
	}

	var results []*registryBuilder.User
	for i, u := range users {
		password := u.Password.ValueString()
		if password == "" {
			password = configUsers[i].PasswordWO.ValueString()
		}
		results = append(results, &registryBuilder.User{
			UserName:   u.Name.ValueString(),
			Password:   password,
			Permission: iaastypes.EContainerRegistryPermission(u.Permission.ValueString()),
		})
	}
	return results
}
