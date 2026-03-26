// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type attrTypes = map[string]attr.Type

type dataSourceClient struct {
	client *v1.Client
	name   string
}

func dataSourceNamed(name string) dataSourceClient { return dataSourceClient{name: name} }

func (d *dataSourceClient) Configure(_ context.Context, req datasource.ConfigureRequest, res *datasource.ConfigureResponse) {
	client := common.GetApiClientFromProvider(req.ProviderData, &res.Diagnostics)

	if client == nil {
		return
	}

	d.client = client.AppRunDedicatedClient
}

func (d *dataSourceClient) Metadata(_ context.Context, req datasource.MetadataRequest, res *datasource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_apprun_dedicated_" + d.name
}

func (d *dataSourceClient) schemaID() (ret dschema.StringAttribute) {
	ret = common.SchemaDataSourceId(d.name).(dschema.StringAttribute)
	ret.Required = false
	ret.Optional = true
	ret.Computed = true
	ret.Validators = []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		sacloudvalidator.UUIDValidator,
	}

	return
}

func (d *dataSourceClient) schemaName() (ret dschema.StringAttribute) {
	ret = common.SchemaDataSourceName(d.name).(dschema.StringAttribute)
	ret.Required = false
	ret.Optional = true
	ret.Computed = true
	ret.Validators = []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}

	return
}

func (d *dataSourceClient) schemaClusterID() dschema.StringAttribute {
	return dschema.StringAttribute{
		Required:    true,
		Description: fmt.Sprintf("The cluster ID that the %s belongs to", d.name),
		Validators:  []validator.String{sacloudvalidator.UUIDValidator},
	}
}

func (d *dataSourceClient) schemaASGID() dschema.StringAttribute {
	return dschema.StringAttribute{
		Required:    true,
		Description: fmt.Sprintf("The auto scaling group ID that the %s belongs to", d.name),
		Validators:  []validator.String{sacloudvalidator.UUIDValidator},
	}
}

func (d *dataSourceClient) schemaCreatedAt() dschema.StringAttribute {
	return common.SchemaDataSourceCreatedAt(d.name).(dschema.StringAttribute)
}

////////////////////////////////////////////////////////////////

type resourceClient struct {
	client *v1.Client
	name   string
}

func resourceNamed(name string) resourceClient { return resourceClient{name: name} }

func (r *resourceClient) Configure(_ context.Context, req resource.ConfigureRequest, res *resource.ConfigureResponse) {
	client := common.GetApiClientFromProvider(req.ProviderData, &res.Diagnostics)

	if client == nil {
		return
	}

	r.client = client.AppRunDedicatedClient
}

func (r *resourceClient) Metadata(_ context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_apprun_dedicated_" + r.name
}

func (*resourceClient) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, res)
}

func (r *resourceClient) schemaID() rschema.StringAttribute {
	return common.SchemaResourceId(r.name).(rschema.StringAttribute)
}

func (r *resourceClient) schemaName(validators ...validator.String) (ret rschema.StringAttribute) {
	ret = common.SchemaResourceName(r.name).(rschema.StringAttribute)
	ret.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	ret.Validators = append([]validator.String{stringvalidator.LengthBetween(1, 20)}, validators...)

	return
}

func (r *resourceClient) schemaClusterID() (ret rschema.StringAttribute) {
	ret = common.SchemaResourceId("cluster").(rschema.StringAttribute)
	ret.Computed = false
	ret.Required = true
	ret.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	ret.Validators = []validator.String{sacloudvalidator.UUIDValidator}

	return
}

func (r *resourceClient) schemaASGID() (ret rschema.StringAttribute) {
	ret = common.SchemaResourceId("auto scaling group").(rschema.StringAttribute)
	ret.Computed = false
	ret.Required = true
	ret.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	ret.Validators = []validator.String{sacloudvalidator.UUIDValidator}

	return
}

func (r *resourceClient) schemaCreatedAt() rschema.StringAttribute {
	return common.SchemaResourceCreatedAt(r.name).(rschema.StringAttribute)
}
