/**
 * This enum determines the state of the context column in the leaderboard
 */
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
