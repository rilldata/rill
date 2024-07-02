import type { DashboardMutables } from "./types";

export const setComparisonDimension = (
  { dashboard }: DashboardMutables,
  dimensionName: string | undefined,
) => {
  dashboard.selectedComparisonDimension = dimensionName;
};

export const comparisonActions = {
  /**
   * Sets the comparison dimension for the dashboard.
   */
  setComparisonDimension,
};
