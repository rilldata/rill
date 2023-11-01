import type { DashboardMutatorFnGeneralArgs } from "./types";

export const setPrimaryDimension = (
  { dashboard }: DashboardMutatorFnGeneralArgs,

  dimensionName: string | undefined
) => {
  dashboard.selectedDimensionName = dimensionName;
};

export const dimensionActions = {
  /**
   * Sets the primary dimension for the dashboard, which
   * activates the dimension table. Setting the primary dimension
   * to undefined closes the dimension table.
   */
  setPrimaryDimension,
};
