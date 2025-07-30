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
	if e.olap.Dialect() != drivers.DialectClickHouse || len(ast.comparisonDimFields) == 0 {
		return
	}

	e.wrapClickhouseCompareTimeDimWalk(ast, ast.Root)
}

func (e *Executor) wrapClickhouseCompareTimeDimWalk(a *AST, n *SelectNode) {
	// find node with alias "comparison"
	if n == nil {
		return
	} else if n.Alias == "comparison" {
		timeDim := e.metricsView.TimeDimension
		if a.query.TimeRange != nil && a.query.TimeRange.TimeDimension != "" {
			timeDim = a.query.TimeRange.TimeDimension
		}
		// warp the select node in an outer select if it has a comparison time dimension, but change the name of inner time dimension
		for _, qd := range a.query.Dimensions {
			if a.query.ComparisonTimeRange != nil && qd.Compute != nil && qd.Compute.TimeFloor != nil {
				if strings.EqualFold(qd.Compute.TimeFloor.Dimension, timeDim) {
					// append hash of time dim to the compare name to have diff alias as the column name otherwise clickhouse bypasses the date calculations
					// calculate non crypto hash of the name which should be fast
					hash := crc32.ChecksumIEEE([]byte(qd.Name))
					uniqName := fmt.Sprintf("%s_comp_%x", qd.Name, hash)
					a.wrapSelect(n, a.generateIdentifier())
					for i, f := range n.FromSelect.DimFields {
						if f.Name == qd.Name {
							// change the name of the inner query dimension
							n.FromSelect.DimFields[i].Name = uniqName
						}
					}
					for i, f := range n.DimFields {
						if f.Name == qd.Name {
							// select inner dim in the outer query using uniqName but keep alias as actual query dimension name so change the expression and keep f.Name as is
							n.DimFields[i].Expr = a.sqlForMember(n.FromSelect.Alias, uniqName)
						}
					}
				}
			}
		}
	}
	if n.JoinComparisonSelect != nil {
		e.wrapClickhouseCompareTimeDimWalk(a, n.JoinComparisonSelect)
	}
	if n.FromSelect != nil {
		e.wrapClickhouseCompareTimeDimWalk(a, n.FromSelect)
	}
}
