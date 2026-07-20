package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

const (
	// SummaryDimensionsLimit is the maximum number of dimensions to summarize.
	SummaryDimensionsLimit = 15
	// SummaryTimeDimensionsLimit is the maximum number of time dimensions to summarize.
	SummaryTimeDimensionsLimit = 2
	// SummarySampleInterval is the time offset from the max timestamp used for sampling values.
	SummarySampleInterval = 24 * time.Hour
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

	// Compile the security policy's row filter and query filter so the summary only reflects rows the user can access
	securityFilter, securityFilterArgs, err := metricsview.SecurityFilterSQL(e.metricsView, e.security, e.olap.Dialect())
	if err != nil {
		return nil, fmt.Errorf("failed to compile the security policy's filters: %w", err)
	}

	// Gather the categorical and time dimensions
	var dimensions, timeDimensions []*runtimev1.MetricsViewSpec_Dimension
	for _, dim := range e.metricsView.Dimensions {
		// Skip dimensions that the user cannot access first
		if !e.security.CanAccessField(dim.Name) {
			continue
		}

		if dim.Type == runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME {
			timeDimensions = append(timeDimensions, dim)
		} else {
			dimensions = append(dimensions, dim)
		}
	}

	// Add the default time dimension if it wasn't in the list (it's not currently guaranteed to be included in the dimensions list).
	if e.metricsView.TimeDimension != "" {
		var found bool
		for _, dim := range timeDimensions {
			if strings.EqualFold(dim.Name, e.metricsView.TimeDimension) {
				found = true
				break
			}
		}

		if !found {
			// Prepend it so it doesn't get truncated
			head := []*runtimev1.MetricsViewSpec_Dimension{{
				Name:     e.metricsView.TimeDimension,
				DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP},
			}}
			timeDimensions = append(head, timeDimensions...)
		}
	}

	// Apply the default dimension limits
	if len(dimensions) > SummaryDimensionsLimit {
		dimensions = dimensions[0:SummaryDimensionsLimit]
	}
	if len(timeDimensions) > SummaryTimeDimensionsLimit {
		timeDimensions = timeDimensions[0:SummaryTimeDimensionsLimit]
	}

	// Compute the time dimension summaries
	var summaries []DimensionSummary
	var defaultTimeDimensionSummary DimensionSummary
	for _, dim := range timeDimensions {
		timeRange, err := e.summaryTimestamps(ctx, dim.Name, securityFilter, securityFilterArgs)
		if err != nil {
			return nil, fmt.Errorf("failed to get time range for dimension %q: %w", dim.Name, err)
		}
		summary := DimensionSummary{
			Name: dim.Name,
		}
		if dim.DataType != nil {
			summary.DataType = dim.DataType.Code.String()
		}
		if !timeRange.Min.IsZero() {
			summary.MinValue = timeRange.Min
		}
		if !timeRange.Max.IsZero() {
			summary.MaxValue = timeRange.Max
		}
		summaries = append(summaries, summary)
		if strings.EqualFold(dim.Name, e.metricsView.TimeDimension) {
			defaultTimeDimensionSummary = summary
		}
	}

	// Create summaries for normal dimensions
	if len(dimensions) == 0 {
		return &SummaryResult{
			Dimensions:           summaries,
			DefaultTimeDimension: defaultTimeDimensionSummary,
		}, nil
	}

	// Build and execute the summary query
	selectClauses := make([]string, 0, len(dimensions)*4)
	for _, dim := range dimensions {
		expr, err := e.olap.Dialect().MetricsViewDimensionExpression(dim)
		if err != nil {
			return nil, fmt.Errorf("failed to get expression for dimension %s: %w", dim.Name, err)
		}

		// Add select clauses for min, max, has_nulls, and sample value
		dialect := e.olap.Dialect()
		selectClauses = append(selectClauses,
			fmt.Sprintf("%s AS %s", dialect.MinDimensionExpression(expr), dialect.EscapeIdentifier(dim.Name+"__min")),
			fmt.Sprintf("%s AS %s", dialect.MaxDimensionExpression(expr), dialect.EscapeIdentifier(dim.Name+"__max")),
			fmt.Sprintf("COUNT(*) - COUNT(%s) > 0 AS %s", expr, dialect.EscapeIdentifier(dim.Name+"__has_nulls")),
			fmt.Sprintf("ANY_VALUE(%s) AS %s", expr, dialect.EscapeIdentifier(dim.Name+"__example")),
		)
	}

	// Determine the default time dimension expression
	var timeDimExpr string
	for _, dim := range timeDimensions {
		if strings.EqualFold(dim.Name, e.metricsView.TimeDimension) {
			expr, err := e.olap.Dialect().MetricsViewDimensionExpression(dim)
			if err == nil {
				timeDimExpr = expr
			}
			break
		}
	}

	// Create a where clause that applies the SummarySampleInterval to the default time dimension and the security policy's filters
	var whereClauses []string
	var args []any
	if timeDimExpr != "" && defaultTimeDimensionSummary.MaxValue != nil {
		maxTime, _ := defaultTimeDimensionSummary.MaxValue.(time.Time)
		if !maxTime.IsZero() {
			whereClauses = append(whereClauses, fmt.Sprintf("%s >= ?", timeDimExpr))
			args = append(args, maxTime.Add(-SummarySampleInterval))
		}
	}
	if securityFilter != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(%s)", securityFilter))
		args = append(args, securityFilterArgs...)
	}

	// Build the SQL query
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("SELECT ")
	sqlBuilder.WriteString(strings.Join(selectClauses, ", "))
	sqlBuilder.WriteString(" FROM ")
	sqlBuilder.WriteString(e.olap.Dialect().EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table))
	if len(whereClauses) > 0 {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(strings.Join(whereClauses, " AND "))
	}
	sqlBuilder.WriteString(" LIMIT 1")
	sql := sqlBuilder.String()

	// Execute the SQL query
	rows, err := e.olap.Query(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
		QueryAttributes:  e.queryAttributes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute dimension summary query: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("dimension summary query returned no results")
	}

	// Scan the results
	values := make([]any, len(dimensions)*4)
	for i := range values {
		values[i] = new(any)
	}
	if err := rows.Scan(values...); err != nil {
		return nil, fmt.Errorf("failed to scan dimension summary results: %w", err)
	}

	// Build the result summaries
	for i, dim := range dimensions {
		var dataType string
		if dim.DataType != nil {
			dataType = dim.DataType.Code.String()
		}

		index := i * 4
		minValue := *values[index].(*any)
		maxValue := *values[index+1].(*any)
		hasNullsInterface := *values[index+2].(*any)
		hasNulls, ok := hasNullsInterface.(bool)
		if !ok {
			hasNulls = false
		}
		exampleValue := *values[index+3].(*any)

		summaries = append(summaries, DimensionSummary{
			Name:         dim.Name,
			DataType:     dataType,
			MinValue:     minValue,
			MaxValue:     maxValue,
			HasNulls:     hasNulls,
			ExampleValue: exampleValue,
		})
	}

	return &SummaryResult{
		Dimensions:           summaries,
		DefaultTimeDimension: defaultTimeDimensionSummary,
	}, nil
}

