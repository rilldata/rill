import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import { get } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";
import { cancelDashboardQueries } from "../dashboard-queries";

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
