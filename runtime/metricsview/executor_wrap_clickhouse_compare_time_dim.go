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
	}
	if n.Alias == "comparison" {
		timeDim := e.metricsView.TimeDimension
		if a.query.TimeRange != nil && a.query.TimeRange.TimeDimension != "" {
			timeDim = a.query.TimeRange.TimeDimension
		}
		for _, qd := range a.query.Dimensions {
			if a.query.ComparisonTimeRange != nil && qd.Compute != nil && qd.Compute.TimeFloor != nil {
				if strings.EqualFold(qd.Compute.TimeFloor.Dimension, timeDim) {
					dim, err := a.lookupDimension(timeDim, false)
					if err != nil { // this should never happen
						panic(fmt.Errorf("failed to lookup time dimension %q: %w", timeDim, err))
					}
					if !strings.EqualFold(dim.Name, dim.Column) {
						// if the underlying column name is already different from dim name that will be used in the query, we don't need to wrap
						break
					}
					// append hash of time dim to the compare name to have diff alias as the column name otherwise clickhouse bypasses the date calculations
					// calculate non crypto hash of the name which should be fast
					hash := crc32.ChecksumIEEE([]byte(dim.Name))
					uniqName := fmt.Sprintf("%s_comp_%x", dim.Name, hash)
					// warp the select node in an outer select, but change the name of inner time dimension
					a.wrapSelect(n, a.generateIdentifier())
					for i, f := range n.FromSelect.DimFields {
						if f.Name == dim.Name {
							// change the name of the inner query dimension
							n.FromSelect.DimFields[i].Name = uniqName
						}
					}
					for i, f := range n.DimFields {
						if f.Name == dim.Name {
							// select inner dim in the outer query using uniqName but keep alias as actual query dimension name so change the expression and keep f.Name as is
							n.DimFields[i].Expr = a.sqlForMember(n.FromSelect.Alias, uniqName)
						}
					}
					break
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
