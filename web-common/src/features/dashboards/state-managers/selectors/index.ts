import { sortingSelectors } from "./sorting";
import { derived, type Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

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

/**
 * A SelectorFn is a pure function that takes dashboard data
 * (a MetricsExplorerEntity) and returns some derived value from it.
 */
type SelectorFn<T> = (dashboard: MetricsExplorerEntity) => T;

/**
 * A SelectorFnsObj object is a collection of pure SelectorFn functions.
 */
type SelectorFnsObj = {
  [key: string]: SelectorFn<unknown>;
};

/**
 * A ReadablesObj object is a collection readables that are connected
 * to the live dashboard store and can be
 * used to select data from the dashboard.
 */
type ReadablesObj<T extends SelectorFnsObj> = Expand<{
  [P in keyof T]: Readable<ReturnType<T[P]>>;
}>;

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
