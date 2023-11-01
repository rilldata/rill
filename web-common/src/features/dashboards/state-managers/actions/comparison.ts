import { setDisplayComparison } from "../../stores/dashboard-stores";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export const setComparisonDimension = (
  dash: MetricsExplorerEntity,
  dimensionName: string | undefined
) => {
  if (dimensionName === undefined) {
    setDisplayComparison(dash, true);
  } else {
    setDisplayComparison(dash, false);
  }
  dash.selectedComparisonDimension = dimensionName;
};

export const comparisonActions = {
  /**
   * Sets the comparison dimension for the dashboard.
   */
  setComparisonDimension,
};
