package executor

import (
	"fmt"

	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteQueryDedupeDimensions collapses query dimensions that share a name, since a result can only hold one column per name.
// The AST relies on a name resolving to a single value, and dialects disagree on two expressions under one alias:
// ClickHouse rejects the query, while DuckDB returns both columns and the second overwrites the first when the rows are read.
//
// A time floor takes precedence over the plain dimension with the same name, since it refines the dimension it is named after.
// The UI requests both when a metrics view's time dimension is also used as a regular dimension, e.g. as the comparison dimension in the time dimension detail view.
// Identical duplicates keep one copy, and different computations under one name are rejected.
func (e *Executor) rewriteQueryDedupeDimensions(qry *metricsview.Query) error {
	byName := make(map[string]int, len(qry.Dimensions))
	dims := make([]metricsview.Dimension, 0, len(qry.Dimensions))
	for _, qd := range qry.Dimensions {
		i, seen := byName[qd.Name]
		if !seen {
			byName[qd.Name] = len(dims)
			dims = append(dims, qd)
			continue
		}
		kept := dims[i]
		switch {
		case kept.Compute == nil && qd.Compute == nil:
			// The same plain dimension again.
		case kept.Compute == nil:
			// The time floor replaces the plain dimension.
			dims[i] = qd
		case qd.Compute == nil:
			// The time floor is already kept.
		case kept.Compute.TimeFloor != nil && qd.Compute.TimeFloor != nil && *kept.Compute.TimeFloor == *qd.Compute.TimeFloor:
			// The same time floor again.
		default:
			return fmt.Errorf("dimension %q is requested multiple times with different computations", qd.Name)
		}
	}
	qry.Dimensions = dims
	return nil
}
