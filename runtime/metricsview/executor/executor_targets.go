package executor

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"google.golang.org/protobuf/types/known/structpb"
)

// extractTimeGrainFromQuery extracts the time grain from the query.
// It checks the first dimension with TimeGrain, or falls back to TimeRange.RoundToGrain.
func extractTimeGrainFromQuery(qry *metricsview.Query) metricsview.TimeGrain {
	// Check dimensions for time grain
	for _, dim := range qry.Dimensions {
		if dim.Compute != nil && dim.Compute.TimeFloor != nil {
			return dim.Compute.TimeFloor.Grain
		}
	}

	// Fall back to time range round to grain
	if qry.TimeRange != nil {
		return qry.TimeRange.RoundToGrain
	}

	return metricsview.TimeGrainUnspecified
}

// aggregateTargets aggregates target values based on the query and measure type.
// This is a placeholder for the aggregation logic that will match targets to query results.
func aggregateTargets(
	ctx context.Context,
	targetRows []map[string]any,
	qry *metricsview.Query,
	mv *runtimev1.MetricsViewSpec,
	queryResultData []*structpb.Struct,
	measureName string,
) ([]*structpb.Struct, error) {
	// TODO: Implement target aggregation logic
	// 1. Match targets to query rows based on time period and dimensions
	// 2. Aggregate targets based on measure type (sum/divide vs average/repeat for time, sum vs sum over sum for dimensions)
	// 3. Return target values aligned with query result rows

	// For now, return empty - this will be implemented in the next step
	return []*structpb.Struct{}, nil
}

// getMeasureType returns the measure type for aggregation strategy
func getMeasureType(mv *runtimev1.MetricsViewSpec, measureName string) runtimev1.MetricsViewSpec_MeasureType {
	for _, m := range mv.Measures {
		if m.Name == measureName {
			return m.Type
		}
	}
	return runtimev1.MetricsViewSpec_MEASURE_TYPE_UNSPECIFIED
}

// shouldSumAggregate returns true if the measure type should use sum aggregation for targets
func shouldSumAggregate(measureType runtimev1.MetricsViewSpec_MeasureType) bool {
	switch measureType {
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_UNSPECIFIED:
		// Default to sum for simple measures
		return true
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE:
		return true
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON:
		// Time comparison measures are based on simple measures, so use sum
		return true
	default:
		return false
	}
}

// shouldAverageAggregate returns true if the measure type should use average aggregation for targets
func shouldAverageAggregate(measureType runtimev1.MetricsViewSpec_MeasureType) bool {
	// For now, only derived measures might use average
	// This will need to be refined based on the actual measure expression
	return measureType == runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED
}

// matchTargetToTimePeriod matches a target row to a time period based on the query's time grain
func matchTargetToTimePeriod(targetTime time.Time, queryTime time.Time, grain metricsview.TimeGrain, tz *time.Location) bool {
	// Truncate both times to the grain
	targetTruncated := truncateTimeToGrain(targetTime, grain, tz)
	queryTruncated := truncateTimeToGrain(queryTime, grain, tz)
	return targetTruncated.Equal(queryTruncated)
}

// truncateTimeToGrain truncates a time to the specified grain
func truncateTimeToGrain(t time.Time, grain metricsview.TimeGrain, tz *time.Location) time.Time {
	// Convert to the timezone first
	t = t.In(tz)
	switch grain {
	case metricsview.TimeGrainDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tz)
	case metricsview.TimeGrainWeek:
		// Find the start of the week (Monday)
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday becomes 7
		}
		daysFromMonday := weekday - 1
		return time.Date(t.Year(), t.Month(), t.Day()-daysFromMonday, 0, 0, 0, 0, tz)
	case metricsview.TimeGrainMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, tz)
	case metricsview.TimeGrainQuarter:
		quarter := (int(t.Month()) - 1) / 3
		month := time.Month(quarter*3 + 1)
		return time.Date(t.Year(), month, 1, 0, 0, 0, 0, tz)
	case metricsview.TimeGrainYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, tz)
	default:
		// For other grains, return as-is (they should use timeutil functions)
		return t
	}
}

// matchTargetDimensions matches target dimension values to query filter values
func matchTargetDimensions(targetRow map[string]any, queryDimensions []metricsview.Dimension, queryResultRow *structpb.Struct) bool {
	// TODO: Implement dimension matching logic
	// This should check if target dimension values match the query result row's dimension values
	// Or if target has _null_ for a dimension (indicating aggregate target)
	return true
}

// parseTargetTime parses the time value from a target row
func parseTargetTime(targetRow map[string]any) (time.Time, error) {
	timeVal, ok := targetRow["time"]
	if !ok {
		return time.Time{}, fmt.Errorf("target row missing 'time' field")
	}

	switch v := timeVal.(type) {
	case time.Time:
		return v, nil
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse time: %w", err)
		}
		return t, nil
	default:
		return time.Time{}, fmt.Errorf("unexpected time type: %T", v)
	}
}

