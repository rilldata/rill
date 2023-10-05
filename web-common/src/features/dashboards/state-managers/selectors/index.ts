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
    sorting: createSortingSelectors(dashboardStore),
  };
};
