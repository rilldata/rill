import { sortingSelectors } from "./sorting";
import { derived, type Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import type { ReadablesObj, SelectorFnsObj } from "./types";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type {
  RpcStatus,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { formattingSelectors } from "./data-formatting";
import { contextColSelectors } from "./context-column";
import { activeMeasureSelectors } from "./active-measure";

export type StateManagerReadables = ReturnType<
  typeof createStateManagerReadables
>;

export const createStateManagerReadables = (
  dashboardStore: Readable<MetricsExplorerEntity>,
  metricsSpecQueryResultStore: Readable<
    QueryObserverResult<V1MetricsViewSpec, RpcStatus>
  >
) => {
  return {
    /**
     * Readables related to the sorting state of the dashboard.
     */
    sorting: createReadablesFromSelectors(
      sortingSelectors,
      dashboardStore,
      metricsSpecQueryResultStore
    ),

    /**
     * Readables related to number formatting for the dashboard.
     */
    numberFormat: createReadablesFromSelectors(
      formattingSelectors,
      dashboardStore,
      metricsSpecQueryResultStore
    ),

    /**
     * Readables related to the dashboard context column.
     */
    contextColumn: createReadablesFromSelectors(
      contextColSelectors,
      dashboardStore,
      metricsSpecQueryResultStore
    ),

    /**
     * Readables related to the primary active measure in the
     * leaderboard.
     */
    activeMeasure: createReadablesFromSelectors(
      activeMeasureSelectors,
      dashboardStore,
      metricsSpecQueryResultStore
    ),
  };
};

function createReadablesFromSelectors<T extends SelectorFnsObj>(
  selectors: T,
  dashboardStore: Readable<MetricsExplorerEntity>,
  metricsSpecQueryResultStore: Readable<
    QueryObserverResult<V1MetricsViewSpec, RpcStatus>
  >
): ReadablesObj<T> {
  return Object.fromEntries(
    Object.entries(selectors).map(([key, selectorFn]) => [
      key,
      derived(
        // Note: creating a svelte derived store from multiple stores
        // requires supplying a tuple of stores.
        // To simplify the selector function, we pack this into a single
        // selectorFnArgs object.
        [dashboardStore, metricsSpecQueryResultStore],
        ([dashboard, metricsSpecQueryResult]) =>
          selectorFn({
            dashboard,
            metricsSpecQueryResult,
          })
      ),
    ])
  ) as ReadablesObj<T>;
}
