import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
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
 * Extracts the current canvas state from the URL.
 * Canvas state is already encoded in URL parameters (tr, f.*, compare_tr, grain, tz).
 * This function returns the URL state as a string to be passed to the magic auth token API.
 */
export function getCanvasStateUrl(currentUrl: URL): string {
  const searchParams = new URLSearchParams(currentUrl.search);
  return searchParams.toString();
}
