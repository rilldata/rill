import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

/**
 * Extracts filters from a canvas entity, scoped per metrics view.
 * Returns a map of metrics view name to filter expression.
 */
export function getCanvasFilters(
  canvasEntity: CanvasEntity,
): Record<string, V1Expression> | undefined {
  const filtersMap = get(canvasEntity.filterManager.filterMapStore);

  // Check if there are any non-empty filters
  const hasFilters = Array.from(filtersMap.values()).some(
    (expr) => expr?.cond?.exprs && expr.cond.exprs.length > 0,
  );

  if (!hasFilters) {
    return undefined;
  }

  // Convert Map to plain object for API
  const metricsViewFilters: Record<string, V1Expression> = {};
  filtersMap.forEach((expr, metricsViewName) => {
    if (expr?.cond?.exprs && expr.cond.exprs.length > 0) {
      metricsViewFilters[metricsViewName] = expr;
    }
  });

  return Object.keys(metricsViewFilters).length > 0
    ? metricsViewFilters
    : undefined;
}

/**
 * Checks if the canvas has any active filters.
 */
export function hasCanvasFilters(canvasEntity: CanvasEntity): boolean {
  const filtersMap = get(canvasEntity.filterManager.filterMapStore);
  return Array.from(filtersMap.values()).some(
    (expr) => expr?.cond?.exprs && expr.cond.exprs.length > 0,
  );
}

/**
 * Returns the sanitized canvas state from the URL.
 * Removes filter parameters (f and f.*) so locked filters don't appear in the shared URL.
 * This ensures we do not leak hidden filter information to the URL recipient.
 */
export function getSanitizedCanvasStateUrl(currentUrl: URL): string {
  const searchParams = new URLSearchParams(currentUrl.search);
  const filterPrefix: string = ExploreStateURLParams.Filters; // "f"

  // Remove all filter-related parameters (f, f.metricsViewName, etc.)
  const keysToDelete: string[] = [];
  searchParams.forEach((_, key) => {
    if (key === filterPrefix || key.startsWith(`${filterPrefix}.`)) {
      keysToDelete.push(key);
    }
  });
  keysToDelete.forEach((key) => searchParams.delete(key));

  return searchParams.toString();
}
