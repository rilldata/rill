import type { DashboardDataSources } from "./types";

/**
 * Returns a function that can be used to get the sorted query body
 * for a leaderboard for the given dimension.
 *
 * If there is no active measure, or if the active measure
 */
export function isBeingCompared(
  dashData: DashboardDataSources,
): (dimensionName: string) => boolean {
  return (dimensionName: string) =>
    dashData.dashboard?.selectedComparisonDimension === dimensionName;
}

export const comparisonSelectors = {
  /**
   * Readable containg a function that will return
   * true if the dimension given is being compared.
   */
  isBeingCompared,
};
