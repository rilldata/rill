import { sortingSelectors } from "./sorting";
import { derived, type Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import type { ReadablesObj, SelectorFnsObj } from "./types";

export type StateManagerReadables = ReturnType<
  typeof createStateManagerReadables
>;

export const createStateManagerReadables = (
  dashboardStore: Readable<MetricsExplorerEntity>
) => {
  return {
    /**
     * Readables related to the sorting state of the dashboard.
     */
    sorting: createReadablesFromSelectors(sortingSelectors, dashboardStore),
  };
};

function createReadablesFromSelectors<T extends SelectorFnsObj>(
  selectors: T,
  dashboardStore: Readable<MetricsExplorerEntity>
): ReadablesObj<T> {
  return Object.fromEntries(
    Object.entries(selectors).map(([key, selectorFn]) => [
      key,
      derived(dashboardStore, selectorFn),
    ])
  ) as ReadablesObj<T>;
}
