import { chartSelectors } from "@rilldata/web-common/features/dashboards/state-managers/selectors/charts";
import { measureFilterSelectors } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import type { ExploreValidSpecResponse } from "@rilldata/web-common/features/explores/selectors";
import type {
  RpcStatus,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient, QueryObserverResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import { activeMeasureSelectors } from "./active-measure";
import { comparisonSelectors } from "./comparisons";
import { contextColSelectors } from "./context-column";
import { formattingSelectors } from "./data-formatting";
import { dimensionFilterSelectors } from "./dimension-filters";
import { dimensionTableSelectors } from "./dimension-table";
import { dimensionSelectors } from "./dimensions";
import { measureSelectors } from "./measures";
import { pivotSelectors } from "./pivot";
import { sortingSelectors } from "./sorting";
import { timeRangeSelectors } from "./time-range";
import type { ReadablesObj, SelectorFnsObj } from "./types";

export type DashboardDataReadables = {
  dashboardStore: Readable<MetricsExplorerEntity>;
  validSpecStore: Readable<
    QueryObserverResult<ExploreValidSpecResponse, RpcStatus>
  >;
  timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  >;
  queryClient: QueryClient;
};

export type StateManagerReadables = ReturnType<
  typeof createStateManagerReadables
>;

export const createStateManagerReadables = (
  dashboardDataReadables: DashboardDataReadables,
) => {
  return {
    /**
     * Readables related to the sorting state of the dashboard.
     */
    sorting: createReadablesFromSelectors(
      sortingSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to number formatting for the dashboard.
     */
    numberFormat: createReadablesFromSelectors(
      formattingSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to the dashboard context column.
     */
    contextColumn: createReadablesFromSelectors(
      contextColSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to the primary active measure in the
     * leaderboard.
     */
    activeMeasure: createReadablesFromSelectors(
      activeMeasureSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to the dimensions available in the
     * leaderboard.
     */
    dimensions: createReadablesFromSelectors(
      dimensionSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to the dimension table.
     *
     * These are valid when the dimension table is visible, and
     * should only be used from within dimension table components.
     */
    dimensionTable: createReadablesFromSelectors(
      dimensionTableSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to selected (aka "filtered")
     * dimension values in the leaderboard, including
     * whether or not a dimension is in include or exclude mode.
     */
    dimensionFilters: createReadablesFromSelectors(
      dimensionFilterSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to measure filters applied to a dimension leaderboard.
     */
    measureFilters: createReadablesFromSelectors(
      measureFilterSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to the state of the time range selector
     * for the dashboard.
     */
    timeRangeSelectors: createReadablesFromSelectors(
      timeRangeSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to the dashboard comparison state
     */
    comparison: createReadablesFromSelectors(
      comparisonSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to dashboard measures
     */
    measures: createReadablesFromSelectors(
      measureSelectors,
      dashboardDataReadables,
    ),

    /**
     * Readables related to pivot state
     */
    pivot: createReadablesFromSelectors(pivotSelectors, dashboardDataReadables),

    /**
     * Readables related to the chart interactions state
     */
    charts: createReadablesFromSelectors(
      chartSelectors,
      dashboardDataReadables,
    ),
  };
};

function createReadablesFromSelectors<T extends SelectorFnsObj>(
  selectors: T,
  readables: DashboardDataReadables,
): ReadablesObj<T> {
  return Object.fromEntries(
    Object.entries(selectors).map(([key, selectorFn]) => [
      key,
      derived(
        // Note: creating a svelte derived store from multiple stores
        // requires supplying a tuple of stores.
        // To simplify the selector function, we pack this into a single
        // selectorFnArgs object.
        [
          readables.dashboardStore,
          readables.validSpecStore,
          readables.timeRangeSummaryStore,
        ],
        ([dashboard, validSpec, timeRangeSummary]) =>
          selectorFn({
            dashboard,
            validMetricsView: validSpec.data?.metricsView,
            validExplore: validSpec.data?.explore,
            timeRangeSummary,
            queryClient: readables.queryClient,
          }),
      ),
    ]),
  ) as ReadablesObj<T>;
}
