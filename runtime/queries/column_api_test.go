package queries

import (
	"context"
	"testing"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func BenchmarkColumnNullCount(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &ColumnNullCount{
			TableName:  "ad_bids",
			ColumnName: "publisher",
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkColumnDescriptiveStatistics(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &ColumnDescriptiveStatistics{
			TableName:  "ad_bids",
			ColumnName: "bid_price",
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkColumnTimeGrain(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &ColumnTimeGrain{
			TableName:  "ad_bids",
			ColumnName: "timestamp",
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkColumnNumericHistogram(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &ColumnNumericHistogram{
			TableName:  "ad_bids",
			ColumnName: "bid_price",
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkColumnRugHistogram(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &ColumnRugHistogram{
			TableName:  "ad_bids",
			ColumnName: "bid_price",
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkColumnTimeRange(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &ColumnTimeRange{
			TableName:  "ad_bids",
			ColumnName: "timestamp",
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkColumnCardinality(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &ColumnCardinality{
			TableName:  "ad_bids",
			ColumnName: "publisher",
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
