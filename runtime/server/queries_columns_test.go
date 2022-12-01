package server

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TestServer_GetNullCount(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	res, err := server.GetNullCount(context.Background(), &runtimev1.GetNullCountRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, 1.0, res.Count)

	res, err = server.GetNullCount(context.Background(), &runtimev1.GetNullCountRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 0.0, res.Count)
}

func TestServer_GetDescriptiveStatistics(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	_, err := server.GetDescriptiveStatistics(context.Background(), &runtimev1.GetDescriptiveStatisticsRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	if err != nil {
		// "col" is a varchar column, so this should fail
		require.ErrorContains(t, err, "No function matches the given name and argument types 'approx_quantile(VARCHAR, DECIMAL(3,2))'")
	}

	res, err := server.GetDescriptiveStatistics(context.Background(), &runtimev1.GetDescriptiveStatisticsRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 1.0, res.NumericSummary.GetNumericStatistics().Min)
	require.Equal(t, 5.0, res.NumericSummary.GetNumericStatistics().Max)
	require.Equal(t, 2.2, res.NumericSummary.GetNumericStatistics().Mean)
	require.Equal(t, 1.0, res.NumericSummary.GetNumericStatistics().Q25)
	require.Equal(t, 1.0, res.NumericSummary.GetNumericStatistics().Q50)
	require.Equal(t, 4.0, res.NumericSummary.GetNumericStatistics().Q75)
	require.Equal(t, 1.6, res.NumericSummary.GetNumericStatistics().Sd)
}

func TestServer_EstimateSmallestTimeGrain(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	_, err := server.EstimateSmallestTimeGrain(context.Background(), &runtimev1.EstimateSmallestTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	if err != nil {
		// "val" is a numeric column, so this should fail
		require.ErrorContains(t, err, "Binder Error: No function matches the given name and argument types 'date_part(VARCHAR, INTEGER)'")
	}
	res, err := server.EstimateSmallestTimeGrain(context.Background(), &runtimev1.EstimateSmallestTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "TIME_GRAIN_DAY", res.TimeGrain.String())
}

func TestServer_EstimateSmallestTimeGrain_empty(t *testing.T) {
	server, instanceId := getColumnTestServerWithEmptyModel(t)

	_, err := server.EstimateSmallestTimeGrain(context.Background(), &runtimev1.EstimateSmallestTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	if err != nil {
		// "val" is a numeric column, so this should fail
		require.ErrorContains(t, err, "Binder Error: No function matches the given name and argument types 'date_part(VARCHAR, INTEGER)'")
	}
	res, err := server.EstimateSmallestTimeGrain(context.Background(), &runtimev1.EstimateSmallestTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "TIME_GRAIN_UNSPECIFIED", res.TimeGrain.String())
}

func TestServer_GetNumericHistogram(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	res, err := server.GetNumericHistogram(context.Background(), &runtimev1.GetNumericHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3, len(res.NumericSummary.GetNumericHistogramBins().Bins))
	require.Equal(t, 0, int(res.NumericSummary.GetNumericHistogramBins().Bins[0].Bucket))
	require.Equal(t, 1.0, res.NumericSummary.GetNumericHistogramBins().Bins[0].Low)
	require.Equal(t, 2.333333333333333, res.NumericSummary.GetNumericHistogramBins().Bins[0].High)
	require.Equal(t, 3.0, res.NumericSummary.GetNumericHistogramBins().Bins[0].Count)
}

func TestServer_GetCategoricalHistogram(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	res, err := server.GetRugHistogram(context.Background(), &runtimev1.GetRugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3, len(res.NumericSummary.GetNumericOutliers().Outliers))
	require.Equal(t, 0, int(res.NumericSummary.GetNumericOutliers().Outliers[0].Bucket))
	require.Equal(t, 1.0, res.NumericSummary.GetNumericOutliers().Outliers[0].Low)
	require.Equal(t, 1.008, res.NumericSummary.GetNumericOutliers().Outliers[0].High)
	require.Equal(t, true, res.NumericSummary.GetNumericOutliers().Outliers[0].Present)
	require.True(t, res.NumericSummary.GetNumericOutliers().Outliers[0].Count > 0)

	// works only with numeric columns
	_, err = server.GetRugHistogram(context.Background(), &runtimev1.GetRugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.ErrorContains(t, err, "Conversion Error: Unimplemented type for cast (TIMESTAMP -> DOUBLE)")
}

func TestServer_GetTimeRangeSummary(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	// Get Time Range Summary works with timestamp columns
	res, err := server.GetTimeRangeSummary(context.Background(), &runtimev1.GetTimeRangeSummaryRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, parseTime(t, "2022-11-01T00:00:00Z"), res.TimeRangeSummary.Min)
	require.Equal(t, parseTime(t, "2022-11-03T00:00:00Z"), res.TimeRangeSummary.Max)
	require.Equal(t, int32(0), res.TimeRangeSummary.Interval.Months)
	require.Equal(t, int32(2), res.TimeRangeSummary.Interval.Days)
	require.Equal(t, int64(0), res.TimeRangeSummary.Interval.Micros)
}

func TestServer_GetTimeRangeSummary_empty(t *testing.T) {
	server, instanceId := getColumnTestServerWithEmptyModel(t)

	// Get Time Range Summary works with timestamp columns
	res, err := server.GetTimeRangeSummary(context.Background(), &runtimev1.GetTimeRangeSummaryRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.TimeRangeSummary.Max)

	require.Equal(t, int32(0), res.TimeRangeSummary.Interval.Months)
	require.Equal(t, int32(0), res.TimeRangeSummary.Interval.Days)
	require.Equal(t, int64(0), res.TimeRangeSummary.Interval.Micros)
}

func TestServer_GetTimeRangeSummary_Date_Column(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	// Test Get Time Range Summary with Date type column
	res, err := server.GetTimeRangeSummary(context.Background(), &runtimev1.GetTimeRangeSummaryRequest{InstanceId: instanceId, TableName: "test", ColumnName: "dates"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, parseTime(t, "2007-04-01T00:00:00Z"), res.TimeRangeSummary.Min)
	require.Equal(t, parseTime(t, "2011-06-30T00:00:00Z"), res.TimeRangeSummary.Max)
	require.Equal(t, int32(0), res.TimeRangeSummary.Interval.Months)
	require.Equal(t, int32(1551), res.TimeRangeSummary.Interval.Days)
	require.Equal(t, int64(0), res.TimeRangeSummary.Interval.Micros)
}

func parseTime(tst *testing.T, t string) *timestamppb.Timestamp {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return timestamppb.New(ts)
}

func TestServer_GetCardinalityOfColumn(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	// Get Cardinality of Column works with all columns
	res, err := server.GetCardinalityOfColumn(context.Background(), &runtimev1.GetCardinalityOfColumnRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3.0, res.CategoricalSummary.GetCardinality())

	res, err = server.GetCardinalityOfColumn(context.Background(), &runtimev1.GetCardinalityOfColumnRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3.0, res.CategoricalSummary.GetCardinality())

	res, err = server.GetCardinalityOfColumn(context.Background(), &runtimev1.GetCardinalityOfColumnRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3.0, res.CategoricalSummary.GetCardinality())
}

func getColumnTestServer(t *testing.T) (*Server, string) {
	sql := `
		SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times, DATE '2007-04-01' AS dates
		UNION ALL 
		SELECT 'def' AS col, 5 AS val, TIMESTAMP '2022-11-02 00:00:00' AS times, DATE '2009-06-01' AS dates
		UNION ALL 
		SELECT 'abc' AS col, 3 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-04-11' AS dates
		UNION ALL 
		SELECT null AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-11-21' AS dates
		UNION ALL 
		SELECT 12 AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2011-06-30' AS dates
	`

	return getColumnTestServerWithModel(t, sql, 5)
}

func getColumnTestServerWithEmptyModel(t *testing.T) (*Server, string) {
	sql := `
		SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times, DATE '2007-04-01' AS dates where 1<>1
	`
	return getColumnTestServerWithModel(t, sql, 0)
}

func getColumnTestServerWithModel(t *testing.T, sql string, expectation int) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", sql)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	olap, err := rt.OLAP(context.Background(), instanceID)
	require.NoError(t, err)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) FROM test"})
	require.NoError(t, err)

	defer res.Close()
	var n int
	for res.Next() {
		err := res.Scan(&n)
		require.NoError(t, err)
	}
	require.Equal(t, expectation, n)

	return server, instanceID
}
