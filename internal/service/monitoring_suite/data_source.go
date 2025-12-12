package monitoring_suite

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    monitoringsuitev1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
    "github.com/sacloud/terraform-provider-sakura/internal/common"
)

type alertProjectDataSource struct {
    client *monitoringsuitev1.Client
}

var (
    _ datasource.DataSource              = &alertProjectDataSource{}
    _ datasource.DataSourceWithConfigure = &alertProjectDataSource{}
)

func NewAlertProjectDataSource() datasource.DataSource {
    return &alertProjectDataSource{}
}

func (d *alertProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_project"
}

func (d *alertProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
    if apiClient == nil {
        return
    }

    d.client = apiClient.MonitoringSuiteClient
}
