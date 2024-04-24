import type { ResolvedMeasureFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { additionalMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
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
): (
  resolvedMeasureFilter: ResolvedMeasureFilter,
) => QueryServiceMetricsViewComparisonBody {
  return (resolvedMeasureFilter: ResolvedMeasureFilter) => {
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
      measuresForDimensionTable(dashData),
      timeControlsState(dashData),
      sortingSelectors.sortMeasure(dashData),
      sortingSelectors.sortType(dashData),
      sortingSelectors.sortedAscending(dashData),
      filters,
      resolvedMeasureFilter.filter,
      250,
    );
  };
}

function measuresForDimensionTable(dashData: DashboardDataSources) {
  const allMeasures = new Set([
    ...selectedMeasureNames(dashData),
    ...additionalMeasures(dashData),
  ]);
  return [...allMeasures];
}

export function dimensionTableTotalQueryBody(
  dashData: DashboardDataSources,
): (
  resolvedMeasureFilter: ResolvedMeasureFilter,
) => QueryServiceMetricsViewTotalsBody {
  return (resolvedMeasureFilter: ResolvedMeasureFilter) => {
    const dimensionName = dashData.dashboard.selectedDimensionName;
    if (!dimensionName) {
      return {};
    }
    return leaderboardDimensionTotalQueryBody(dashData)(
      dimensionName,
      resolvedMeasureFilter,
    );
  };
}

/**
 * Returns a function that can be used to get the sorted query body
 * for a leaderboard for the given dimension.
 */
export function leaderboardSortedQueryBody(
  dashData: DashboardDataSources,
): (
  dimensionName: string,
  resolvedMeasureFilter: ResolvedMeasureFilter,
) => QueryServiceMetricsViewComparisonBody {
  return (
    dimensionName: string,
    resolvedMeasureFilter: ResolvedMeasureFilter,
  ) =>
    prepareSortedQueryBody(
      dimensionName,
      additionalMeasures(dashData),
      timeControlsState(dashData),
      sortingSelectors.sortMeasure(dashData),
      sortingSelectors.sortType(dashData),
      sortingSelectors.sortedAscending(dashData),
      getFiltersForOtherDimensions(dashData)(dimensionName),
      resolvedMeasureFilter.filter,
      8,
    );
}

export function leaderboardSortedQueryOptions(
  dashData: DashboardDataSources,
): (
  dimensionName: string,
  resolvedMeasureFilter: ResolvedMeasureFilter,
  enabled: boolean,
) => { query: { enabled: boolean } } {
  return (
    dimensionName: string,
    resolvedMeasureFilter: ResolvedMeasureFilter,
    enabled: boolean,
  ) => {
    const sortedQueryEnabled =
      timeControlsState(dashData).ready === true &&
      !!getFiltersForOtherDimensions(dashData)(dimensionName);
    return {
      query: {
        enabled: enabled && sortedQueryEnabled && resolvedMeasureFilter.ready,
      },
    };
  };
}

export function leaderboardDimensionTotalQueryBody(
  dashData: DashboardDataSources,
): (
  dimensionName: string,
  resolvedMeasureFilter: ResolvedMeasureFilter,
) => QueryServiceMetricsViewTotalsBody {
  return (
    dimensionName: string,
    resolvedMeasureFilter: ResolvedMeasureFilter,
  ) => ({
    measureNames: [activeMeasureName(dashData)],
    where: sanitiseExpression(
      getFiltersForOtherDimensions(dashData)(dimensionName),
      resolvedMeasureFilter.filter,
    ),
    timeStart: timeControlsState(dashData).timeStart,
    timeEnd: timeControlsState(dashData).timeEnd,
  });
}

export function leaderboardDimensionTotalQueryOptions(
  dashData: DashboardDataSources,
): (
  dimensionName: string,
  resolvedMeasureFilter: ResolvedMeasureFilter,
) => { query: { enabled: boolean } } {
  return (
    dimensionName: string,
    resolvedMeasureFilter: ResolvedMeasureFilter,
  ) => {
    return {
      query: {
        enabled:
          isAnyMeasureSelected(dashData) &&
          isTimeControlReady(dashData) &&
          !!getFiltersForOtherDimensions(dashData)(dimensionName) &&
          resolvedMeasureFilter.ready,
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
