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
			name:  "no cap",
			query: &metricsview.Query{TimeRange: tr(365)},
		},
		{
			name:  "spec cap, range under",
			spec:  "P30D",
			query: &metricsview.Query{TimeRange: tr(7)},
		},
		{
			name:    "spec cap, range over",
			spec:    "P30D",
			query:   &metricsview.Query{TimeRange: tr(60)},
			wantErr: "max_query_time_range",
		},
		{
			name:      "caller cap wins when tighter than spec",
			spec:      "P90D",
			callerCap: 30,
			query:     &metricsview.Query{TimeRange: tr(60)},
			wantErr:   "rill.ai.max_time_range_days",
		},
		{
			name:      "caller cap, range under",
			callerCap: 30,
			query:     &metricsview.Query{TimeRange: tr(7)},
		},
		{
			name:    "comparison range over cap",
			spec:    "P30D",
			query:   &metricsview.Query{TimeRange: tr(7), ComparisonTimeRange: tr(60)},
			wantErr: "max_query_time_range",
		},
		{
			name:    "primary range over cap, comparison fits",
			spec:    "P30D",
			query:   &metricsview.Query{TimeRange: tr(60), ComparisonTimeRange: tr(7)},
			wantErr: "max_query_time_range",
		},
		{
			name:  "spec set, no time range on query",
			spec:  "P30D",
			query: &metricsview.Query{},
		},
		{
			name: "require_time_range without time range",
			query: &metricsview.Query{QueryLimits: &metricsview.QueryLimits{
				RequireTimeRange: true,
			}},
			wantErr: "valid time_range",
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
