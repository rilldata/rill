import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { sortTypeForContextColumnType } from "../../stores/dashboard-stores";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export const setContextColumn = (
  metricsExplorer: MetricsExplorerEntity,
  contextColumn: LeaderboardContextColumn
) => {
  const initialSort = sortTypeForContextColumnType(
    metricsExplorer.leaderboardContextColumn
  );
  switch (contextColumn) {
    case LeaderboardContextColumn.DELTA_ABSOLUTE:
    case LeaderboardContextColumn.DELTA_PERCENT: {
      // if there is no time comparison, then we can't show
      // these context columns, so return with no change
      if (metricsExplorer.showTimeComparison === false) return;

      metricsExplorer.leaderboardContextColumn = contextColumn;
      break;
    }
    default:
      metricsExplorer.leaderboardContextColumn = contextColumn;
  }

  // if we have changed the context column, and the leaderboard is
  // sorted by the context column from before we made the change,
  // then we also need to change
  // the sort type to match the new context column
  if (metricsExplorer.dashboardSortType === initialSort) {
    metricsExplorer.dashboardSortType =
      sortTypeForContextColumnType(contextColumn);
  }
};

export const contextColActions = {
  /**
   * Updates the dashboard to use the context column of the given type,
   * as well as updating to sort by that context column.
   */
  setContextColumn,
};