// summaryTimestamps resolves the time range for a time dimension, applying the compiled security filter when one is present.
// Without a security filter it delegates to Timestamps;
// with one it mirrors the resolution logic of Timestamps, minus rollups (not used by the summary)
// and minus the executor-level cache (which doesn't account for security filters).
func (e *Executor) summaryTimestamps(ctx context.Context, timeDim, securityFilter string, securityFilterArgs []any) (metricsview.TimestampsResult, error) {
	if securityFilter == "" {
		return e.Timestamps(ctx, timeDim)
	}

	timeExpr, err := e.timeColumnOrExpr(timeDim)
	if err != nil {
		return metricsview.TimestampsResult{}, fmt.Errorf("failed to resolve time column or expression: %w", err)
	}

	if timeDim == e.metricsView.TimeDimension && e.metricsView.DataTimeRange != "" {
		// The declared data_time_range is metadata, not derived from restricted rows, so no filter is needed.
		res, err := e.resolveDeclaredTimestamps(e.metricsView.DataTimeRange)
		if err != nil {
			return metricsview.TimestampsResult{}, fmt.Errorf(`failed to resolve "data_time_range": %w`, err)
		}
		return res, nil
	}

	return e.resolveTimestampsForTable(ctx, e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table, timeExpr, e.metricsView.WatermarkExpression, securityFilter, securityFilterArgs)
}
