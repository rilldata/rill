import type { Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

/**
 * A SelectorFn is a pure function that takes dashboard data
 * (a MetricsExplorerEntity) and returns some derived value from it.
 */
export type SelectorFn<T> = (dashboard: MetricsExplorerEntity) => T;

/**
 * A SelectorFnsObj object is a collection of pure SelectorFn functions.
 */
export type SelectorFnsObj = {
  [key: string]: SelectorFn<unknown>;
};

/**
 * A ReadablesObj object is a collection readables that are connected
 * to the live dashboard store and can be
 * used to select data from the dashboard.
 */
export type ReadablesObj<T extends SelectorFnsObj> = Expand<{
  [P in keyof T]: Readable<ReturnType<T[P]>>;
}>;
