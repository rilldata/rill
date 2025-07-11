package metricsview

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type SummaryResult struct {
	Metadata   map[string]any     `json:"metadata"`
	TimeRange  TimestampsResult   `json:"time_range"`
	Dimensions []DimensionSummary `json:"dimensions"`
}

type DimensionSummary struct {
	Name          string `json:"name"`
	DataType      string `json:"data_type"`
	SampleValues  []any  `json:"sample_values"`
	NullCount     int64  `json:"null_count"`
	DistinctCount int64  `json:"distinct_count"`
}

// Summary provides statistics for all dimensions and measures in the metrics view.
func (e *Executor) Summary(ctx context.Context, timeDimension string) (*SummaryResult, error) {
	if !e.security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	dims := []*runtimev1.MetricsViewSpec_Dimension{}
	for _, dim := range e.metricsView.Dimensions {
		if e.security.CanAccessField(dim.Name) {
			dims = append(dims, dim)
		}
	}

	dimensionSummaries := make([]DimensionSummary, 0, len(dims))
	for _, dim := range dims {
		dimSummary, err := e.dimensionSummary(ctx, dim)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze dimension %s: %w", dim.Name, err)
		}
		dimensionSummaries = append(dimensionSummaries, *dimSummary)
	}

	totalRows, err := e.count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get row count: %w", err)
	}

	metadata := map[string]any{
		"total_rows":       totalRows,
		"dimensions_count": len(dimensionSummaries),
	}

	timeRange, err := e.Timestamps(ctx, timeDimension)
	if err != nil {
		return nil, fmt.Errorf("failed to get time range: %w", err)
	}

	return &SummaryResult{
		Metadata:   metadata,
		TimeRange:  timeRange,
		Dimensions: dimensionSummaries,
	}, nil
}

// dimensionSummary retrieves statistics for a single dimension in the metrics view.
func (e *Executor) dimensionSummary(ctx context.Context, dim *runtimev1.MetricsViewSpec_Dimension) (*DimensionSummary, error) {
	dimExpr, err := e.olap.Dialect().MetricsViewDimensionExpression(dim)
	if err != nil {
		return nil, fmt.Errorf("failed to get dimension expression for %s: %w", dim.Name, err)
	}

	query := fmt.Sprintf(`SELECT COUNT(*) as total_count, COUNT(%s) as non_null_count, COUNT(DISTINCT %s) as distinct_count FROM %s`, dimExpr, dimExpr, e.metricsView.Table)

	res, err := e.olap.Query(ctx, &drivers.Statement{Query: query, Priority: e.priority})
	if err != nil {
		return nil, fmt.Errorf("failed to query dimension statistics for %s: %w", dim.Name, err)
	}
	defer res.Close()

	var totalCount, nonNullCount, distinctCount int64
	if res.Next() {
		if err := res.Scan(&totalCount, &nonNullCount, &distinctCount); err != nil {
			return nil, fmt.Errorf("failed to scan dimension statistics for %s: %w", dim.Name, err)
		}
	}
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("error reading dimension statistics for %s: %w", dim.Name, err)
	}

	coreSample := fmt.Sprintf("SELECT DISTINCT CAST(%s AS STRING) AS sample_value FROM %s WHERE %s IS NOT NULL LIMIT 5", dimExpr, e.metricsView.Table, dimExpr)

	coreSampleRes, err := e.olap.Query(ctx, &drivers.Statement{Query: coreSample, Priority: e.priority})
	if err != nil {
		return nil, fmt.Errorf("failed to query sample values for dimension %s: %w", dim.Name, err)
	}
	defer coreSampleRes.Close()

	var sampleValues []any
	for coreSampleRes.Next() {
		var value any
		if err := coreSampleRes.Scan(&value); err != nil {
			return nil, fmt.Errorf("failed to scan sample value for dimension %s: %w", dim.Name, err)
		}
		sampleValues = append(sampleValues, value)
	}
	if err := coreSampleRes.Err(); err != nil {
		return nil, fmt.Errorf("error reading sample values for dimension %s: %w", dim.Name, err)
	}

	nullCount := totalCount - nonNullCount

	return &DimensionSummary{
		Name:          dim.Name,
		DataType:      dim.DataType.Code.String(),
		SampleValues:  sampleValues,
		NullCount:     nullCount,
		DistinctCount: distinctCount,
	}, nil
}
