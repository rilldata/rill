import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";

export function clearAllFilters({
  dashboard,
  cancelQueries,
}: DashboardMutables) {
  const hasFilters =
    dashboard.whereFilter.cond?.exprs?.length ||
    dashboard.havingFilter.cond?.exprs?.length;
  if (!hasFilters) {
    return;
  }

  cancelQueries();

  dashboard.whereFilter = createAndExpression([]);
  dashboard.havingFilter = createAndExpression([]);
  dashboard.temporaryFilterName = null;
  dashboard.dimensionFilterExcludeMode.clear();
  dashboard.pinIndex = -1;
}

export function setTemporaryFilterName(
  { dashboard }: DashboardMutables,
  name: string,
) {
  dashboard.temporaryFilterName = name;
}

export const filterActions = {
  /**
   * Clears all filters and resets related fields
   */
  clearAllFilters,

  setTemporaryFilterName,
};
