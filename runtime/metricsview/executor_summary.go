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

	// statsPerDimension is the number of statistics returned per dimension
	statsPerDimension = 4
	// summaryQueryLimit is the maximum number of rows to return in the summary query
	summaryQueryLimit = 1
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
func (e *Executor) Summary(ctx context.Context, timeDimension string) (*SummaryResult, error) {
	if !e.security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	dimensions := make([]*runtimev1.MetricsViewSpec_Dimension, 0, DefaultSummaryLimit)
	timeDimensions := make([]*runtimev1.MetricsViewSpec_Dimension, 0, 1)

	for _, dim := range e.metricsView.Dimensions {
		// Skip dimensions that the user cannot access first
		if !e.security.CanAccessField(dim.Name) {
			continue
		}

		// If we have reached the limit, stop adding dimensions
		if len(dimensions)+len(timeDimensions) >= DefaultSummaryLimit {
			break
		}

		// If the dimension is the time dimension, handle it separately
		if dim.Name == timeDimension {
			timeDimensions = append(timeDimensions, dim)
		} else {
			dimensions = append(dimensions, dim)
		}
	}

	// Get the default time range for the metrics view
	defaultTimeRange, err := e.Timestamps(ctx, timeDimension)
	if err != nil {
		return nil, fmt.Errorf("failed to get default time range: %w", err)
	}

	timeDimensionSummaries := make([]DimensionSummary, 0, len(timeDimensions))

	// Validate time range
	if defaultTimeRange.Max.IsZero() {
		return &SummaryResult{
			Dimensions: timeDimensionSummaries,
			TimeRange:  defaultTimeRange,
		}, nil
	}

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
			failedDimensions = append(failedDimensions, dimName)
			continue // Skip this dimension instead of failing entirely
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

	timeDimExpr := e.olap.Dialect().EscapeIdentifier(e.metricsView.TimeDimension)
	if e.metricsView.TimeDimension != "" {
		for _, dim := range e.metricsView.Dimensions {
			if dim.Name == e.metricsView.TimeDimension {
				expr, err := e.olap.Dialect().MetricsViewDimensionExpression(dim)
				if err == nil {
					timeDimExpr = expr
				}
				break
			}
		}
	}

	// Use a recent time window with proper timestamp formatting
	dimensionTimeRange := defaultTimeRange.Max.Add(-DefaultSampleTimeWindow)
	whereClause := fmt.Sprintf("WHERE %s >= '%s'", timeDimExpr, dimensionTimeRange.Format(time.RFC3339))

	escapedTableName := e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)
	sql := fmt.Sprintf("SELECT %s FROM %s %s LIMIT %d",
		strings.Join(selectClauses, ", "),
		escapedTableName,
		whereClause,
		summaryQueryLimit,
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

	values := make([]interface{}, len(dimensions)*statsPerDimension)

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

		minValue, maxValue, sampleValue, hasNulls, err := extractDimensionStats(values, i)
		if err != nil {
			return nil, fmt.Errorf("failed to extract stats for dimension %s: %w", dim.Name, err)
		}

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

func extractDimensionStats(values []interface{}, index int) (min, max, sample interface{}, hasNulls bool, err error) {
	baseIndex := index * statsPerDimension

	minPtr, ok := values[baseIndex].(*interface{})
	if !ok {
		return nil, nil, nil, false, fmt.Errorf("invalid min value type")
	}

	maxPtr, ok := values[baseIndex+1].(*interface{})
	if !ok {
		return nil, nil, nil, false, fmt.Errorf("invalid max value type")
	}

	hasNullsPtr, ok := values[baseIndex+2].(*bool)
	if !ok {
		return nil, nil, nil, false, fmt.Errorf("invalid has_nulls type")
	}

	samplePtr, ok := values[baseIndex+3].(*interface{})
	if !ok {
		return nil, nil, nil, false, fmt.Errorf("invalid sample value type")
	}

	return *minPtr, *maxPtr, *samplePtr, *hasNullsPtr, nil
}

func getDimensionDataType(dim *runtimev1.MetricsViewSpec_Dimension) string {
	if dim.DataType != nil {
		return dim.DataType.Code.String()
	}
	return ""
}
