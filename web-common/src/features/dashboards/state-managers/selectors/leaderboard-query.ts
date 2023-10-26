import type { QueryServiceMetricsViewComparisonBody } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { prepareSortedQueryBody } from "../../dashboard-utils";
import { activeMeasureName } from "./active-measure";
import { sortingSelectors } from "./sorting";
import { timeControlsState } from "./time-range";
import { getFiltersForOtherDimensions } from "./dimension-filters";

/**
 * Returns a function that can be used to get the sorted query body
 * for a leaderboard for the given dimension.
 *
 * If there is no active measure, or if the active measure
 */
function leaderboardSortedQueryBody(
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

export const leaderboardQuerySelectors = {
  /**
   * Readable containg a function that will return
   * the sorted query body for a leaderboard for the given dimension.
   */
  leaderboardSortedQueryBody,
};
