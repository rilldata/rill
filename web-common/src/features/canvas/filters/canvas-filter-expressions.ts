import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

/**
 * Extracts filters from a canvas entity, scoped per metrics view.
 * Returns a map of metrics view name to filter expression.
 */
export function getCanvasFilters(
  canvasEntity: CanvasEntity,
): Record<string, V1Expression> | undefined {
  // Check if there are any non-empty filters
  const hasFilters = Object.values(
    canvasEntity.expressionFilterManager.managerByMetricsView,
  ).some(
    (manager) =>
      manager.expr?.cond?.exprs && manager.expr.cond.exprs.length > 0,
  );

  if (!hasFilters) {
    return undefined;
  }

  // Convert Map to plain object for API
  const metricsViewFilters: Record<string, V1Expression> = {};
  Object.entries(
    canvasEntity.expressionFilterManager.managerByMetricsView,
  ).forEach(([metricsViewName, manager]) => {
    if (manager.expr?.cond?.exprs && manager.expr?.cond.exprs.length > 0) {
      metricsViewFilters[metricsViewName] = manager.expr;
    }
  });

  return Object.keys(metricsViewFilters).length > 0
    ? metricsViewFilters
    : undefined;
}
