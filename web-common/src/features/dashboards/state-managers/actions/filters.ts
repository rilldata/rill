import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";

export function clearAllFilters({ dashboard }: DashboardMutables) {
  const hasFilters =
    dashboard.whereFilter.cond?.exprs?.length ||
    dashboard.havingFilter.cond?.exprs?.length ||
    dashboard.dimensionThresholdFilters?.length;
  if (!hasFilters) {
    return;
  }

  dashboard.whereFilter = createAndExpression([]);
  dashboard.havingFilter = createAndExpression([]);
  dashboard.dimensionThresholdFilters = [];
  dashboard.temporaryFilterName = null;
  dashboard.dimensionFilterExcludeMode.clear();
  dashboard.tdd.pinIndex = -1;
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
