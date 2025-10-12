import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";

/**
 * Transforms time and filter store data into partial explore state
 */
export function transformTimeAndFiltersToExploreState(
  timeAndFilterStore: TimeAndFilterStore,
): Partial<ExploreState> {
  const exploreState: Partial<ExploreState> = {};

  if (timeAndFilterStore.where) {
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      timeAndFilterStore.where,
    );
    exploreState.whereFilter = dimensionFilters;
    exploreState.dimensionThresholdFilters = dimensionThresholdFilters;
  }

  if (timeAndFilterStore.timeRangeState) {
    exploreState.selectedTimeRange =
      timeAndFilterStore.timeRangeState.selectedTimeRange;
    exploreState.selectedTimezone =
      timeAndFilterStore?.timeRange?.timeZone || "UTC";

    if (timeAndFilterStore.showTimeComparison) {
      exploreState.showTimeComparison = true;
      exploreState.selectedComparisonTimeRange =
        timeAndFilterStore.comparisonTimeRangeState?.selectedComparisonTimeRange;
    } else {
      exploreState.showTimeComparison = false;
      exploreState.selectedComparisonTimeRange = undefined;
    }
  }

  return exploreState;
}
