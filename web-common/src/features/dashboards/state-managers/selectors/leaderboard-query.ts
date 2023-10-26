import type { QueryServiceMetricsViewComparisonBody } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { prepareSortedQueryBody } from "../../dashboard-utils";
import { activeMeasure, activeMeasureName } from "./active-measure";
import { sortingSelectors } from "./sorting";

/**
 * Returns a function that can be used to get the sorted query body
 * for a leaderboard for the given dimension.
 *
 * If there is no active measure, or if the active measure
 */
function leaderboardSortedQueryBody(
  dashData: DashboardDataSources
):
  | undefined
  | ((dimensionName: string) => QueryServiceMetricsViewComparisonBody) {
  const measure = activeMeasure(dashData);
  if (!measure) {
    return undefined;
  }

  const timeControls = undefined;

  return (dimensionName: string) =>
    prepareSortedQueryBody(
      dimensionName,
      [activeMeasureName(dashData)],
      timeControls,
      sortingSelectors.sortMeasure(dashData),
      sortingSelectors.sortType(dashData),
      sortingSelectors.sortedAscending(dashData),
      undefined
    );
}

export const leaderboardQuerySelectors = { leaderboardSortedQueryBody };
