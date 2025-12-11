// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package container_registry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	registryBuilder "github.com/sacloud/iaas-service-go/containerregistry/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type containerRegistryBaseModel struct {
	common.SakuraBaseModel
	AccessLevel    types.String                  `tfsdk:"access_level"`
	VirtualDomain  types.String                  `tfsdk:"virtual_domain"`
	SubDomainLabel types.String                  `tfsdk:"subdomain_label"`
	FQDN           types.String                  `tfsdk:"fqdn"`
	IconID         types.String                  `tfsdk:"icon_id"`
	User           []*containerRegistryUserModel `tfsdk:"user"`
}

type containerRegistryUserModel struct {
	Name       types.String `tfsdk:"name"`
	Password   types.String `tfsdk:"password"`
	Permission types.String `tfsdk:"permission"`
}

func (model *containerRegistryBaseModel) updateState(ctx context.Context, c *common.APIClient, reg *iaas.ContainerRegistry, includePassword bool, diags *diag.Diagnostics) {
	users := getContainerRegistryUsers(ctx, c, reg)
	if users == nil {
		diags.AddError("Get Users Error", "could not get users for SakuraCloud ContainerRegistry")
		return
	}

	model.UpdateBaseState(reg.ID.String(), reg.Name, reg.Description, reg.Tags)
	model.AccessLevel = types.StringValue(string(reg.AccessLevel))
	model.VirtualDomain = types.StringValue(reg.VirtualDomain)
	model.SubDomainLabel = types.StringValue(reg.SubDomainLabel)
	model.FQDN = types.StringValue(reg.FQDN)
	model.User = flattenContainerRegistryUsers(model.User, users, includePassword)
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

func expandContainerRegistryUsers(users []*containerRegistryUserModel) []*registryBuilder.User {
	if len(users) == 0 {
		return nil
	}

	var results []*registryBuilder.User
	for _, u := range users {
		results = append(results, &registryBuilder.User{
			UserName:   u.Name.ValueString(),
			Password:   u.Password.ValueString(),
			Permission: iaastypes.EContainerRegistryPermission(u.Permission.ValueString()),
		})
	}
	return results
}

func flattenContainerRegistryUsers(conf []*containerRegistryUserModel, users []*iaas.ContainerRegistryUser, includePassword bool) []*containerRegistryUserModel {
	inputs := expandContainerRegistryUsers(conf)

	var results []*containerRegistryUserModel
	for _, user := range users {
		v := &containerRegistryUserModel{
			Name:       types.StringValue(user.UserName),
			Permission: types.StringValue(string(user.Permission)),
		}
		if includePassword {
			password := ""
			for _, i := range inputs {
				if i.UserName == user.UserName {
					password = i.Password
					break
				}
			}
			v.Password = types.StringValue(password)
		}
		results = append(results, v)
	}
	return results
}
