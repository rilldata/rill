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
  contextColumn: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn,
  isHidden: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.HIDDEN,

  isDeltaPercent: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn ===
    LeaderboardContextColumn.DELTA_PERCENT,
  isDeltaAbsolute: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn ===
    LeaderboardContextColumn.DELTA_ABSOLUTE,
  isPercentOfTotal: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.PERCENT,

  /**
   * `true` if the context column is either percent or delta percent,
   * `false` otherwise.
   */
  isAPercentColumn: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.leaderboardContextColumn ===
      LeaderboardContextColumn.DELTA_PERCENT ||
    dashboard.leaderboardContextColumn === LeaderboardContextColumn.PERCENT,

  /**
   * returns a css style string specifying the width of the context
   * column in the leaderboards.
   */
  widthPx: ([dashboard, _]: SelectorFnArgs) =>
    contextColumnWidth(dashboard.leaderboardContextColumn),
};
