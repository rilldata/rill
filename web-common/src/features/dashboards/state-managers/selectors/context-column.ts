import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import type { SelectorFnArgs } from "./types";

const contextColumnWidth = (contextType: LeaderboardContextColumn): string => {
  switch (contextType) {
    case LeaderboardContextColumn.DELTA_ABSOLUTE:
    case LeaderboardContextColumn.DELTA_PERCENT:
      return "56px";
    case LeaderboardContextColumn.PERCENT:
      return "44px";
    case LeaderboardContextColumn.HIDDEN:
      return "0px";
    default:
      throw new Error("Invalid context column, all cases must be handled");
  }
};

export const contextColSelectors = {
  /**
   * Gets the active context column type for the dashboard.
   */
  contextColumn: ({ dashboard }: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn,

  /**
   * Is the context column hidden in the leaderboards?
   */
  isHidden: ({ dashboard }: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.HIDDEN,

  /**
   * Is the Percentage change context column currently active in the leaderboards?
   */
  isDeltaPercent: ({ dashboard }: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn ===
    LeaderboardContextColumn.DELTA_PERCENT,

  /**
   * Is the absolute change context column currently active in the leaderboards?
   */
  isDeltaAbsolute: ({ dashboard }: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn ===
    LeaderboardContextColumn.DELTA_ABSOLUTE,

  /**
   * Is the percent-of-total context column currently active in the leaderboards?
   */
  isPercentOfTotal: ({ dashboard }: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.PERCENT,

  /**
   * `true` if the context column is either percent or delta percent,
   * `false` otherwise.
   */
  isAPercentColumn: ({ dashboard }: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn ===
      LeaderboardContextColumn.DELTA_PERCENT ||
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.PERCENT,

  /**
   * returns a css style string specifying the width of the context
   * column in the leaderboards.
   */
  widthPx: ({ dashboard }: SelectorFnArgs) =>
    contextColumnWidth(dashboard.leaderboardContextColumn),
};
