import type {
  QueryServiceMetricsViewComparisonBody,
  QueryServiceMetricsViewTotalsBody,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { prepareSortedQueryBody } from "../../dashboard-utils";
import { activeMeasureName, isAnyMeasureSelected } from "./active-measure";
import { sortingSelectors } from "./sorting";
import { isTimeControlReady, timeControlsState } from "./time-range";
import { getFiltersForOtherDimensions } from "./dimension-filters";

/**
 * Returns a function that can be used to get the sorted query body
 * for a leaderboard for the given dimension.
 */
export function leaderboardSortedQueryBody(
  dashData: DashboardDataSources
): (dimensionName: string) => QueryServiceMetricsViewComparisonBody {
  return (dimensionName: string) =>
    prepareSortedQueryBody(
      dimensionName,
      [activeMeasureName(dashData)],
      timeControlsState(dashData),
      sortingSelectors.sortMeasure(dashData),
      sortingSelectors.sortType(dashData),
      sortingSelectors.sortedAscending(dashData),
      getFiltersForOtherDimensions(dashData)(dimensionName)
    );
}

export function leaderboardSortedQueryOptions(
  dashData: DashboardDataSources
): (dimensionName: string) => { query: { enabled: boolean } } {
  return (dimensionName: string) => {
    const sortedQueryEnabled =
      timeControlsState(dashData).ready === true &&
      !!getFiltersForOtherDimensions(dashData)(dimensionName);
    return {
      query: {
        enabled: sortedQueryEnabled,
      },
    };
  };
}

export function leaderboardDimensionTotalQueryBody(
  dashData: DashboardDataSources
): (dimensionName: string) => QueryServiceMetricsViewTotalsBody {
  return (dimensionName: string) => ({
    measureNames: [activeMeasureName(dashData)],
    filter: getFiltersForOtherDimensions(dashData)(dimensionName),
    timeStart: timeControlsState(dashData).timeStart,
    timeEnd: timeControlsState(dashData).timeEnd,
  });
}

export function leaderboardDimensionTotalQueryOptions(
  dashData: DashboardDataSources
): (dimensionName: string) => { query: { enabled: boolean } } {
  return (dimensionName: string) => {
    return {
      query: {
        enabled:
          isAnyMeasureSelected(dashData) &&
          isTimeControlReady(dashData) &&
          !!getFiltersForOtherDimensions(dashData)(dimensionName),
      },
    };
  };
}

export const leaderboardQuerySelectors = {
  /**
   * Readable containg a function that will return
   * the sorted query body for a leaderboard for the given dimension.
   */
  leaderboardSortedQueryBody,

  /**
   * Readable containg a function that will return
   * the sorted query options for a leaderboard for the given dimension.
   */
  leaderboardSortedQueryOptions,

  /**
   * Readable containg a function that will return
   * the totals query body for a leaderboard for the given dimension.
   */
  leaderboardDimensionTotalQueryBody,

  /**
   * Readable containg a function that will return
   * the totals query options for a leaderboard for the given dimension.
   */
  leaderboardDimensionTotalQueryOptions,
};
