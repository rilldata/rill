package metricsview

import (
	"context"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

const (
	// DefaultSummaryLimit is the maximum number of dimensions to summarize.
	DefaultSummaryLimit = 15
	// DefaultSampleTimeWindow is the time window used for sampling values
	DefaultSampleTimeWindow = 24 * time.Hour
)

// SummaryResult is the statistics for a single metrics view.
type SummaryResult struct {
	Dimensions []DimensionSummary `json:"dimensions"`
	TimeRange  TimestampsResult   `json:"time_range"`
}

// DimensionSummary provides statistics for a single dimension in the metrics view.
type DimensionSummary struct {
	Name      string           `json:"name"`
	DataType  string           `json:"data_type"`
	Value     any              `json:"value,omitempty"`
	HasNulls  bool             `json:"has_nulls,omitempty"`
	MinValue  any              `json:"min_value,omitempty"`
	MaxValue  any              `json:"max_value,omitempty"`
	TimeRange TimestampsResult `json:"time_range,omitempty"`
}

// Summary provides statistics for all dimensions and measures in the metrics view.
func (e *Executor) Summary(ctx context.Context) (*SummaryResult, error) {
	if !e.security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	dimensions := make([]*runtimev1.MetricsViewSpec_Dimension, 0, DefaultSummaryLimit)
	timeDimensions := make([]*runtimev1.MetricsViewSpec_Dimension, 0, 2)

	// Track the default time dimension name to ensure it's included
	defaultTimeDimName := e.metricsView.TimeDimension

	for _, dim := range e.metricsView.Dimensions {
		// Skip dimensions that the user cannot access first
		if !e.security.CanAccessField(dim.Name) {
			continue
		}

		// If we have reached the limit, stop adding dimensions
		if len(dimensions)+len(timeDimensions) >= DefaultSummaryLimit {
			break
		}

		// Check if this dimension is a time dimension (has timestamp data type)
		isTimeDimension := false
		if dim.DataType != nil {
			switch dim.DataType.Code {
			case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME:
				isTimeDimension = true
			}
		}

		if isTimeDimension {
			timeDimensions = append(timeDimensions, dim)
		} else {
			dimensions = append(dimensions, dim)
		}
	}

	// Ensure the default time dimension is included even if it's not in the dimensions list
	defaultTimeDimFound := false
	if defaultTimeDimName != "" {
		for _, dim := range timeDimensions {
			if dim.Name == defaultTimeDimName {
				defaultTimeDimFound = true
				break
			}
		}
	}

	// Get the default time range for the metrics view using the default time dimension
	// This will be used for time range filtering in other queries
	var defaultTimeRange TimestampsResult
	var err error
	if defaultTimeDimName != "" {
		defaultTimeRange, err = e.Timestamps(ctx, defaultTimeDimName)
		if err != nil {
			return nil, fmt.Errorf("failed to get default time range: %w", err)
		}
	}

	timeDimensionSummaries := make([]DimensionSummary, 0, len(timeDimensions)+1)

	// Validate time range
	if defaultTimeRange.Max.IsZero() {
		// If we have no time data, still process time dimensions we found
		for _, dim := range timeDimensions {
			var dataType string
			if dim.DataType != nil {
				dataType = dim.DataType.Code.String()
			}

			timeDimensionSummaries = append(timeDimensionSummaries, DimensionSummary{
				Name:     dim.Name,
				DataType: dataType,
			})
		}

		// Include default time dimension if it wasn't found in explicit dimensions
		if !defaultTimeDimFound && defaultTimeDimName != "" && e.security.CanAccessField(defaultTimeDimName) {
			timeDimensionSummaries = append(timeDimensionSummaries, DimensionSummary{
				Name:     defaultTimeDimName,
				DataType: "TIMESTAMP", // Assume timestamp for default time dimension
			})
		}

		return &SummaryResult{
			Dimensions: timeDimensionSummaries,
			TimeRange:  defaultTimeRange,
		}, nil
	}

	// Process each time dimension found in the dimensions list
	for _, dim := range timeDimensions {
		timeRange, err := e.Timestamps(ctx, dim.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get time range for dimension %s: %w", dim.Name, err)
		}

		var dataType string
		if dim.DataType != nil {
			dataType = dim.DataType.Code.String()
		}

		timeDimensionSummaries = append(timeDimensionSummaries, DimensionSummary{
			Name:      dim.Name,
			TimeRange: timeRange,
			DataType:  dataType,
		})
	}

	// Include default time dimension if it wasn't found in explicit dimensions
	if !defaultTimeDimFound && defaultTimeDimName != "" && e.security.CanAccessField(defaultTimeDimName) {
		timeDimensionSummaries = append(timeDimensionSummaries, DimensionSummary{
			Name:      defaultTimeDimName,
			TimeRange: defaultTimeRange,
			DataType:  "TIMESTAMP", // Assume timestamp for default time dimension
		})
	}

	// Create summaries for normal dimensions
	if len(dimensions) == 0 {
		return &SummaryResult{
			Dimensions: timeDimensionSummaries,
			TimeRange:  defaultTimeRange,
		}, nil
	}

	// Build and execute the summary query
	selectClauses := make([]string, 0, len(dimensions)*4)
	var failedDimensions []string

	for _, dim := range dimensions {
		dimName := dim.Name
		expr, err := e.olap.Dialect().MetricsViewDimensionExpression(dim)
		if err != nil {
			return nil, fmt.Errorf("failed to get expression for dimension %s: %w", dimName, err)
		}

		dialect := e.olap.Dialect()

		// Add select clauses for min, max, has_nulls, and sample value
		selectClauses = append(selectClauses,
			fmt.Sprintf("MIN(%s) AS %s", expr, dialect.EscapeIdentifier(dimName+"__min")),
			fmt.Sprintf("MAX(%s) AS %s", expr, dialect.EscapeIdentifier(dimName+"__max")),
			fmt.Sprintf("COUNT(*) - COUNT(%s) > 0 AS %s", expr, dialect.EscapeIdentifier(dimName+"__has_nulls")),
			fmt.Sprintf("ANY_VALUE(%s) AS %s", expr, dialect.EscapeIdentifier(dimName+"__sample")),
		)
	}

	if len(selectClauses) == 0 {
		if len(failedDimensions) > 0 {
			return nil, fmt.Errorf("failed to get expressions for all dimensions: %v", failedDimensions)
		}
		return nil, fmt.Errorf("no dimensions to summarize")
	}

	var timeDimExpr string
	if defaultTimeDimName != "" {
		timeDimExpr = e.olap.Dialect().EscapeIdentifier(defaultTimeDimName)
		for _, dim := range e.metricsView.Dimensions {
			if dim.Name == defaultTimeDimName {
				expr, err := e.olap.Dialect().MetricsViewDimensionExpression(dim)
				if err == nil {
					timeDimExpr = expr
				}
				break
			}
		}
	}

	// Use a recent time window with proper timestamp formatting
	// Use the default time dimension for filtering to ensure it's likely indexed
	var whereClause string
	if defaultTimeDimName != "" && !defaultTimeRange.Max.IsZero() {
		dimensionTimeRange := defaultTimeRange.Max.Add(-DefaultSampleTimeWindow)
		whereClause = fmt.Sprintf("WHERE %s >= '%s'", timeDimExpr, dimensionTimeRange.Format(time.RFC3339))
	} else {
		whereClause = "" // No time filtering if no default time dimension
	}

	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)
	sql := fmt.Sprintf("SELECT %s FROM %s %s LIMIT %d",
		strings.Join(selectClauses, ", "),
		escapedTableName,
		whereClause,
		1,
	)

	rows, err := e.olap.Query(ctx, &drivers.Statement{
		Query:            sql,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute dimension summary query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("dimension summary query returned no results")
	}

	values := make([]interface{}, len(dimensions)*4)

	// Create scan targets
	for i := 0; i < len(values); i++ {
		values[i] = new(interface{})
	}

	if err := rows.Scan(values...); err != nil {
		return nil, fmt.Errorf("failed to scan dimension summary results: %w", err)
	}

	// Build the result summaries
	summaries := make([]DimensionSummary, len(dimensions))

	for i, dim := range dimensions {
		dataType := getDimensionDataType(dim)

		index := i * 4
		minValue := *values[index].(*interface{})
		maxValue := *values[index+1].(*interface{})
		hasNulls := *values[index+2].(*bool)
		sampleValue := *values[index+3].(*interface{})

		summaries[i] = DimensionSummary{
			Name:     dim.Name,
			DataType: dataType,
			MinValue: minValue,
			MaxValue: maxValue,
			HasNulls: hasNulls,
			Value:    sampleValue,
		}
	}

	return &SummaryResult{
		Dimensions: append(timeDimensionSummaries, summaries...),
		TimeRange:  defaultTimeRange,
	}, nil
}

func getDimensionDataType(dim *runtimev1.MetricsViewSpec_Dimension) string {
	if dim.DataType != nil {
		return dim.DataType.Code.String()
	}
	return ""
}
