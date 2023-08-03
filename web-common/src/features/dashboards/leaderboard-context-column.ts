/**
 * This enum determines the state of the context column in the leaderboard
 */
export enum LeaderboardContextColumn {
  // show percent-of-total
  PERCENT = "percent",
  // show percent change of the value compared to the previous time range
  DELTA_CHANGE = "delta_change",
  // Do not show the context column
  HIDDEN = "hidden",
}
