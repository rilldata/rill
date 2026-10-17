/**
 * The surfaces of an explore dashboard that carry precision chips. Each holds
 * its own exact selection (see `ExactSelection`), so a click on a chip only
 * re-runs the queries of the surface it is in.
 */
export const exploreExactSurface = {
  /** The big numbers and their time series charts. */
  measures: "measures",
  dimensionTable: "dimension-table",
  /** The Time Dimension Detail table. */
  tdd: "tdd",
  pivot: "pivot",
  leaderboard: (dimensionName: string) => `leaderboard:${dimensionName}`,
} as const;
