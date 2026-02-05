import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getAllIdentifiers } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";

export function hasDashboardWhereFilter(exploreState: ExploreState) {
  return exploreState.whereFilter?.cond?.exprs?.length;
}

export function hasDashboardDimensionThresholdFilter(
  exploreState: ExploreState,
) {
  return exploreState.dimensionThresholdFilters?.length;
}

export function getExploreFields(
  exploreState: ExploreState,
  visibleDimensions: MetricsViewSpecDimension[],
  visibleMeasures: MetricsViewSpecMeasure[],
): string[] | undefined {
  const hasFilter =
    hasDashboardWhereFilter(exploreState) ||
    hasDashboardDimensionThresholdFilter(exploreState);

  const everythingIsVisible =
    exploreState.allDimensionsVisible &&
    exploreState.allMeasuresVisible &&
    !hasFilter;

  if (everythingIsVisible) return undefined; // Not specifying any fields means all fields are visible

  // Check both where and threshold filters for dimensions
  const dimensionsWithThresholdFilters = exploreState.dimensionThresholdFilters
    .filter((dt) => dt.filters.length > 0)
    .map((dt) => dt.name);
  const filteredDimensions = getAllIdentifiers(exploreState.whereFilter).concat(
    dimensionsWithThresholdFilters,
  );

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
export function getSanitizedExploreStateParam(
  exploreState: ExploreState,
  metricsViewFields: string[] | undefined,
  exploreSpec: V1ExploreSpec,
): string {
  // If no metrics view fields are specified, everything is visible, and there's no need to sanitize
  if (!metricsViewFields)
    return getProtoFromDashboardState(exploreState, exploreSpec);

  // Else, explicitly add the sanitized state that we want to remember.
  const sanitizedDashboardState = {
    // Remove any measures not specified in the metrics view fields
    visibleMeasures: exploreState.visibleMeasures.filter((measure) =>
      metricsViewFields?.includes(measure),
    ),
    allMeasuresVisible: exploreState.allMeasuresVisible,
    // Remove any dimensions not specified in the metrics view fields
    visibleDimensions: exploreState.visibleDimensions.filter((dimension) =>
      metricsViewFields?.includes(dimension),
    ),
    allDimensionsVisible: exploreState.allDimensionsVisible,
    leaderboardSortByMeasureName: exploreState.leaderboardSortByMeasureName,
    dashboardSortType: exploreState.dashboardSortType,
    sortDirection: exploreState.sortDirection,

    // Remove the filters
    // whereFilter: dashboard.whereFilter,
    // dimensionThresholdFilters: exploreState.dimensionThresholdFilters,
    // dimensionFilterExcludeMode: exploreState.dimensionFilterExcludeMode,

    // There's no need to share filters-in-progress
    // temporaryFilterName: dashboard.temporaryFilterName,
    selectedTimeRange: exploreState.selectedTimeRange,
    selectedScrubRange: exploreState.selectedScrubRange,
    // There's no need to share the user's previous scrub range
    // lastDefinedScrubRange: dashboard.lastDefinedScrubRange,
    selectedComparisonTimeRange: exploreState.selectedComparisonTimeRange,
    // When TDD, we remove the selected comparison dimension (because, if filtered, it's locked & hidden)
    selectedComparisonDimension:
      exploreState.activePage ===
      DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL
        ? undefined
        : exploreState.selectedComparisonDimension,
    // We do not support sharing the dimension table page (because, if filtered, the dimension is locked & hidden)
    activePage:
      exploreState.activePage === DashboardState_ActivePage.DIMENSION_TABLE
        ? DashboardState_ActivePage.DEFAULT
        : exploreState.activePage,
    selectedTimezone: exploreState.selectedTimezone,
    showTimeComparison: exploreState.showTimeComparison,
    leaderboardContextColumn: exploreState.leaderboardContextColumn,
    contextColumnWidths: exploreState.contextColumnWidths,
    selectedDimensionName: exploreState.selectedDimensionName,
    tdd: exploreState.tdd,
    pivot: {
      ...exploreState.pivot,
      rows: exploreState.pivot.rows.filter(
        (chip) =>
          metricsViewFields?.includes(chip.id) ||
          chip.type === PivotChipType.Time,
      ),
      columns: exploreState.pivot.columns.filter(
        (chip) =>
          metricsViewFields?.includes(chip.id) ||
          chip.type === PivotChipType.Time,
      ),
    },
  } as ExploreState;

  return getProtoFromDashboardState(sanitizedDashboardState, exploreSpec);
}
