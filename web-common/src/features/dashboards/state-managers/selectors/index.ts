import { sortingSelectors } from "./sorting";
import { derived, type Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import type { ReadablesObj, SelectorFnsObj } from "./types";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type {
  RpcStatus,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { activeMeasure } from "./core-selectors";
import { formattingSelectors } from "./data-formatting";
import { contextColSelectors } from "./context-column";

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

    // Note: for now, some core selectors are kept in the root of the
    // selectors object. Can revisit that later if we want to move them.

    /**
     * The active measure for the dashboard.
     */
    activeMeasure: derived(
      [dashboardStore, metricsSpecQueryResultStore],
      activeMeasure
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
      derived([dashboardStore, metricsSpecQueryResultStore], selectorFn),
    ])
  ) as ReadablesObj<T>;
}
