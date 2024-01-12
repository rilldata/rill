import type {
  RpcStatus,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import type { Expand } from "../types";

/**
 * DashboardDataSources collects all the information about a dashboard
 * that is needed to select data from it, including the local dashboard
 * state and query reseults. This is the *instantaneous* state of the
 * dashboard, after extracting the data from query results and other readables.
 *
 * Since this is a snapshot of the dashboard state, it is easier to build
 * selectors that are pure functions of that instaneous dashboard state.
 *
 * These simple functions can be composed into more complex selectors
 * outside of component code,
 * and utimately wrapped in Readables for use in components.
 */
export type DashboardDataSources = {
  dashboard: MetricsExplorerEntity;
  metricsSpecQueryResult: QueryObserverResult<V1MetricsViewSpec, RpcStatus>;
  timeRangeSummary: QueryObserverResult<
    V1MetricsViewTimeRangeResponse,
    unknown
  >;
};

/**
 * A SelectorFn is a pure function that takes dashboard data
 * (a DashboardDataSources object) and returns some derived value from it.
 */
export type SelectorFn<T> = (args: DashboardDataSources) => T;

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
