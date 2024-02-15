import {
  contextColWidthDefaults,
  type ContextColWidths,
  type MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  getPersistentDashboardStore,
  initPersistentDashboardStore,
} from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import {
  V1MetricsViewTimeRangeResponse,
  createQueryServiceMetricsViewTimeRange,
  type RpcStatus,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient, QueryObserverResult } from "@tanstack/svelte-query";
import { getContext } from "svelte";
import { Readable, Writable, derived, get, writable } from "svelte/store";
import {
  MetricsExplorerStoreType,
  metricsExplorerStore,
  updateMetricsExplorerByName,
  useDashboardStore,
} from "web-common/src/features/dashboards/stores/dashboard-stores";
import {
  ResourceKind,
  useResource,
} from "../../entity-management/resource-selectors";
import { createStateManagerActions, type StateManagerActions } from "./actions";
import type { DashboardCallbackExecutor } from "./actions/types";
import {
  StateManagerReadables,
  createStateManagerReadables,
} from "./selectors";

export type StateManagers = {
  runtime: Writable<Runtime>;
  metricsViewName: Writable<string>;
  metricsStore: Readable<MetricsExplorerStoreType>;
  dashboardStore: Readable<MetricsExplorerEntity>;
  timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
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
  /**
   * Store to track the width of the context columns in leaderboards.
   * FIXME: this was implemented as a low-risk fix for in advance of
   * the new branding release 2024-01-31, but should be revisted since
   * it's a one-off solution that introduces another new pattern.
   */
  contextColumnWidths: Writable<ContextColWidths>;
};

export const DEFAULT_STORE_KEY = Symbol("state-managers");

export function getStateManagers(): StateManagers {
  return getContext(DEFAULT_STORE_KEY);
}

export function createStateManagers({
  queryClient,
  metricsViewName,
  extraKeyPrefix,
}: {
  queryClient: QueryClient;
  metricsViewName: string;
  extraKeyPrefix?: string;
}): StateManagers {
  const metricsViewNameStore = writable(metricsViewName);
  const dashboardStore: Readable<MetricsExplorerEntity> = derived(
    [metricsViewNameStore],
    ([name], set) => {
      const store = useDashboardStore(name);
      return store.subscribe(set);
    },
  );

  // Note: this is equivalent to `useMetricsView`
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
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  > = derived(
    [runtime, metricsViewNameStore, metricsSpecStore],
    ([runtime, mvName, metricsView], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtime.instanceId,
        mvName,
        {},
        {
          query: {
            queryClient: queryClient,
            enabled: !!metricsView.data?.timeDimension,
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

  const contextColumnWidths = writable<ContextColWidths>(
    contextColWidthDefaults,
  );

  // TODO: once we move everything from dashboard-stores to here, we can get rid of the global
  initPersistentDashboardStore((extraKeyPrefix || "") + metricsViewName);
  const persistentDashboardStore = getPersistentDashboardStore();

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
      queryClient,
    }),
    /**
     * A collection of functions that update the dashboard data model.
     */
    actions: createStateManagerActions({
      updateDashboard,
      cancelQueries: () => {
        queryClient.cancelQueries();
      },
      persistentDashboardStore,
    }),
    contextColumnWidths,
  };
}
