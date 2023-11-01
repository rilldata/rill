import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export const setPrimaryDimension = (
  dash: MetricsExplorerEntity,
  dimensionName: string | undefined
) => {
  dash.selectedDimensionName = dimensionName;
};

export const dimensionActions = {
  /**
   * Sets the primary dimension for the dashboard, which
   * activates the dimension table. Setting the primary dimension
   * to undefined closes the dimension table.
   */
  setPrimaryDimension,
};
