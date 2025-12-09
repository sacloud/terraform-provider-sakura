// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cloudhsm

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/cloudhsm-api-go"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type cloudHSMLicenseDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &cloudHSMLicenseDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudHSMLicenseDataSource{}
)

func NewCloudHSMLicenseDataSource() datasource.DataSource {
	return &cloudHSMLicenseDataSource{}
}

func (d *cloudHSMLicenseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudhsm_license"
}

func (d *cloudHSMLicenseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type cloudHSMLicenseDataSourceModel struct {
	cloudHSMLicenseBaseModel
}

func (d *cloudHSMLicenseDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("CloudHSM License"),
			"name":        common.SchemaDataSourceName("CloudHSM License"),
			"description": common.SchemaDataSourceDescription("CloudHSM License"),
			"tags":        common.SchemaDataSourceTags("CloudHSM License"),
			"zone":        schemaDataSourceZone("CloudHSM License"),
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation date of the CloudHSM License",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "The modification date of the CloudHSM License",
			},
		},
		MarkdownDescription: "Get information about an existing CloudHSM License.",
	}
}

func (d *cloudHSMLicenseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudHSMLicenseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()
	if len(name) == 0 && len(id) == 0 {
		resp.Diagnostics.AddError("Missing Attribute", "either 'id' or 'name' must be specified.")
		return
	}

	zone := getZone(data.Zone, d.client, &resp.Diagnostics)
	client := createClient(zone, d.client)
	licenseOp := cloudhsm.NewLicenseOp(client)
	var chsmLicense *v1.CloudHSMSoftwareLicense
	var err error
	if len(name) > 0 {
		chsmLicenses, err := licenseOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("List Error", fmt.Sprintf("failed to list CloudHSM License resource: %s", err))
			return
		}
		chsmLicense, err = FilterCloudHSMLicenseByName(chsmLicenses, name)
		if err != nil {
			resp.Diagnostics.AddError("Filter Error", err.Error())
			return
		}
	} else {
		chsmLicense, err = licenseOp.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to find CloudHSM License resource[%s]: %s", id, err.Error()))
			return
		}
	}

	data.updateState(chsmLicense, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FilterCloudHSMLicenseByName(cloudHSMs []v1.CloudHSMSoftwareLicense, name string) (*v1.CloudHSMSoftwareLicense, error) {
	match := slices.Collect(func(yield func(v1.CloudHSMSoftwareLicense) bool) {
		for _, v := range cloudHSMs {
			if name != v.Name {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple CloudHSM License resources found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
