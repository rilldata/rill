import { get } from "svelte/store";
import type { BusinessModel } from "../business-model/business-model";
import { cancelDashboardQueries } from "../dashboard-queries";
import { metricsExplorerStore } from "../dashboard-stores";

export function clearFilterForDimension(
  ctx: BusinessModel,
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

export function clearAllFilters(ctx: BusinessModel) {
  const filters = get(ctx.dashboardStore).filters;
  const hasFilters =
    (filters && filters.include.length > 0) || filters.exclude.length > 0;
  const metricViewName = get(ctx.metricsViewName);
  if (hasFilters) {
    cancelDashboardQueries(ctx.queryClient, metricViewName);
    metricsExplorerStore.clearFilters(metricViewName);
  }
}

export function toggleDimensionValue(ctx: BusinessModel, event, item) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);
  metricsExplorerStore.toggleFilter(metricViewName, item.name, event.detail);
}

export function toggleFilterMode(ctx: BusinessModel, dimensionName) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);
  metricsExplorerStore.toggleFilterMode(metricViewName, dimensionName);
}
