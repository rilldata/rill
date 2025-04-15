import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import { setLeaderboardSortByMeasureName } from "./leaderboard";
import type { DashboardMutables } from "./types";

export const isValueBasedSort = (sortType: SortType): boolean => {
  return (
    sortType === SortType.VALUE ||
    sortType === SortType.DELTA_ABSOLUTE ||
    sortType === SortType.DELTA_PERCENT ||
    sortType === SortType.PERCENT
  );
};

export const toggleSortDirection = (
  currentDirection: SortDirection,
): SortDirection => {
  return currentDirection === SortDirection.ASCENDING
    ? SortDirection.DESCENDING
    : SortDirection.ASCENDING;
};

export const toggleSort = (
  args: DashboardMutables,
  sortType: SortType,
  measureName?: string,
) => {
  const { dashboard } = args;

  // Handle measure name change if provided
  if (
    measureName !== undefined &&
    measureName !== dashboard.leaderboardSortByMeasureName &&
    isValueBasedSort(sortType)
  ) {
    setLeaderboardSortByMeasureName(args, measureName);
    dashboard.dashboardSortType = sortType;
    dashboard.sortDirection = SortDirection.DESCENDING;
    return;
  }

  // Handle sort type and direction changes
  if (sortType === undefined || dashboard.dashboardSortType === sortType) {
    dashboard.sortDirection = toggleSortDirection(dashboard.sortDirection);
  } else {
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

export const sortByDimensionValue = (args: DashboardMutables) => {
  toggleSort(args, SortType.DIMENSION);
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
  sortByDimensionValue,

  /**
   * Sets the sort direction to descending.
   */
  setSortDescending,
};
