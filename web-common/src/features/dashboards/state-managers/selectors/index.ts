import type { Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../dashboard-stores";
import { createSortingSelectors, type SortingSelectors } from "./sorting";

export type StateManagerSelectors = {
  sorting: SortingSelectors;
};

export const createStateManagerSelectors = (
  dashboardStore: Readable<MetricsExplorerEntity>
): StateManagerSelectors => {
  return {
    /**
     * Selectors related to the sort state of the dashboard.
     */
    sorting: createSortingSelectors(dashboardStore),
  };
};
