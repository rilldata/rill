import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getAllIdentifiers } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
} from "@rilldata/web-common/runtime-client";

export function hasDashboardWhereFilter(dashboardStore: MetricsExplorerEntity) {
  return dashboardStore.whereFilter?.cond?.exprs?.length;
}

export function getMetricsViewFields(
  dashboardStore: MetricsExplorerEntity,
  visibleDimensions: MetricsViewSpecDimensionV2[],
  visibleMeasures: MetricsViewSpecMeasureV2[],
) {
  const hasFilter = hasDashboardWhereFilter(dashboardStore);

  const everythingIsVisible =
    dashboardStore.allDimensionsVisible &&
    dashboardStore.allMeasuresVisible &&
    !hasFilter;

  if (everythingIsVisible) return undefined; // Not specifying any fields means all fields are visible

  const filteredDimensions = getAllIdentifiers(dashboardStore.whereFilter);

  return [
    ...visibleDimensions
      .map((dimension) => dimension.name)
      .filter(
        // Hide all dimensions that are filtered
        // Including `!!dimension` fixes a hidden TS error
        (dimension) => !!dimension && !filteredDimensions.includes(dimension),
      ),
    ...visibleMeasures.map((measure) => measure.name),
  ];
}

export function convertDateToMinutes(date: string) {
  const now = new Date();
  const future = new Date(date);
  const diff = future.getTime() - now.getTime();
  return Math.floor(diff / 60000);
}

/**
 * Returns serialized `state` for the current dashboard *without* filters.
 * Removing the filter state ensures that the URL does not leak hidden filters to the URL recipient.
 */
export function getDashboardStateParamWithoutFilters(
  dashboard: MetricsExplorerEntity,
): string {
  const dashboardWithoutFilters = {
    ...dashboard,
    whereFilter: undefined,
    dimensionThresholdFilters: [],
  };

  return getProtoFromDashboardState(dashboardWithoutFilters);
}
