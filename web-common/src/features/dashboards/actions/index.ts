import { get } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";
import { cancelDashboardQueries } from "../dashboard-queries";
import { metricsExplorerStore } from "../dashboard-stores";

export function clearFilterForDimension(
  ctx: StateManagers,
  dimensionId,
  include: boolean
) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);
  metricsExplorerStore.clearFilterForDimension(
    metricViewName,
    dimensionId,
    include
  );
}

export function clearAllFilters(ctx: StateManagers) {
  const filters = get(ctx.dashboardStore).filters;
  const hasFilters =
    (filters && filters.include.length > 0) || filters.exclude.length > 0;
  const metricViewName = get(ctx.metricsViewName);
  if (hasFilters) {
    cancelDashboardQueries(ctx.queryClient, metricViewName);
    metricsExplorerStore.clearFilters(metricViewName);
  }
}

export function toggleDimensionValue(ctx: StateManagers, event, item) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);
  metricsExplorerStore.toggleFilter(metricViewName, item.name, event.detail);
}

export function toggleFilterMode(ctx: StateManagers, dimensionName) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);
  metricsExplorerStore.toggleFilterMode(metricViewName, dimensionName);
}
