import {
  createAndExpression,
  negateExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import { get } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";
import { cancelDashboardQueries } from "../dashboard-queries";
import { removeIfExists } from "@rilldata/web-common/lib/arrayUtils";

export function clearFilterForDimension(
  ctx: StateManagers,
  dimensionName: string
) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);
  ctx.updateDashboard((dashboard) => {
    if (!dashboard.whereFilter.cond?.exprs) {
      return;
    }
    removeIfExists(
      dashboard.whereFilter.cond.exprs,
      (e) => e.cond?.exprs?.[0].ident === dimensionName
    );
    dashboard.pinIndex = -1;
  });
}

export function clearAllFilters(ctx: StateManagers) {
  const hasFilters =
    get(ctx.dashboardStore).whereFilter.cond?.exprs?.length ||
    get(ctx.dashboardStore).havingFilter.cond?.exprs?.length;
  const metricViewName = get(ctx.metricsViewName);
  if (hasFilters) {
    cancelDashboardQueries(ctx.queryClient, metricViewName);
    ctx.updateDashboard((dashboard) => {
      dashboard.whereFilter = createAndExpression([]);
      dashboard.havingFilter = createAndExpression([]);
      dashboard.dimensionFilterExcludeMode.clear();
      dashboard.pinIndex = -1;
    });
  }
}
