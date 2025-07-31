package metricsview

import (
	"fmt"
	"hash/crc32"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// wrapClickhouseCompareTimeDim wraps the comparison AST if it has computed time dimension to have different alias than the time column name.
// Example, in comparison queries we have expression like this: (date_trunc('day', "TIME_DIM") - INTERVAL (DATEDIFF('day', base_time_start, compare_time_start)) day) AS "TIME_DIM"
// This does not work correctly in ClickHouse and it just does not subtract the interval from the time dimension and return "TIME_DIM" as it is.
func (e *Executor) wrapClickhouseCompareTimeDim(ast *AST) {
	if e.olap.Dialect() != drivers.DialectClickHouse || ast.query.ComparisonTimeRange == nil {
		return
	}

	timeDim := e.metricsView.TimeDimension
	if ast.query.TimeRange != nil && ast.query.TimeRange.TimeDimension != "" {
		timeDim = ast.query.TimeRange.TimeDimension
	}
	computedDims := make(map[string]string)
	for _, qd := range ast.query.Dimensions {
		if qd.Compute == nil || qd.Compute.TimeFloor == nil || !strings.EqualFold(qd.Compute.TimeFloor.Dimension, timeDim) {
			continue
		}
		dim, err := ast.lookupDimension(timeDim, false)
		if err != nil { // this should never happen
			panic(fmt.Errorf("failed to lookup time dimension %q: %w", timeDim, err))
		}
		if !strings.EqualFold(dim.Name, dim.Column) {
			continue
		}
		hash := crc32.ChecksumIEEE([]byte(dim.Name))
		uniqName := fmt.Sprintf("%s_uniq_%x", dim.Name, hash)
		computedDims[dim.Name] = uniqName
	}

	if len(computedDims) == 0 {
		// no computed time dimensions found, nothing to wrap
		return
	}

	e.wrapClickhouseCompareTimeDimWalk(ast, ast.Root, computedDims)
}

func (e *Executor) wrapClickhouseCompareTimeDimWalk(a *AST, n *SelectNode, computedDims map[string]string) {
	wrapNeeded := false
	for _, f := range n.DimFields {
		if _, ok := computedDims[f.Name]; ok {
			wrapNeeded = true
			break
		}
	}

	if wrapNeeded {
		a.wrapSelect(n, a.generateIdentifier())
		for i, f := range n.DimFields {
			if uniqName, ok := computedDims[f.Name]; ok {
				// change the name of the inner query dimension
				n.FromSelect.DimFields[i].Name = uniqName
			}
		}
		for i, f := range n.DimFields {
			if uniqName, ok := computedDims[f.Name]; ok {
				// select inner dim in the outer query using uniqName but keep alias as actual query dimension name so change the expression and keep f.Name as is
				n.DimFields[i].Expr = a.sqlForMember(n.FromSelect.Alias, uniqName)

			}
		}
	}

	if n.FromSelect != nil {
		e.wrapClickhouseCompareTimeDimWalk(a, n.FromSelect, computedDims)
	}
	if n.JoinComparisonSelect != nil {
		e.wrapClickhouseCompareTimeDimWalk(a, n.JoinComparisonSelect, computedDims)
	}
}
