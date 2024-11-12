import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getAllIdentifiers } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
} from "@rilldata/web-common/runtime-client";

export function hasDashboardWhereFilter(dashboardStore: MetricsExplorerEntity) {
  return dashboardStore.whereFilter?.cond?.exprs?.length;
}

export function hasDashboardDimensionThresholdFilter(
  dashboardStore: MetricsExplorerEntity,
) {
  return dashboardStore.dimensionThresholdFilters?.length;
}

export function getExploreFields(
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
 * It removes all state that refers to fields that will be hidden, like filters, pivot chips, and visible field keys.
 * This ensures we do not leak hidden information to the URL recipient.
 */
export function getSanitizedDashboardStateParam(
  dashboard: MetricsExplorerEntity,
  metricsViewFields: string[] | undefined,
): string {
  // If no metrics view fields are specified, everything is visible, and there's no need to sanitize
  if (!metricsViewFields) return getProtoFromDashboardState(dashboard);

  // Else, explicitly add the sanitized state that we want to remember.
  const sanitizedDashboardState = {
    // Remove any measures not specified in the metrics view fields
    visibleMeasureKeys: new Set(
      [...dashboard.visibleMeasureKeys].filter((measure) =>
        metricsViewFields?.includes(measure),
      ),
    ),
    allMeasuresVisible: dashboard.allMeasuresVisible,
    // Remove any dimensions not specified in the metrics view fields
    visibleDimensionKeys: new Set(
      [...dashboard.visibleDimensionKeys].filter((dimension) =>
        metricsViewFields?.includes(dimension),
      ),
    ),
    allDimensionsVisible: dashboard.allDimensionsVisible,
    leaderboardMeasureName: dashboard.leaderboardMeasureName,
    dashboardSortType: dashboard.dashboardSortType,
    sortDirection: dashboard.sortDirection,
    // Remove the where filter
    // whereFilter: dashboard.whereFilter,
    havingFilter: dashboard.havingFilter,
    dimensionThresholdFilters: dashboard.dimensionThresholdFilters,
    dimensionFilterExcludeMode: dashboard.dimensionFilterExcludeMode,
    // There's no need to share filters-in-progress
    // temporaryFilterName: dashboard.temporaryFilterName,
    selectedTimeRange: dashboard.selectedTimeRange,
    selectedScrubRange: dashboard.selectedScrubRange,
    // There's no need to share the user's previous scrub range
    // lastDefinedScrubRange: dashboard.lastDefinedScrubRange,
    selectedComparisonTimeRange: dashboard.selectedComparisonTimeRange,
    // When TDD, we remove the selected comparison dimension (because, if filtered, it's locked & hidden)
    selectedComparisonDimension:
      dashboard.activePage === DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL
        ? undefined
        : dashboard.selectedComparisonDimension,
    // We do not support sharing the dimension table page (because, if filtered, the dimension is locked & hidden)
    activePage:
      dashboard.activePage === DashboardState_ActivePage.DIMENSION_TABLE
        ? DashboardState_ActivePage.DEFAULT
        : dashboard.activePage,
    selectedTimezone: dashboard.selectedTimezone,
    showTimeComparison: dashboard.showTimeComparison,
    leaderboardContextColumn: dashboard.leaderboardContextColumn,
    contextColumnWidths: dashboard.contextColumnWidths,
    selectedDimensionName: dashboard.selectedDimensionName,
    tdd: dashboard.tdd,
    pivot: {
      ...dashboard.pivot,
      rows: {
        dimension: dashboard.pivot.rows.dimension.filter(
          (chip) =>
            metricsViewFields?.includes(chip.id) ||
            chip.type === PivotChipType.Time,
        ),
      },
      columns: {
        measure: dashboard.pivot.columns.measure.filter(
          (chip) =>
            metricsViewFields?.includes(chip.id) ||
            chip.type === PivotChipType.Time,
        ),
        dimension: dashboard.pivot.columns.dimension.filter(
          (chip) =>
            metricsViewFields?.includes(chip.id) ||
            chip.type === PivotChipType.Time,
        ),
      },
    },
  } as MetricsExplorerEntity;

  return getProtoFromDashboardState(sanitizedDashboardState);
}
