import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

/**
 * Extracts the filters of a canvas filter manager, scoped per metrics view.
 * Returns a map of metrics view name to filter expression.
 */
export function getCanvasFilters(
  expressionFilterManager: ExpressionFilterManager,
): Record<string, V1Expression> | undefined {
  const exprByMv = expressionFilterManager.topLevelJoiner.expr;

  // Check if there are any non-empty filters
  const hasFilters = Object.values(exprByMv).some(
    (expr) => expr?.cond?.exprs && expr.cond.exprs.length > 0,
  );

  if (!hasFilters) {
    return undefined;
  }

  return Object.keys(exprByMv).length > 0 ? exprByMv : undefined;
}
