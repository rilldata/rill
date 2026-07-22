import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function clearAllFilters({ dashboard }: DashboardMutables) {
  const hasFilters =
    dashboard.whereFilter.cond?.exprs?.length ||
    dashboard.dimensionThresholdFilters?.length;
  if (!hasFilters) {
    return;
  }

  dashboard.whereFilter = createAndExpression([]);
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

export function setFilter(
  { dashboard }: DashboardMutables,
  expr: V1Expression,
) {
  dashboard.whereFilter = expr;
}

export const filterActions = {
  /**
   * Clears all filters and resets related fields
   */
  clearAllFilters,

  setTemporaryFilterName,

  setFilter,
};
