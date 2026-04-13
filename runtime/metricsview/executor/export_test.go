package executor

import (
	"context"

	"github.com/rilldata/rill/runtime/metricsview"
)

// RewriteQueryForRollupTest exposes rewriteQueryForRollup for integration tests. This is to prevent cyclic dependency error.
func (e *Executor) RewriteQueryForRollupTest(ctx context.Context, qry *metricsview.Query) (string, bool, error) {
	spec, err := e.rewriteQueryForRollup(ctx, qry)
	if err != nil {
		return "", false, err
	}
	if spec == nil {
		return "", false, nil
	}
	return spec.Table, true, nil
}
