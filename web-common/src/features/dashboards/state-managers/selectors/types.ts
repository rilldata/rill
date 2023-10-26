import type { Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type {
  RpcStatus,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Arguments to a selector function. By putting these in a tuple,
 * these become compatible with the svelte derived store function arguments.
 */
export type SelectorFnArgs = {
  dashboard: MetricsExplorerEntity;
  metricsSpecQueryResult: QueryObserverResult<V1MetricsViewSpec, RpcStatus>;
};

/**
 * A SelectorFn is a pure function that takes dashboard data
 * (a MetricsExplorerEntity) and returns some derived value from it.
 */
export type SelectorFn<T> = (args: SelectorFnArgs) => T;

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
