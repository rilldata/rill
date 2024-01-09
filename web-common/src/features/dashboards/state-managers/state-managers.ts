import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { writable, Writable, Readable, derived, get } from "svelte/store";
import { getContext } from "svelte";
import type { QueryClient, QueryObserverResult } from "@tanstack/svelte-query";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  MetricsExplorerStoreType,
  metricsExplorerStore,
  updateMetricsExplorerByName,
  useDashboardStore,
} from "web-common/src/features/dashboards/stores/dashboard-stores";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  StateManagerReadables,
  createStateManagerReadables,
} from "./selectors";
import { createStateManagerActions, type StateManagerActions } from "./actions";
import type { DashboardCallbackExecutor } from "./actions/types";
import {
  ResourceKind,
  useResource,
} from "../../entity-management/resource-selectors";
import {
  createQueryServiceColumnTimeRange,
  V1ColumnTimeRangeResponse,
  type RpcStatus,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export type StateManagers = {
  runtime: Writable<Runtime>;
  metricsViewName: Writable<string>;
  metricsStore: Readable<MetricsExplorerStoreType>;
  dashboardStore: Readable<MetricsExplorerEntity>;
  timeRangeSummaryStore: Readable<
    QueryObserverResult<V1ColumnTimeRangeResponse, unknown>
  >;
  queryClient: QueryClient;
  setMetricsViewName: (s: string) => void;
  updateDashboard: DashboardCallbackExecutor;
  /**
   * A collection of Readables that can be used to select data from the dashboard.
   */
  selectors: StateManagerReadables;
  /**
   * A collection of functions that update the dashboard data model.
   */
  actions: StateManagerActions;
};

export const DEFAULT_STORE_KEY = Symbol("state-managers");

export function getStateManagers(): StateManagers {
  return getContext(DEFAULT_STORE_KEY);
}

export function createStateManagers({
  queryClient,
  metricsViewName,
}: {
  queryClient: QueryClient;
  metricsViewName: string;
}): StateManagers {
  const metricsViewNameStore = writable(metricsViewName);
  const dashboardStore: Readable<MetricsExplorerEntity> = derived(
    [metricsViewNameStore],
    ([name], set) => {
      const store = useDashboardStore(name);
      return store.subscribe(set);
    },
  );

  // Note: this is equivalent to `useMetaQuery`
  const metricsSpecStore: Readable<
    QueryObserverResult<V1MetricsViewSpec, RpcStatus>
  > = derived([runtime, metricsViewNameStore], ([r, metricViewName], set) => {
    useResource(
      r.instanceId,
      metricViewName,
      ResourceKind.MetricsView,
      (data) => data.metricsView?.state?.validSpec,
      queryClient,
    ).subscribe(set);
  });

  const timeRangeSummaryStore: Readable<
    QueryObserverResult<V1ColumnTimeRangeResponse, unknown>
  > = derived([runtime, metricsSpecStore], ([runtime, metricsView], set) =>
    createQueryServiceColumnTimeRange(
      runtime.instanceId,
      metricsView.data?.table ?? "",
      {
        columnName: metricsView.data?.timeDimension,
      },
      {
        query: {
          enabled: !!metricsView.data?.timeDimension,
          queryClient: queryClient,
        },
      },
    ).subscribe(set),
  );

  const updateDashboard = (
    callback: (metricsExplorer: MetricsExplorerEntity) => void,
  ) => {
    const name = get(dashboardStore).name;
    // TODO: Remove dependency on MetricsExplorerStore singleton and its exports
    updateMetricsExplorerByName(name, callback);
  };

  return {
    runtime: runtime,
    metricsViewName: metricsViewNameStore,
    metricsStore: metricsExplorerStore,
    timeRangeSummaryStore,
    queryClient,
    dashboardStore,
    setMetricsViewName: (name) => {
      metricsViewNameStore.set(name);
    },
    updateDashboard,
    /**
     * A collection of Readables that can be used to select data from the dashboard.
     */
    selectors: createStateManagerReadables({
      dashboardStore,
      metricsSpecQueryResultStore: metricsSpecStore,
      timeRangeSummaryStore,
    }),
    /**
     * A collection of functions that update the dashboard data model.
     */
    actions: createStateManagerActions({
      updateDashboard,
      cancelQueries: () => {
        queryClient.cancelQueries();
      },
    }),
  };
}
