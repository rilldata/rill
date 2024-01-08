import { setDisplayComparison } from "../../stores/dashboard-stores";
import type { DashboardMutables } from "./types";

export const setComparisonDimension = (
  { dashboard }: DashboardMutables,
  dimensionName: string | undefined,
) => {
  if (dimensionName === undefined) {
    setDisplayComparison(dashboard, true);
  } else {
    setDisplayComparison(dashboard, false);
  }
  dashboard.selectedComparisonDimension = dimensionName;
};

export const comparisonActions = {
  /**
   * Sets the comparison dimension for the dashboard.
   */
  setComparisonDimension,
};
