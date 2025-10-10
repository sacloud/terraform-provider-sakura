// Copyright 2016-2025 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	objectstorage "github.com/sacloud/object-storage-api-go"
	v1 "github.com/sacloud/object-storage-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageSiteDataSource struct {
	client *objectstorage.Client
}

var (
	_ datasource.DataSource              = &objectStorageSiteDataSource{}
	_ datasource.DataSourceWithConfigure = &objectStorageSiteDataSource{}
)

func NewObjectStorageSiteDataSource() datasource.DataSource {
	return &objectStorageSiteDataSource{}
}

func (d *objectStorageSiteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_site"
}

func (d *objectStorageSiteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.ObjectStorageClient
}

type objectStorageSiteDataSourceModel struct {
	ID           types.String                  `tfsdk:"id"`
	DisplayName  types.String                  `tfsdk:"display_name"`
	DisplayOrder types.Int32                   `tfsdk:"display_order"`
	Endpoint     types.String                  `tfsdk:"endpoint"`
	IamEndpoint  types.String                  `tfsdk:"iam_endpoint"`
	S3Endpoint   types.String                  `tfsdk:"s3_endpoint"`
	ApiZone      types.Set                     `tfsdk:"api_zone"`
	StorageZone  types.Set                     `tfsdk:"storage_zone"`
	Status       *objectStorageSiteStatusModel `tfsdk:"status"`
}

type objectStorageSiteStatusModel struct {
	AcceptNew types.Bool   `tfsdk:"accept_new"`
	Message   types.String `tfsdk:"message"`
	StartedAt types.String `tfsdk:"started_at"`
	Code      types.Int64  `tfsdk:"code"`
	Status    types.String `tfsdk:"status"`
}

func (d *objectStorageSiteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaDataSourceId("Object Storage Site"),
			"display_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The display name of the Object Storage Site.",
			},
			"display_order": schema.Int32Attribute{
				Computed:    true,
				Description: "The display order of the Object Storage Site.",
			},
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "The API endpoint URL for the Object Storage Site.",
			},
			"iam_endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "The IAM endpoint URL for the Object Storage Site.",
			},
			"s3_endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "The S3 endpoint URL for the Object Storage Site.",
			},
			"api_zone": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A set of API zone",
			},
			"storage_zone": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A set of storage zone",
			},
			"status": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"accept_new": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the site is accepting new request.",
					},
					"message": schema.StringAttribute{
						Computed:    true,
						Description: "Extra message for the site.",
					},
					"started_at": schema.StringAttribute{
						Computed:    true,
						Description: "The time when the site started.",
					},
					"code": schema.Int64Attribute{
						Computed:    true,
						Description: "The status code.",
					},
					"status": schema.StringAttribute{
						Computed:    true,
						Description: "The current status.",
					},
				},
			},
		},
	}
}

func (d *objectStorageSiteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data objectStorageSiteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var site *v1.Cluster
	var err error
	if data.ID.ValueString() == "" && data.DisplayName.ValueString() == "" {
		site, err = getSite(d.client, ctx, "isk01", "")
	} else {
		site, err = getSite(d.client, ctx, data.ID.ValueString(), data.DisplayName.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to get site: %s", err.Error()))
		return
	}
	statusOp := objectstorage.NewSiteStatusOp(d.client)
	status, err := statusOp.Read(ctx, site.Id)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to get site status: %s", err.Error()))
		return
	}

	data.ID = types.StringValue(site.Id)
	data.DisplayName = types.StringValue(site.DisplayName)
	data.DisplayOrder = types.Int32Value(int32(site.DisplayOrder))
	data.Endpoint = types.StringValue(site.EndpointBase)
	data.IamEndpoint = types.StringValue(site.IamEndpoint)
	data.S3Endpoint = types.StringValue(site.S3Endpoint)
	data.ApiZone = common.StringsToTset(site.ApiZone)
	data.StorageZone = common.StringsToTset(site.StorageZone)
	data.Status = &objectStorageSiteStatusModel{
		AcceptNew: types.BoolValue(status.AcceptNew),
		Message:   types.StringValue(status.Message),
		StartedAt: types.StringValue(status.StartedAt.String()),
		Code:      types.Int64Value(int64(status.StatusCode.Id)),
		Status:    types.StringValue(status.StatusCode.Status),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getSite(client *objectstorage.Client, ctx context.Context, id string, displayName string) (*v1.Cluster, error) {
	siteOp := objectstorage.NewSiteOp(client)

	if id != "" {
		site, err := siteOp.Read(ctx, id)
		if err != nil {
			return nil, err
		}
		return site, nil
	} else {
		sites, err := siteOp.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, s := range sites {
			if strings.HasPrefix(s.DisplayName, displayName) || strings.HasPrefix(s.DisplayNameJa, displayName) || strings.HasPrefix(s.DisplayNameEnUs, displayName) {
				return s, nil
			}
		}
	}

	return nil, fmt.Errorf("object storage site not found")
}
