import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export const toggleSort = (
  metricsExplorer: MetricsExplorerEntity,
  sortType: SortType
) => {
  // if sortType is not provided,  or if it is provided
  // and is the same as the current sort type,
  // then just toggle the current sort direction
  if (
    sortType === undefined ||
    metricsExplorer.dashboardSortType === sortType
  ) {
    metricsExplorer.sortDirection =
      metricsExplorer.sortDirection === SortDirection.ASCENDING
        ? SortDirection.DESCENDING
        : SortDirection.ASCENDING;
  } else {
    // if the sortType is different from the current sort type,
    //  then update the sort type and set the sort direction
    // to descending
    metricsExplorer.dashboardSortType = sortType;
    metricsExplorer.sortDirection = SortDirection.DESCENDING;
  }
};

const contextColumnToSortType = {
  [LeaderboardContextColumn.DELTA_PERCENT]: SortType.DELTA_PERCENT,
  [LeaderboardContextColumn.DELTA_ABSOLUTE]: SortType.DELTA_ABSOLUTE,
  [LeaderboardContextColumn.PERCENT]: SortType.PERCENT,
};

export const toggleSortByActiveContextColumn = (
  metricsExplorer: MetricsExplorerEntity
) => {
  const contextColumnSortType =
    contextColumnToSortType[metricsExplorer.dashboardSortType];
  toggleSort(metricsExplorer, contextColumnSortType);
};

export const sortActions = {
  /**
   * Sets the sort type for the dashboard (value, percent, delta, etc.)
   */
  toggleSort,

  /**
   * Toggles the sort type according to the active context column.
   */
  toggleSortByActiveContextColumn,

  /**
   * Sets the dashboard to be sorted by dimension value.
   * Note that this should only be used in the dimension table
   */
  sortByDimensionValue: (metricsExplorer: MetricsExplorerEntity) =>
    toggleSort(metricsExplorer, SortType.DIMENSION),

  /**
   * Sets the sort direction to descending.
   */
  setSortDescending: (metricsExplorer: MetricsExplorerEntity) => {
    metricsExplorer.sortDirection = SortDirection.DESCENDING;
  },
};
