import type { QueryServiceMetricsViewComparisonBody } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { prepareSortedQueryBody } from "../../dashboard-utils";
import { activeMeasure } from "./active-measure";
import { sortingSelectors } from "./sorting";

/**
 * Returns a function that can be used to get the sorted query body
 * for a leaderboard for the given dimension.
 *
 * If there is no active measure, or if the active measure
 */
function leaderboardSortedQueryBody(
  selectorArgs: DashboardDataSources
):
  | undefined
  | ((dimensionName: string) => QueryServiceMetricsViewComparisonBody) {
  const measure = activeMeasure(selectorArgs);

  const timeControls = undefined;

  return (dimensionName: string) =>
    prepareSortedQueryBody(
      dimensionName,
      [measure],
      timeControls,
      sortingSelectors.sortMeasure(selectorArgs),
      sortingSelectors.sortType(selectorArgs),
      sortingSelectors.sortedAscending(selectorArgs),
      undefined
    );
}

export const leaderboardQuerySelectors = { leaderboardSortedQueryBody };
