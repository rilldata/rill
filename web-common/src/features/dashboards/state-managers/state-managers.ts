import {
  type ContextColWidths,
  type MetricsExplorerEntity,
  contextColWidthDefaults,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import {
  type ExploreValidSpecResponse,
  useExploreValidSpec,
} from "@rilldata/web-common/features/explores/selectors";
import {
  type RpcStatus,
  type V1ExplorePreset,
  type V1MetricsViewTimeRangeResponse,
  createQueryServiceMetricsViewTimeRange,
} from "@rilldata/web-common/runtime-client";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient, QueryObserverResult } from "@tanstack/svelte-query";
import { getContext } from "svelte";
import {
  type Readable,
  type Writable,
  derived,
  get,
  writable,
} from "svelte/store";
import {
  type MetricsExplorerStoreType,
  metricsExplorerStore,
  updateMetricsExplorerByName,
  useExploreState,
} from "web-common/src/features/dashboards/stores/dashboard-stores";
import { type StateManagerActions, createStateManagerActions } from "./actions";
import type { DashboardCallbackExecutor } from "./actions/types";
import {
  type StateManagerReadables,
  createStateManagerReadables,
} from "./selectors";

export type StateManagers = {
  runtime: Writable<Runtime>;
  metricsViewName: Writable<string>;
  exploreName: Writable<string>;
  metricsStore: Readable<MetricsExplorerStoreType>;
  dashboardStore: Readable<MetricsExplorerEntity>;
  timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  >;
  validSpecStore: Readable<
    QueryObserverResult<ExploreValidSpecResponse, RpcStatus>
  >;
  queryClient: QueryClient;
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
  defaultExploreState: Readable<V1ExplorePreset>;
};

export const DEFAULT_STORE_KEY = Symbol("state-managers");

export function getStateManagers(): StateManagers {
  return getContext(DEFAULT_STORE_KEY);
}

export function createStateManagers({
  queryClient,
  metricsViewName,
  exploreName,
}: {
  queryClient: QueryClient;
  metricsViewName: string;
  exploreName: string;
}): StateManagers {
  const metricsViewNameStore = writable(metricsViewName);
  const exploreNameStore = writable(exploreName);

  const dashboardStore: Readable<MetricsExplorerEntity> = derived(
    [exploreNameStore],
    ([name], set) => {
      const exploreState = useExploreState(name);
      return exploreState.subscribe(set);
    },
  );

  const validSpecStore: Readable<
    QueryObserverResult<ExploreValidSpecResponse, RpcStatus>
  > = derived([runtime, exploreNameStore], ([r, exploreName], set) =>
    useExploreValidSpec(r.instanceId, exploreName, { queryClient }).subscribe(
      set,
    ),
  );

  const timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  > = derived(
    [runtime, metricsViewNameStore, validSpecStore],
    ([runtime, mvName, validSpec], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtime.instanceId,
        mvName,
        {},
        {
          query: {
            queryClient,
            enabled: !!validSpec?.data?.metricsView?.timeDimension,
            staleTime: Infinity,
            cacheTime: Infinity,
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

  const defaultExploreState = derived(
    [validSpecStore, timeRangeSummaryStore],
    ([validSpec, timeRangeSummary]) => {
      if (!validSpec.data?.explore) {
        return {};
      }
      return getDefaultExplorePreset(
        validSpec.data?.explore ?? {},
        validSpec.data.metricsView ?? {},
        timeRangeSummary.data,
      );
    },
  );

  return {
    runtime: runtime,
    metricsViewName: metricsViewNameStore,
    exploreName: exploreNameStore,
    metricsStore: metricsExplorerStore,
    timeRangeSummaryStore,
    validSpecStore,
    queryClient,
    dashboardStore,

    updateDashboard,
    /**
     * A collection of Readables that can be used to select data from the dashboard.
     */
    selectors: createStateManagerReadables({
      dashboardStore,
      validSpecStore,
      timeRangeSummaryStore,
      queryClient,
    }),
    /**
     * A collection of functions that update the dashboard data model.
     */
    actions: createStateManagerActions({
      updateDashboard,
    }),
    contextColumnWidths,
    defaultExploreState,
  };
}
