package executor

import (
	"fmt"
	"hash/crc32"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

// wrapClickhouseComputedTimeDim wraps any select node in the AST having a computed time dimension in an outer select so that inner select has different alias than the time column name and outer select, selects this column using the original time alias.
// Example, in comparison queries we have expression like this (date_trunc('day', "TIME_DIM") - INTERVAL (DATEDIFF('day', base_time_start, compare_time_start)) day) AS "TIME_DIM".
// This does not work correctly in ClickHouse and it just does not subtract the interval from the time dimension and return "TIME_DIM" as it is.
// Another example, if there is an expression like date_trunc('day', "TIME_DIM") AS "TIME_DIM", and if "TIME_DIM" is used in where clause then it will use the underlying "TIME_DIM" column not the truncated one.
// Relevant issue - https://github.com/ClickHouse/ClickHouse/issues/9715
func (e *Executor) wrapClickhouseComputedTimeDim(ast *metricsview.AST) error {
	if e.olap.Dialect() != drivers.DialectClickHouse {
		return nil
	}

	computedTimeDims := make(map[string]string)
	for _, qd := range ast.Query.Dimensions {
		if qd.Compute == nil || qd.Compute.TimeFloor == nil || qd.Compute.TimeFloor.Dimension == "" {
			continue
		}
		dim, err := ast.LookupDimension(qd.Compute.TimeFloor.Dimension, false)
		if err != nil { // this should never happen
			return fmt.Errorf("failed to lookup time dimension %q: %w", qd.Compute.TimeFloor.Dimension, err)
		}
		// if the dimension name is already different from the column name, we don't need to wrap it
		if !strings.EqualFold(dim.Name, dim.Column) {
			continue
		}
		hash := crc32.ChecksumIEEE([]byte(dim.Name))
		uniqName := fmt.Sprintf("%s_uniq_%x", dim.Name, hash)
		computedTimeDims[dim.Name] = uniqName
	}

	if len(computedTimeDims) == 0 {
		// no computed time dimensions found, nothing to wrap
		return nil
	}

	e.wrapClickhouseComputedTimeDimWalk(ast, ast.Root, computedTimeDims)

	return nil
}

func (e *Executor) wrapClickhouseComputedTimeDimWalk(a *metricsview.AST, n *metricsview.SelectNode, computedTimeDims map[string]string) {
	leaf := true
	if n.FromSelect != nil {
		e.wrapClickhouseComputedTimeDimWalk(a, n.FromSelect, computedTimeDims)
		leaf = false
	}
	if n.SpineSelect != nil {
		e.wrapClickhouseComputedTimeDimWalk(a, n.SpineSelect, computedTimeDims)
		leaf = false
	}
	for _, ljs := range n.LeftJoinSelects {
		e.wrapClickhouseComputedTimeDimWalk(a, ljs, computedTimeDims)
		leaf = false
	}
	if n.JoinComparisonSelect != nil {
		e.wrapClickhouseComputedTimeDimWalk(a, n.JoinComparisonSelect, computedTimeDims)
		leaf = false
	}

	// only wrap the inner most select node that has computed time dimensions
	if !leaf {
		return
	}

	wrapNeeded := false
	for _, f := range n.DimFields {
		if _, ok := computedTimeDims[f.Name]; ok {
			wrapNeeded = true
			break
		}
	}

	if wrapNeeded {
		a.WrapSelect(n, a.GenerateIdentifier())
		// first change alias in the inner query
		dims := make([]metricsview.FieldNode, 0, len(n.FromSelect.DimFields))
		for i, f := range n.FromSelect.DimFields {
			dims = append(dims, f)
			if uniqName, ok := computedTimeDims[f.Name]; ok {
				dims[i].Name = uniqName
			}
		}
		n.FromSelect.DimFields = dims

		// check if there are order bys in the inner query and use new alias
		orderBys := make([]metricsview.OrderFieldNode, 0, len(n.FromSelect.OrderBy))
		for i, order := range n.FromSelect.OrderBy {
			orderBys = append(orderBys, order)
			if uniqName, ok := computedTimeDims[order.Name]; ok {
				// change the order by name to the uniqName
				orderBys[i].Name = uniqName
			}
		}
		n.FromSelect.OrderBy = orderBys

		// last change the expressions in the outer query to use the uniqName but keep the return alias same as the actual query dimension name so change the expression and keep f.Name as is
		dims = make([]metricsview.FieldNode, 0, len(n.DimFields))
		for i, f := range n.DimFields {
			dims = append(dims, f)
			if uniqName, ok := computedTimeDims[f.Name]; ok {
				dims[i].Expr = a.Dialect.EscapeMember(n.FromSelect.Alias, uniqName)
			}
		}
		n.DimFields = dims
	}
}
