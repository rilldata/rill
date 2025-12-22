import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";

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
  } else if (
    timeAndFilterStore.timeRange &&
    timeAndFilterStore.timeRange.start &&
    timeAndFilterStore.timeRange.end
  ) {
    exploreState.selectedTimeRange = {
      name: TimeRangePreset.CUSTOM,
      interval: timeAndFilterStore.timeGrain,
      start: new Date(timeAndFilterStore.timeRange.start),
      end: new Date(timeAndFilterStore.timeRange.end),
    };
    exploreState.selectedTimezone =
      timeAndFilterStore.timeRange.timeZone || "UTC";
  }

  return exploreState;
}
