package executor

import (
	"context"

	"github.com/rilldata/rill/runtime/metricsview"
)

// RewriteLookupGroupByKey exposes the lookup_group_by_key rewrite to the executor_test package,
// which can start a test runtime (an internal test can't, because testruntime imports this package).
func (e *Executor) RewriteLookupGroupByKey(ctx context.Context, ast *metricsview.AST) {
	e.rewriteLookupGroupByKey(ctx, ast)
}
