import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { sortTypeForContextColumnType } from "../../stores/dashboard-stores";
import {
  type ContextColWidths,
  contextColWidthDefaults,
} from "../../stores/metrics-explorer-entity";
import type { DashboardMutables } from "./types";

export const CONTEXT_COL_MAX_WIDTH = 100;

export const setContextColumn = (
  { dashboard }: DashboardMutables,

  contextColumn: LeaderboardContextColumn,
) => {
  const initialSort = sortTypeForContextColumnType(
    dashboard.leaderboardContextColumn,
  );

  // reset context column width to default when changing context column
  resetAllContextColumnWidths(dashboard.contextColumnWidths);

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

export const resetAllContextColumnWidths = (
  contextColumnWidths: ContextColWidths,
) => {
  for (const contextColumn in contextColumnWidths) {
    contextColumnWidths[contextColumn as LeaderboardContextColumn] =
      contextColWidthDefaults[contextColumn as LeaderboardContextColumn];
  }
};

/**
 * Observe this width value, updating the overall width of
 * the context column if the given width is larger than the
 * current width.
 */
export const observeContextColumnWidth = (
  { dashboard }: DashboardMutables,
  contextColumn: LeaderboardContextColumn,
  width: number,
) => {
  dashboard.contextColumnWidths[contextColumn] = Math.min(
    Math.max(width, dashboard.contextColumnWidths[contextColumn]),
    CONTEXT_COL_MAX_WIDTH,
  );
};

export const contextColActions = {
  /**
   * Updates the dashboard to use the context column of the given type,
   * as well as updating to sort by that context column.
   */
  setContextColumn,

  /**
   * Observe this width value, updating the overall width of
   * the context column if the given width is larger than the
   * current width.
   */
  observeContextColumnWidth,
};
