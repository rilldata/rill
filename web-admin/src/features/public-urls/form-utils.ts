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
): string[] | undefined {
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
  ] as string[];
}

export function convertDateToMinutes(date: string) {
  const now = new Date();
  const future = new Date(date);
  const diff = future.getTime() - now.getTime();
  return Math.floor(diff / 60000);
}

/**
 * Returns the serialized *sanitized* `state` for the current dashboard.
 * It removes any filters and any pivot chips that refer to those filters.
 * This ensures we do not leak hidden filters to the URL recipient.
 */
export function getSanitizedDashboardStateParam(
  dashboard: MetricsExplorerEntity,
  metricsViewFields: string[] | undefined,
): string {
  // If no metrics view fields are specified, everything is visible, and there's no need to sanitize
  if (!metricsViewFields) return getProtoFromDashboardState(dashboard);

  const sanitizedDashboardState = {
    ...dashboard,
    whereFilter: {},
    dimensionThresholdFilters: [],
    pivot: {
      ...dashboard.pivot,
      rows: {
        dimension: dashboard.pivot.rows.dimension.filter((chip) =>
          metricsViewFields?.includes(chip.id),
        ),
      },
      columns: {
        measure: dashboard.pivot.columns.measure.filter((chip) =>
          metricsViewFields?.includes(chip.id),
        ),
        dimension: dashboard.pivot.columns.dimension.filter((chip) =>
          metricsViewFields?.includes(chip.id),
        ),
      },
    },
  } as MetricsExplorerEntity;

  return getProtoFromDashboardState(sanitizedDashboardState);
}
