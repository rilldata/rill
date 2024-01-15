import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  QueryServiceMetricsViewComparisonBody,
  QueryServiceMetricsViewTotalsBody,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { prepareSortedQueryBody } from "../../dashboard-utils";
import {
  activeMeasureName,
  isAnyMeasureSelected,
  selectedMeasureNames,
} from "./active-measure";
import { sortingSelectors } from "./sorting";
import { isTimeControlReady, timeControlsState } from "./time-range";
import { getFiltersForOtherDimensions } from "./dimension-filters";
import { updateFilterOnSearch } from "../../dimension-table/dimension-table-utils";
import { dimensionTableSearchString } from "./dimension-table";

/**
 * Returns the sorted query body for the dimension table for the
 * active dimension.
 *
 * Safety: this readable should ONLY be used when the active dimension
 * is not undefined, ie then the dimension table is visible.
 */
export function dimensionTableSortedQueryBody(
  dashData: DashboardDataSources,
): QueryServiceMetricsViewComparisonBody {
  const dimensionName = dashData.dashboard.selectedDimensionName;
  if (!dimensionName) {
    return {};
  }
  let filters = getFiltersForOtherDimensions(dashData)(dimensionName);
  const searchString = dimensionTableSearchString(dashData);
  if (searchString !== undefined) {
    filters = updateFilterOnSearch(filters, searchString, dimensionName);
  }

  return prepareSortedQueryBody(
    dimensionName,
    selectedMeasureNames(dashData),
    timeControlsState(dashData),
    sortingSelectors.sortMeasure(dashData),
    sortingSelectors.sortType(dashData),
    sortingSelectors.sortedAscending(dashData),
    filters,
  );
}

export function dimensionTableTotalQueryBody(
  dashData: DashboardDataSources,
): QueryServiceMetricsViewTotalsBody {
  const dimensionName = dashData.dashboard.selectedDimensionName;
  if (!dimensionName) {
    return {};
  }
  return leaderboardDimensionTotalQueryBody(dashData)(dimensionName);
}

/**
 * Returns a function that can be used to get the sorted query body
 * for a leaderboard for the given dimension.
 */
export function leaderboardSortedQueryBody(
  dashData: DashboardDataSources,
): (dimensionName: string) => QueryServiceMetricsViewComparisonBody {
  return (dimensionName: string) =>
    prepareSortedQueryBody(
      dimensionName,
      [activeMeasureName(dashData)],
      timeControlsState(dashData),
      sortingSelectors.sortMeasure(dashData),
      sortingSelectors.sortType(dashData),
      sortingSelectors.sortedAscending(dashData),
      getFiltersForOtherDimensions(dashData)(dimensionName),
    );
}

export function leaderboardSortedQueryOptions(
  dashData: DashboardDataSources,
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
  dashData: DashboardDataSources,
): (dimensionName: string) => QueryServiceMetricsViewTotalsBody {
  return (dimensionName: string) => ({
    measureNames: [activeMeasureName(dashData)],
    where: sanitiseExpression(
      getFiltersForOtherDimensions(dashData)(dimensionName),
    ),
    timeStart: timeControlsState(dashData).timeStart,
    timeEnd: timeControlsState(dashData).timeEnd,
  });
}

export function leaderboardDimensionTotalQueryOptions(
  dashData: DashboardDataSources,
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
   * Readable containing the sorted query body for the dimension table.
   */
  dimensionTableSortedQueryBody,

  /**
   * Readable containing the totals query body for the dimension table.
   */
  dimensionTableTotalQueryBody,

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
