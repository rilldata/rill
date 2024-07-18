import type { DashboardMutables } from "./types";

export const setComparisonDimension = (
  { dashboard }: DashboardMutables,
  dimensionName: string | undefined,
) => {
  // Temporary until we make these not mutually exclusive
  dashboard.showTimeComparison = false;
  dashboard.selectedComparisonDimension = dimensionName;
};

export const comparisonActions = {
  /**
   * Sets the comparison dimension for the dashboard.
   */
  setComparisonDimension,
};
