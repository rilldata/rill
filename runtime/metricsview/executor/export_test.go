package executor

import (
	"context"

	"github.com/rilldata/rill/runtime/metricsview"
)

// RewriteQueryForRollupTest exposes rewriteQueryForRollup for integration tests. This is to prevent cyclic dependency error.
func (e *Executor) RewriteQueryForRollupTest(ctx context.Context, qry *metricsview.Query) (string, bool) {
	rw := e.rewriteQueryForRollup(ctx, qry)
	if rw == nil {
		return "", false
	}
	return rw.spec.Table, true
}
