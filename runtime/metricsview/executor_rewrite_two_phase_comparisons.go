package metricsview

import (
	"context"
	"fmt"
)

// rewriteTwoPhaseComparisons rewrites the query to query base time range first and then use the results to query the comparison time range.
func (e *Executor) rewriteTwoPhaseComparisons(ctx context.Context, qry *Query) error {
	// Check if it's enabled.
	if !e.instanceCfg.MetricsTwoPhaseComparisons {
		return nil
	}

	// Skip if the criteria for a two-phase comparison are not met.
	if qry.ComparisonTimeRange == nil || len(qry.Sort) != 1 || len(qry.Dimensions) == 0 {
		return nil
	}

	// Build a query for the base time range
	baseQry := &Query{
		MetricsView:         qry.MetricsView,
		Dimensions:          qry.Dimensions,
		Measures:            nil,
		PivotOn:             qry.PivotOn,
		Spine:               qry.Spine,
		Sort:                qry.Sort,
		TimeRange:           qry.TimeRange,
		ComparisonTimeRange: nil,
		Where:               qry.Where,
		Having:              qry.Having,
		Limit:               qry.Limit,
		Offset:              qry.Offset,
		TimeZone:            qry.TimeZone,
		UseDisplayNames:     false,
	}

	// only keep base measures and remove derived measures
	var bm []Measure
	for _, m := range qry.Measures {
		if m.Compute == nil {
			bm = append(bm, m)
		} else if m.Compute.ComparisonValue == nil && m.Compute.ComparisonDelta == nil && m.Compute.ComparisonRatio == nil {
			bm = append(bm, m)
		}
	}
	baseQry.Measures = bm

	// Execute the query for the base time range
	baseRes, err := e.Query(ctx, baseQry, nil)
	if err != nil {
		return err
	}
	defer baseRes.Close()

	values := make([]any, len(baseRes.Schema.Fields))
	valuePtrs := make([]any, len(baseRes.Schema.Fields))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Extract the dimension and measure values returned from the inner query.
	dims := make(map[any][]any)
	measures := make(map[any][]any)
	for baseRes.Next() {
		if err := baseRes.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("two phase comparison: base query failed to scan value: %w", err)
		}
		i := 0
		for _, d := range baseQry.Dimensions {
			dims[d.Name] = append(dims[d.Name], values[i])
			i++
		}
		for _, m := range baseQry.Measures {
			measures[m.Name] = append(measures[m.Name], values[i])
			i++
		}
	}

	qry.inlineBaseSelect = true
	qry.inlineDims = dims
	qry.inlineMeasures = measures

	return nil
}
