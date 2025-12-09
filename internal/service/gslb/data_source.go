package gslb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"

	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

var (
	_ datasource.DataSource              = &gslbDataSource{}
	_ datasource.DataSourceWithConfigure = &gslbDataSource{}
)

type gslbDataSource struct {
	client *common.APIClient
}

func NewGSLBDataSource() datasource.DataSource {
	return &gslbDataSource{}
}

type gslbDataSourceModel struct {
	gslbBaseModel
}

func (d *gslbDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gslb"
}

func (d *gslbDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

func (d *gslbDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("GSLB"),
			"name":        common.SchemaDataSourceName("GSLB"),
			"description": common.SchemaDataSourceDescription("GSLB"),
			"tags":        common.SchemaDataSourceTags("GSLB"),
			"icon_id":     common.SchemaDataSourceIconID("GSLB"),
			"fqdn": schema.StringAttribute{
				Computed:    true,
				Description: "The FQDN for accessing to the GSLB. This is typically used as value of CNAME record",
			},
			"health_check": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Health check configuration",
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Computed:    true,
						Description: desc.Sprintf("The protocol used for health checks. This will be one of [%s]", iaastypes.GSLBHealthCheckProtocolStrings),
					},
					"delay_loop": schema.Int64Attribute{
						Computed:    true,
						Description: "The interval in seconds between checks",
					},
					"host_header": schema.StringAttribute{
						Computed:    true,
						Description: "The value of host header send when checking by HTTP/HTTPS",
					},
					"path": schema.StringAttribute{
						Computed:    true,
						Description: "The path used when checking by HTTP/HTTPS",
					},
					"status": schema.StringAttribute{
						Computed:    true,
						Description: "The response-code to expect when checking by HTTP/HTTPS",
					},
					"port": schema.Int64Attribute{
						Computed:    true,
						Description: "The port number used when checking by TCP/HTTP/HTTPS",
					},
				},
			},
			"weighted": schema.BoolAttribute{
				Computed:    true,
				Description: "The flag to enable weighted load-balancing",
			},
			"sorry_server": schema.StringAttribute{
				Computed:    true,
				Description: "The IP address of the SorryServer. This will be used when all servers are down",
			},
			"server": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Computed:    true,
							Description: "The IP address of the server",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "The flag to enable as destination of load balancing",
						},
						"weight": schema.Int64Attribute{
							Computed:    true,
							Description: "The weight used when weighted load balancing is enabled",
						},
					},
				},
			},
			"monitoring_suite": common.SchemaDataSourceMonitoringSuite("GSLB"),
		},
		MarkdownDescription: "Get information about an existing GSLB.",
	}
}

func (d *gslbDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data gslbDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewGSLBOp(d.client)
	res, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Search Error", "failed to find SakuraCloud GSLB resource: "+err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.GSLBs) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.GSLBs[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
