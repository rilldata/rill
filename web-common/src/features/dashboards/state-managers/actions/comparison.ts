import type { DashboardMutables } from "./types";

export const toggleComparisonDimension = (
  { dashboard }: DashboardMutables,
  dimensionName: string | undefined,
) => {
  // Temporary until we make these not mutually exclusive
  dashboard.showTimeComparison = false;
  const isCurrentDimension =
    dashboard.selectedComparisonDimension === dimensionName;
  if (!isCurrentDimension) {
    dashboard.selectedComparisonDimension = dimensionName;
  } else {
    dashboard.selectedComparisonDimension = undefined;
  }
};

export const comparisonActions = {
  /**
   * Sets the comparison dimension for the dashboard.
   */
  toggleComparisonDimension,
};
