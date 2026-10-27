import { toggleInSelection } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
import { getQueryContextKey } from "@rilldata/web-common/features/dashboards/exact-measures/query-context-key";
import type { DashboardMutables } from "./types";

/**
 * Switches a measure of one surface (see `exploreExactSurface`) between its
 * approximate default and its exact variant for the current query context.
 * Other surfaces are untouched, so the click only re-runs that surface's
 * queries. The selection is tab memory only and resets when the query context
 * changes (see `getQueryContextKey`).
 */
export const toggleExactMeasure = (
  { dashboard }: DashboardMutables,
  surface: string,
  measureName: string,
) => {
  dashboard.exactMeasures = toggleInSelection(
    dashboard.exactMeasures,
    getQueryContextKey(dashboard),
    surface,
    measureName,
  );
};

export const exactMeasureActions = {
  toggleExactMeasure,
};
