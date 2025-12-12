package monitoring_suite

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type logStorageBaseModel struct {
	common.SakuraBaseModel

	Classification types.String `tfsdk:"classification"`
	ExpireDay      types.Int64  `tfsdk:"expire_day"`

	AccountID types.String `tfsdk:"account_id"`
	IsSystem  types.Bool   `tfsdk:"is_system"`

	Endpoints *logStorageEndpointsModel `tfsdk:"endpoints"`
	Usage     *logStorageUsageModel     `tfsdk:"usage"`
}

type logStorageEndpointsModel struct {
	Ingester *logStorageIngesterEndpointModel `tfsdk:"ingester"`
}

type logStorageIngesterEndpointModel struct {
	Address  types.String `tfsdk:"address"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

type logStorageUsageModel struct {
	LogRoutings     types.Int64 `tfsdk:"log_routings"`
	LogMeasureRules types.Int64 `tfsdk:"log_measure_rules"`
}

type logStorageResourceModel struct {
	logStorageBaseModel
	Timeouts common.TimeoutsValue `tfsdk:"timeouts"`
}

type logStorageDataSourceModel struct {
	logStorageBaseModel
}

func (m *logStorageBaseModel) updateState(_ context.Context, ls *v1.LogStorage) error {
	if ls == nil {
		return nil
	}

    // TODO

	return nil
}
