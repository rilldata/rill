package clickhouse

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestDateTruncExpr(t *testing.T) {
	dateCol := &runtimev1.MetricsViewSpec_Dimension{Name: "d", Column: "d", DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}}
	dateExpr := &runtimev1.MetricsViewSpec_Dimension{Name: "d", Expression: "toDate(ts)", DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}}
	timestampCol := &runtimev1.MetricsViewSpec_Dimension{Name: "ts", Column: "ts", DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}}
	untypedCol := &runtimev1.MetricsViewSpec_Dimension{Name: "ts", Column: "ts"}

	tests := []struct {
		name           string
		dim            *runtimev1.MetricsViewSpec_Dimension
		grain          runtimev1.TimeGrain
		tz             string
		firstDayOfWeek int
		want           string
	}{
		// ClickHouse rejects date_trunc at sub-day grains on Date values, so those are cast to DateTime64 first.
		{name: "date column at hour grain", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_HOUR, want: `date_trunc('hour', "d"::DateTime64)::DateTime64`},
		{name: "date column at minute grain", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_MINUTE, want: `date_trunc('minute', "d"::DateTime64)::DateTime64`},
		{name: "date column at second grain", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_SECOND, want: `date_trunc('second', "d"::DateTime64)::DateTime64`},
		{name: "date expression at hour grain", dim: dateExpr, grain: runtimev1.TimeGrain_TIME_GRAIN_HOUR, want: `date_trunc('hour', (toDate(ts))::DateTime64)::DateTime64`},
		// Day and coarser grains accept Date values directly.
		{name: "date column at day grain", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_DAY, want: `date_trunc('day', "d")::DateTime64`},
		{name: "date column at month grain", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_MONTH, want: `date_trunc('month', "d")::DateTime64`},
		{name: "date column at week grain with a shifted first day", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_WEEK, firstDayOfWeek: 7, want: `date_trunc('week', "d" + INTERVAL 1 DAY)::DateTime64 - INTERVAL 1 DAY`},
		// The time zone path casts every input already.
		{name: "date column at hour grain in a time zone", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_HOUR, tz: "America/New_York", want: `date_trunc('hour', "d"::DateTime64(6, 'America/New_York'))::DateTime64(6, 'America/New_York')`},
		// Columns of unknown type may be Date values, so they are cast too.
		{name: "untyped column at hour grain", dim: untypedCol, grain: runtimev1.TimeGrain_TIME_GRAIN_HOUR, want: `date_trunc('hour', "ts"::DateTime64)::DateTime64`},
		{name: "untyped column at day grain", dim: untypedCol, grain: runtimev1.TimeGrain_TIME_GRAIN_DAY, want: `date_trunc('day', "ts")::DateTime64`},
		// Known DateTime columns are not wrapped in a cast.
		{name: "timestamp column at hour grain", dim: timestampCol, grain: runtimev1.TimeGrain_TIME_GRAIN_HOUR, want: `date_trunc('hour', "ts")::DateTime64`},
		// Multiples bucket with toStartOfInterval, which also rejects Date values at sub-day intervals.
		{name: "date column at 10 minute grain", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_10_MINUTE, want: `toStartOfInterval("d"::DateTime64, INTERVAL 10 minute)::DateTime64`},
		{name: "untyped column at 6 hour grain", dim: untypedCol, grain: runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, want: `toStartOfInterval("ts"::DateTime64, INTERVAL 6 hour)::DateTime64`},
		{name: "timestamp column at 6 hour grain", dim: timestampCol, grain: runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, want: `toStartOfInterval("ts", INTERVAL 6 hour)::DateTime64`},
		{name: "date column at 6 hour grain in a time zone", dim: dateCol, grain: runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, tz: "America/New_York", want: `toStartOfInterval("d"::DateTime64(6, 'America/New_York'), INTERVAL 6 hour)::DateTime64(6, 'America/New_York')`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstDayOfWeek := tt.firstDayOfWeek
			if firstDayOfWeek == 0 {
				firstDayOfWeek = 1
			}
			got, err := DialectClickhouse.DateTruncExpr(tt.dim, tt.grain, tt.tz, firstDayOfWeek, 1)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
