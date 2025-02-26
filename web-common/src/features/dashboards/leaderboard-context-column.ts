/**
 * This enum determines the state of the context column in the leaderboard
 */
// FIXME: this should technically be LeaderboardContextColumnSortType
export enum LeaderboardContextColumn {
  // show percent-of-total
  PERCENT = "percent",
  // show percent change of the value compared to the previous time range
  DELTA_PERCENT = "delta_change",
  // show absolute change of the value compared to the previous time range
  DELTA_ABSOLUTE = "delta_absolute",
  // Do not show the context column
  HIDDEN = "hidden",
}

export type ContextColWidths = {
  [LeaderboardContextColumn.DELTA_ABSOLUTE]: number;
  [LeaderboardContextColumn.DELTA_PERCENT]: number;
  [LeaderboardContextColumn.PERCENT]: number;
};

export const contextColWidthDefaults: ContextColWidths = {
  [LeaderboardContextColumn.DELTA_ABSOLUTE]: 56,
  [LeaderboardContextColumn.DELTA_PERCENT]: 44,
  [LeaderboardContextColumn.PERCENT]: 44,
};
