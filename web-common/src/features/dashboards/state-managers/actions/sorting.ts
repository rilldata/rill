import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { DashboardMutables } from "./types";

/**
 * Toggles the sort direction between ascending and descending
 */
function toggleSortDirection(dashboard: DashboardMutables["dashboard"]) {
  dashboard.sortDirection =
    dashboard.sortDirection === SortDirection.ASCENDING
      ? SortDirection.DESCENDING
      : SortDirection.ASCENDING;
}

/**
 * Updates the sort type and direction for a measure
 */
function updateMeasureSort(
  dashboard: DashboardMutables["dashboard"],
  sortType: SortType,
  measureName?: string,
) {
  dashboard.dashboardSortType = sortType;
  dashboard.sortDirection = SortDirection.DESCENDING;
  if (measureName) {
    dashboard.sortedMeasureName = measureName;
  }
}

/**
 * Toggles the sort state for a measure or dimension
 *
 * If the sortType matches the current sort type, it toggles the sort direction.
 * Otherwise, it updates the sort type and sets the sort direction to descending.
 *
 * @param dashboard - The dashboard state to update
 * @param sortType - The type of sort to apply (value, delta, percent, etc.)
 * @param measureName - Optional measure name to sort by
 */
export const toggleSort = (
  { dashboard }: DashboardMutables,
  sortType: SortType,
  measureName?: string,
) => {
  if (sortType === undefined || dashboard.dashboardSortType === sortType) {
    toggleSortDirection(dashboard);
  } else {
    updateMeasureSort(dashboard, sortType, measureName);
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
