import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { DashboardMutables } from "./types";
import { setLeaderboardMeasureName } from "./core-actions";

const isMeasureSortType = (sortType: SortType) =>
  sortType === SortType.VALUE ||
  sortType === SortType.DELTA_ABSOLUTE ||
  sortType === SortType.DELTA_PERCENT ||
  sortType === SortType.PERCENT;

const handleNewMeasureName = (args: DashboardMutables, measureName: string) => {
  const { dashboard } = args;
  setLeaderboardMeasureName(args, measureName);
  dashboard.dashboardSortType = SortType.VALUE;
  dashboard.sortDirection = SortDirection.DESCENDING;
};

const toggleSortDirection = (args: DashboardMutables) => {
  const { dashboard } = args;
  dashboard.sortDirection =
    dashboard.sortDirection === SortDirection.ASCENDING
      ? SortDirection.DESCENDING
      : SortDirection.ASCENDING;
};

export const toggleSort = (
  args: DashboardMutables,
  sortType: SortType,
  measureName?: string,
) => {
  const { dashboard } = args;

  if (
    measureName !== undefined &&
    measureName !== dashboard.leaderboardMeasureName &&
    (sortType === SortType.VALUE ||
      sortType === SortType.DELTA_ABSOLUTE ||
      sortType === SortType.DELTA_PERCENT ||
      sortType === SortType.PERCENT)
  ) {
    setLeaderboardMeasureName(args, measureName);
    dashboard.dashboardSortType = sortType;
    dashboard.sortDirection = SortDirection.DESCENDING;
    return;
  }

  if (sortType === undefined || dashboard.dashboardSortType === sortType) {
    dashboard.sortDirection =
      dashboard.sortDirection === SortDirection.ASCENDING
        ? SortDirection.DESCENDING
        : SortDirection.ASCENDING;
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
};
