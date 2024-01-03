import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { sortTypeForContextColumnType } from "../../stores/dashboard-stores";
import type { DashboardMutables } from "./types";

export const setContextColumn = (
  { dashboard }: DashboardMutables,

  contextColumn: LeaderboardContextColumn,
) => {
  const initialSort = sortTypeForContextColumnType(
    dashboard.leaderboardContextColumn,
  );
  switch (contextColumn) {
    case LeaderboardContextColumn.DELTA_ABSOLUTE:
    case LeaderboardContextColumn.DELTA_PERCENT: {
      // if there is no time comparison, then we can't show
      // these context columns, so return with no change
      if (dashboard.showTimeComparison === false) return;

      dashboard.leaderboardContextColumn = contextColumn;
      break;
    }
    default:
      dashboard.leaderboardContextColumn = contextColumn;
  }

  // if we have changed the context column, and the leaderboard is
  // sorted by the context column from before we made the change,
  // then we also need to change
  // the sort type to match the new context column
  if (dashboard.dashboardSortType === initialSort) {
    dashboard.dashboardSortType = sortTypeForContextColumnType(contextColumn);
  }
};

export const contextColActions = {
  /**
   * Updates the dashboard to use the context column of the given type,
   * as well as updating to sort by that context column.
   */
  setContextColumn,
};
