package queries

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/stretchr/testify/require"
)

func BenchmarkMetricsViewsTimeSeries_hourly(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)
	q := &ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		},
		TimeZone: "Asia/Kathmandu",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
		MetricsView: mv,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
func BenchmarkMetricsViewsTimeSeries_daily(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)
	q := &ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
		TimeZone: "Asia/Kathmandu",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
		MetricsView: mv,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsTimeSeries_weekly(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)
	q := &ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		TimeZone: "Asia/Kathmandu",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
		MetricsView: mv,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
func BenchmarkMetricsViewsTimeSeries_weekly_first_day_of_week_Monday(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)

	q := &ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
		MetricsView: mv,
		TimeZone:    "Asia/Kathmandu",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsTimeSeries_weekly_first_day_of_week_Sunday(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)

	q := &ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
		FirstDayOfWeek: 7,
		MetricsView:    mv,
		TimeZone:       "Asia/Kathmandu",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
