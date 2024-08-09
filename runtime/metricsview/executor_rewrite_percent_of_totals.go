package metricsview

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func (e *Executor) rewritePercentOfTotals(ctx context.Context, qry *Query, mv *runtimev1.MetricsViewSpec, sec *runtime.ResolvedSecurity) error {
	var measures []Measure
	var measureIndices []int
	for i, measure := range qry.Measures {
		if measure.Compute != nil && measure.Compute.PercentOfTotal != nil {
			measures = append(measures, Measure{
				Name: measure.Compute.PercentOfTotal.Measure,
			})
			measureIndices = append(measureIndices, i)
		}
	}

	if len(measures) == 0 {
		return nil
	}

	totalsQry := &Query{
		MetricsView: qry.MetricsView,
		Measures:    measures,
		TimeRange:   qry.TimeRange,
		Where:       qry.Where,
		TimeZone:    qry.TimeZone,
	}

	e, err := NewExecutor(ctx, e.rt, e.instanceID, mv, sec, e.priority)
	if err != nil {
		return err
	}
	defer e.Close()

	res, err := e.Query(ctx, totalsQry, nil)
	if err != nil {
		return err
	}
	defer res.Close()

	if !res.Next() {
		return errors.New("query returned no results")
	}

	rowMap := make(map[string]any)
	err = res.MapScan(rowMap)
	if err != nil {
		return err
	}

	for i, measure := range measures {
		t, ok := rowMap[measure.Name]
		if !ok {
			return fmt.Errorf("measure %q didnt return data", measure.Name)
		}
		tf, ok := numberLikeToFloat64(t)
		if !ok {
			return fmt.Errorf("%q is not a number", measure.Name)
		}

		qry.Measures[measureIndices[i]].Compute.PercentOfTotal.Total = tf
	}

	return nil
}

func numberLikeToFloat64(number any) (float64, bool) {
	switch n := number.(type) {
	case *big.Float:
		f, _ := n.Float64()
		return f, true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case *big.Int:
		f, _ := n.Float64()
		return f, true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case int16:
		return float64(n), true
	case int8:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint8:
		return float64(n), true
	case int:
		return float64(n), true
	case uint:
		return float64(n), true
	default:
		return math.NaN(), false
	}
}
