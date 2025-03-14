import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import type { DashboardDataSources } from "./types";

const contextColumnWidth = ({ dashboard }: DashboardDataSources): string => {
  const contextType = dashboard.leaderboardContextColumn;
  const width = dashboard.contextColumnWidths[contextType];
  if (typeof width === "number") {
    return width + "px";
  }
  return "0px";
};

export const contextColSelectors = {
  /**
   * Gets the active context column type for the dashboard.
   */
  contextColumn: ({ dashboard }: DashboardDataSources) =>
    dashboard.leaderboardContextColumn,

  /**
   * Is the context column hidden in the leaderboards?
   */
  isHidden: ({ dashboard }: DashboardDataSources) =>
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.HIDDEN,

  /**
   * Is the Percentage change context column currently active in the leaderboards?
   */
  isDeltaPercent: ({ dashboard }: DashboardDataSources) =>
    dashboard.leaderboardContextColumn ===
    LeaderboardContextColumn.DELTA_PERCENT,

  /**
   * Is the absolute change context column currently active in the leaderboards?
   */
  isDeltaAbsolute: ({ dashboard }: DashboardDataSources) =>
    dashboard.leaderboardContextColumn ===
    LeaderboardContextColumn.DELTA_ABSOLUTE,

  /**
   * Is the percent-of-total context column currently active in the leaderboards?
   */
  isPercentOfTotal: ({ dashboard }: DashboardDataSources) =>
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.PERCENT,

  /**
   * `true` if the context column is either percent or delta percent,
   * `false` otherwise.
   */
  isAPercentColumn: ({ dashboard }: DashboardDataSources) =>
    dashboard.leaderboardContextColumn ===
      LeaderboardContextColumn.DELTA_PERCENT ||
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.PERCENT,

  /**
   * returns a css style string specifying the width of the context
   * column in the leaderboards.
   */
  widthPx: contextColumnWidth,
};
