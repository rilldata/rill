import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { DashboardMutables } from "./types";
import { setLeaderboardMeasureName } from "./core-actions";

// FIXME: similar to handleDimensionMeasureColumnHeaderClick, consolidate this
export const toggleSort = (
  args: DashboardMutables,
  sortType: SortType,
  measureName?: string,
) => {
  const { dashboard } = args;

  // console.log("[sorting.ts] toggleSort: ", SortType[sortType], measureName);

  // If a measureName is provided that's different from the current one,
  // update it for both value sorts and comparison sorts
  if (
    measureName !== undefined &&
    measureName !== dashboard.leaderboardMeasureName &&
    (sortType === SortType.VALUE ||
      sortType === SortType.DELTA_ABSOLUTE ||
      sortType === SortType.DELTA_PERCENT ||
      sortType === SortType.PERCENT)
  ) {
    setLeaderboardMeasureName(args, measureName);
  }

  // if sortType is not provided, or if it is provided
  // and is the same as the current sort type,
  // then just toggle the current sort direction
  if (sortType === undefined || dashboard.dashboardSortType === sortType) {
    dashboard.sortDirection =
      dashboard.sortDirection === SortDirection.ASCENDING
        ? SortDirection.DESCENDING
        : SortDirection.ASCENDING;
  } else {
    // if the sortType is different from the current sort type,
    // then update the sort type and set the sort direction
    // to descending
    dashboard.dashboardSortType = sortType;
    dashboard.sortDirection = SortDirection.DESCENDING;
  }
};

const contextColumnToSortType = {
  [LeaderboardContextColumn.DELTA_PERCENT]: SortType.DELTA_PERCENT,
  [LeaderboardContextColumn.DELTA_ABSOLUTE]: SortType.DELTA_ABSOLUTE,
  [LeaderboardContextColumn.PERCENT]: SortType.PERCENT,
};

export const toggleSortByActiveContextColumn = (
  args: DashboardMutables,
  measureName?: string,
) => {
  const contextColumnSortType =
    contextColumnToSortType[args.dashboard.leaderboardContextColumn];
  toggleSort(args, contextColumnSortType, measureName);
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
