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
	DefaultTimeDimension DimensionSummary   `json:"default_time_dimension,omitempty"`
	Dimensions           []DimensionSummary `json:"dimensions"`
}

// DimensionSummary provides statistics for a single dimension in the metrics view.
type DimensionSummary struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	ExampleValue any    `json:"example_value,omitempty"`
	HasNulls     bool   `json:"has_nulls,omitempty"`
	MinValue     any    `json:"min_value,omitempty"`
	MaxValue     any    `json:"max_value,omitempty"`
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
			return nil, fmt.Errorf("failed to get default time range for dimension %q: %w", defaultTimeDimName, err)
		}
	}

	timeDimensionSummaries := make([]DimensionSummary, 0, len(timeDimensions)+1)

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

		summary := DimensionSummary{
			Name:     dim.Name,
			DataType: dataType,
		}

		// Only populate min/max if we have time data
		if !timeRange.Min.IsZero() {
			summary.MinValue = timeRange.Min
		}
		if !timeRange.Max.IsZero() {
			summary.MaxValue = timeRange.Max
		}

		timeDimensionSummaries = append(timeDimensionSummaries, summary)
	}

	// Include default time dimension if it wasn't found in explicit dimensions
	if !defaultTimeDimFound && defaultTimeDimName != "" && e.security.CanAccessField(defaultTimeDimName) {
		summary := DimensionSummary{
			Name:     defaultTimeDimName,
			DataType: "TIMESTAMP", // Assume timestamp for default time dimension
		}

		// Only populate min/max if we have time data
		if !defaultTimeRange.Min.IsZero() {
			summary.MinValue = defaultTimeRange.Min
		}
		if !defaultTimeRange.Max.IsZero() {
			summary.MaxValue = defaultTimeRange.Max
		}

		timeDimensionSummaries = append(timeDimensionSummaries, summary)
	}

	// Create summaries for normal dimensions
	if len(dimensions) == 0 {
		return &SummaryResult{
			Dimensions:           timeDimensionSummaries,
			DefaultTimeDimension: createDefaultTimeDimensionSummary(defaultTimeDimName, defaultTimeRange),
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
			fmt.Sprintf("ANY_VALUE(%s) AS %s", expr, dialect.EscapeIdentifier(dimName+"__example")),
		)
	}

	if len(selectClauses) == 0 {
		if len(failedDimensions) > 0 {
			return nil, fmt.Errorf("failed to get expressions for dimensions: %v", failedDimensions)
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
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("SELECT ")
	sqlBuilder.WriteString(strings.Join(selectClauses, ", "))
	sqlBuilder.WriteString(" FROM ")
	sqlBuilder.WriteString(escapedTableName)
	if whereClause != "" {
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(whereClause)
	}
	sqlBuilder.WriteString(" LIMIT 1")
	sql := sqlBuilder.String()

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
	for i := range values {
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
		hasNullsInterface := *values[index+2].(*interface{})
		hasNulls, ok := hasNullsInterface.(bool)
		if !ok {
			hasNulls = false
		}

		exampleValue := *values[index+3].(*interface{})

		summaries[i] = DimensionSummary{
			Name:         dim.Name,
			DataType:     dataType,
			MinValue:     minValue,
			MaxValue:     maxValue,
			HasNulls:     hasNulls,
			ExampleValue: exampleValue,
		}
	}

	return &SummaryResult{
		Dimensions:           append(timeDimensionSummaries, summaries...),
		DefaultTimeDimension: createDefaultTimeDimensionSummary(defaultTimeDimName, defaultTimeRange),
	}, nil
}

func getDimensionDataType(dim *runtimev1.MetricsViewSpec_Dimension) string {
	if dim.DataType != nil {
		return dim.DataType.Code.String()
	}
	return ""
}

func createDefaultTimeDimensionSummary(defaultTimeDimName string, timeRange TimestampsResult) DimensionSummary {
	summary := DimensionSummary{}

	if defaultTimeDimName != "" {
		summary.Name = defaultTimeDimName
		summary.DataType = "TIMESTAMP"

		// Only populate min/max values if we have time data
		if !timeRange.Min.IsZero() {
			summary.MinValue = timeRange.Min
		}
		if !timeRange.Max.IsZero() {
			summary.MaxValue = timeRange.Max
		}
	}

	return summary
}
