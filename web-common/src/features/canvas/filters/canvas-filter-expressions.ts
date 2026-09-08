import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

/**
 * Extracts filters from a canvas entity, scoped per metrics view.
 * Returns a map of metrics view name to filter expression.
 */
export function getCanvasFilters(
  canvasEntity: CanvasEntity,
): Record<string, V1Expression> | undefined {
  const exprByMv = canvasEntity.expressionFilterManager.topLevelJoiner.expr;

  // Check if there are any non-empty filters
  const hasFilters = Object.values(exprByMv).some(
    (expr) => expr?.cond?.exprs && expr.cond.exprs.length > 0,
  );

  if (!hasFilters) {
    return undefined;
  }

  return Object.keys(exprByMv).length > 0 ? exprByMv : undefined;
}
