import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { DashboardMutatorFnGeneralArgs } from "./types";

export const toggleSort = (
  { dashboard }: DashboardMutatorFnGeneralArgs,
  sortType: SortType
) => {
  // if sortType is not provided,  or if it is provided
  // and is the same as the current sort type,
  // then just toggle the current sort direction
  if (sortType === undefined || dashboard.dashboardSortType === sortType) {
    dashboard.sortDirection =
      dashboard.sortDirection === SortDirection.ASCENDING
        ? SortDirection.DESCENDING
        : SortDirection.ASCENDING;
  } else {
    // if the sortType is different from the current sort type,
    //  then update the sort type and set the sort direction
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
  args: DashboardMutatorFnGeneralArgs
) => {
  args.cancelQueries();
  const contextColumnSortType =
    contextColumnToSortType[args.dashboard.dashboardSortType];
  toggleSort(args, contextColumnSortType);
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
  sortByDimensionValue: (mutatorArgs: DashboardMutatorFnGeneralArgs) =>
    toggleSort(mutatorArgs, SortType.DIMENSION),

  /**
   * Sets the sort direction to descending.
   */
  setSortDescending: ({ dashboard }: DashboardMutatorFnGeneralArgs) => {
    dashboard.sortDirection = SortDirection.DESCENDING;
  },
};
