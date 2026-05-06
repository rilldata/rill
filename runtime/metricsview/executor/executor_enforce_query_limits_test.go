package executor

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

func TestEnforceQueryLimits(t *testing.T) {
	tr := func(days int) *metricsview.TimeRange {
		end := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		return &metricsview.TimeRange{
			Start: end.AddDate(0, 0, -days),
			End:   end,
		}
	}

	tests := []struct {
		name      string
		spec      string
		callerCap int64
		query     *metricsview.Query
		wantErr   string
	}{
		{
			name:  "no spec, no caller cap — passes",
			query: &metricsview.Query{TimeRange: tr(365)},
		},
		{
			name:  "spec cap with range under cap — passes",
			spec:  "P30D",
			query: &metricsview.Query{TimeRange: tr(7)},
		},
		{
			name:    "spec cap exceeded — fails",
			spec:    "P30D",
			query:   &metricsview.Query{TimeRange: tr(60)},
			wantErr: "max_query_time_range",
		},
		{
			name:      "caller cap tighter than spec — caller wins",
			spec:      "P90D",
			callerCap: 30,
			query:     &metricsview.Query{TimeRange: tr(60)},
			wantErr:   "rill.ai.max_time_range_days",
		},
		{
			name:      "caller cap with range under cap — passes",
			spec:      "",
			callerCap: 30,
			query:     &metricsview.Query{TimeRange: tr(7)},
		},
		{
			name:    "spec cap exceeded by comparison range — fails",
			spec:    "P30D",
			query:   &metricsview.Query{TimeRange: tr(7), ComparisonTimeRange: tr(60)},
			wantErr: "max_query_time_range",
		},
		{
			name:    "spec cap exceeded by primary, comparison fits — fails",
			spec:    "P30D",
			query:   &metricsview.Query{TimeRange: tr(60), ComparisonTimeRange: tr(7)},
			wantErr: "max_query_time_range",
		},
		{
			name:  "spec set but no time range on query — passes",
			spec:  "P30D",
			query: &metricsview.Query{},
		},
		{
			name: "require_time_range without time range — fails",
			query: &metricsview.Query{QueryLimits: &metricsview.QueryLimits{
				RequireTimeRange: true,
			}},
			wantErr: "valid time_range",
		},
		{
			name:    "spec error message names the property",
			spec:    "P30D",
			query:   &metricsview.Query{TimeRange: tr(60)},
			wantErr: "configured via the metrics view's max_query_time_range property",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executor{metricsView: &runtimev1.MetricsViewSpec{MaxQueryTimeRange: tt.spec}}
			q := tt.query
			if tt.callerCap > 0 {
				if q.QueryLimits == nil {
					q.QueryLimits = &metricsview.QueryLimits{}
				}
				q.QueryLimits.MaxTimeRangeDays = tt.callerCap
			}
			err := e.enforceQueryLimits(q)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
