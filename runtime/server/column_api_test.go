package server

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/api"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/stretchr/testify/require"
)

func TestServer_GetTopK(t *testing.T) {
	server, instanceId := getTestServerWithData(t)

	res, err := server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, 3, len(res.TopKResponse.Entries))
	require.Equal(t, "abc", *res.TopKResponse.Entries[0].Value)
	require.Equal(t, 2, int(res.TopKResponse.Entries[0].Count))
	require.Equal(t, "def", *res.TopKResponse.Entries[1].Value)
	require.Equal(t, 1, int(res.TopKResponse.Entries[1].Count))
	require.Nil(t, res.TopKResponse.Entries[2].Value)
	require.Equal(t, 1, int(res.TopKResponse.Entries[2].Count))

	agg := "sum(val)"
	res, err = server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col", Agg: &agg})
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, 3, len(res.TopKResponse.Entries))
	require.Equal(t, "def", *res.TopKResponse.Entries[0].Value)
	require.Equal(t, 5, int(res.TopKResponse.Entries[0].Count))
	require.Equal(t, "abc", *res.TopKResponse.Entries[1].Value)
	require.Equal(t, 4, int(res.TopKResponse.Entries[1].Count))
	require.Nil(t, res.TopKResponse.Entries[2].Value)
	require.Equal(t, 1, int(res.TopKResponse.Entries[2].Count))

	k := int32(1)
	res, err = server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col", K: &k})
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, 1, len(res.TopKResponse.Entries))
	require.Equal(t, "abc", *res.TopKResponse.Entries[0].Value)
	require.Equal(t, 2, int(res.TopKResponse.Entries[0].Count))
}

func TestServer_GetNullCount(t *testing.T) {
	server, instanceId := getTestServerWithData(t)

	res, err := server.GetNullCount(context.Background(), &api.NullCountRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, int64(1), res.Count)

	res, err = server.GetNullCount(context.Background(), &api.NullCountRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(0), res.Count)
}

func TestServer_GetDescriptiveStatistics(t *testing.T) {
	server, instanceId := getTestServerWithData(t)

	_, err := server.GetDescriptiveStatistics(context.Background(), &api.DescriptiveStatisticsRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	if err != nil {
		// "col" is a varchar column, so this should fail
		require.ErrorContains(t, err, "No function matches the given name and argument types 'approx_quantile(VARCHAR, DECIMAL(3,2))'")
	}

	res, err := server.GetDescriptiveStatistics(context.Background(), &api.DescriptiveStatisticsRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 1.0, res.NumericStatistics.Min)
	require.Equal(t, 5.0, res.NumericStatistics.Max)
	require.Equal(t, 2.5, res.NumericStatistics.Mean)
	require.Equal(t, 1.0, res.NumericStatistics.Q25)
	require.Equal(t, 2.0, res.NumericStatistics.Q50)
	require.Equal(t, 4.0, res.NumericStatistics.Q75)
	require.Equal(t, 1.6583123951777, res.NumericStatistics.Sd)
}

func TestServer_EstimateSmallestTimeGrain(t *testing.T) {
	server, instanceId := getTestServerWithData(t)

	_, err := server.EstimateSmallestTimeGrain(context.Background(), &api.EstimateSmallestTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	if err != nil {
		// "val" is a numeric column, so this should fail
		require.ErrorContains(t, err, "Binder Error: No function matches the given name and argument types 'date_part(VARCHAR, INTEGER)'")
	}
	res, err := server.EstimateSmallestTimeGrain(context.Background(), &api.EstimateSmallestTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "DAYS", res.TimeGrain.String())
}

func TestServer_GetNumericHistogram(t *testing.T) {
	server, instanceId := getTestServerWithData(t)

	res, err := server.GetNumericHistogram(context.Background(), &api.NumericHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3, len(res.NumericHistogramBins.Bins))
	require.Equal(t, int64(0), res.NumericHistogramBins.Bins[0].Bucket)
	require.Equal(t, 1.0, res.NumericHistogramBins.Bins[0].Low)
	require.Equal(t, 2.333333333333333, res.NumericHistogramBins.Bins[0].High)
	require.Equal(t, int64(2), res.NumericHistogramBins.Bins[0].Count)
}

func TestServer_GetCategoricalHistogram(t *testing.T) {
	server, instanceId := getTestServerWithData(t)

	res, err := server.GetRugHistogram(context.Background(), &api.RugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3, len(res.NumericOutliers.Outliers))
	require.Equal(t, int64(0), res.NumericOutliers.Outliers[0].Bucket)
	require.Equal(t, 1.0, res.NumericOutliers.Outliers[0].Low)
	require.Equal(t, 1.008, res.NumericOutliers.Outliers[0].High)
	require.Equal(t, true, res.NumericOutliers.Outliers[0].Present)

	// works only with numeric columns
	res, err = server.GetRugHistogram(context.Background(), &api.RugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.ErrorContains(t, err, "Conversion Error: Unimplemented type for cast (TIMESTAMP -> DOUBLE)")
}

func TestServer_GetTimeRangeSummary(t *testing.T) {
	server, instanceId := getTestServerWithData(t)

	// Get Time Range Summary works with timestamp columns
	res, err := server.GetTimeRangeSummary(context.Background(), &api.TimeRangeSummaryRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "2022-11-01 00:00:00 +0000 UTC", res.Min)
	require.Equal(t, "2022-11-03 00:00:00 +0000 UTC", res.Max)
	require.Equal(t, int32(0), res.Interval.Months)
	require.Equal(t, int32(2), res.Interval.Days)
	require.Equal(t, int64(0), res.Interval.Micros)
}

func TestServer_GetCardinalityOfColumn(t *testing.T) {
	server, instanceId := getTestServerWithData(t)

	// Get Cardinality of Column works with all columns
	res, err := server.GetCardinalityOfColumn(context.Background(), &api.CardinalityOfColumnRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(3), *res.Cardinality)

	res, err = server.GetCardinalityOfColumn(context.Background(), &api.CardinalityOfColumnRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(3), *res.Cardinality)

	res, err = server.GetCardinalityOfColumn(context.Background(), &api.CardinalityOfColumnRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(2), *res.Cardinality)
}
