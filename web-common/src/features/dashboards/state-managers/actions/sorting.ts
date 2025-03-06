import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { DashboardMutables } from "./types";

export const toggleSort = (
  generalArgs: DashboardMutables,
  sortType: SortType,
  measureName?: string,
) => {
  const { dashboard } = generalArgs;
  if (sortType === dashboard.dashboardSortType) {
    sortActions.toggleSortDirection(generalArgs);
    if (measureName) {
      dashboard.sortedMeasureName = measureName;
    }
  } else {
    dashboard.dashboardSortType = sortType;
    dashboard.sortDirection = SortDirection.DESCENDING;
    if (measureName) {
      dashboard.sortedMeasureName = measureName;
    }
  }
};

const contextColumnToSortType = {
  [LeaderboardContextColumn.DELTA_PERCENT]: SortType.DELTA_PERCENT,
  [LeaderboardContextColumn.DELTA_ABSOLUTE]: SortType.DELTA_ABSOLUTE,
  [LeaderboardContextColumn.PERCENT]: SortType.PERCENT,
};

export const toggleSortByActiveContextColumn = (args: DashboardMutables) => {
  const contextColumnSortType =
    contextColumnToSortType[args.dashboard.leaderboardContextColumn];
  toggleSort(args, contextColumnSortType);
};

export const setSortDescending = ({ dashboard }: DashboardMutables) => {
  dashboard.sortDirection = SortDirection.DESCENDING;
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
  sortByDimensionValue: (mutatorArgs: DashboardMutables) =>
    toggleSort(mutatorArgs, SortType.DIMENSION),

  /**
   * Sets the sort direction to descending.
   */
  setSortDescending,

  /**
   * Toggles the sort direction for the dashboard.
   */
  toggleSortDirection: (generalArgs: DashboardMutables) => {
    const { dashboard } = generalArgs;
    dashboard.sortDirection =
      dashboard.sortDirection === SortDirection.ASCENDING
        ? SortDirection.DESCENDING
        : SortDirection.ASCENDING;
  },
};
