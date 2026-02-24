import { type ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { type ExploreValidSpecResponse } from "@rilldata/web-common/features/explores/selectors";
import {
  type V1ExplorePreset,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { createQueryServiceMetricsViewTimeRange } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
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
import {
  contextColWidthDefaults,
  type ContextColWidths,
} from "../leaderboard-context-column";

export type StateManagers = {
  runtime: Writable<Runtime>;
  runtimeClient: RuntimeClient;
  metricsViewName: Writable<string>;
  exploreName: Writable<string>;
  metricsStore: Readable<MetricsExplorerStoreType>;
  dashboardStore: Readable<ExploreState>;
  timeDimension: Writable<string | undefined>;
  timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  >;
  validSpecStore: Readable<
    QueryObserverResult<ExploreValidSpecResponse, Error>
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
  runtimeClient,
}: {
  queryClient: QueryClient;
  metricsViewName: string;
  exploreName: string;
  runtimeClient: RuntimeClient;
}): StateManagers {
  const metricsViewNameStore = writable(metricsViewName);
  const exploreNameStore = writable(exploreName);
  const timeDimension = writable<string | undefined>(undefined);

  const dashboardStore: Readable<ExploreState> = derived(
    [exploreNameStore],
    ([name], set) => {
      const exploreState = useExploreState(name);
      return exploreState.subscribe(set);
    },
  );

  const validSpecStore: Readable<
    QueryObserverResult<ExploreValidSpecResponse, Error>
  > = derived([exploreNameStore], ([exploreName], set) =>
    createRuntimeServiceGetExplore(
      runtimeClient,
      { name: exploreName },
      {
        query: {
          select: (data) =>
            <ExploreValidSpecResponse>{
              explore: data.explore?.explore?.state?.validSpec,
              metricsView: data.metricsView?.metricsView?.state?.validSpec,
            },
          enabled: !!exploreName,
        },
      },
      queryClient,
    ).subscribe(set),
  );

  const timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  > = derived(
    [metricsViewNameStore, validSpecStore, dashboardStore],
    ([mvName, validSpec, $dashboardStore], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtimeClient,
        {
          metricsViewName: mvName,
          timeDimension: $dashboardStore?.selectedTimeDimension,
        },
        {
          query: {
            enabled: !!validSpec?.data?.metricsView?.timeDimension,
            staleTime: Infinity,
            gcTime: Infinity,
          },
        },
        queryClient,
      ).subscribe(set),
  );

  const updateDashboard = (callback: (exploreState: ExploreState) => void) => {
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
        timeRangeSummary.data?.timeRangeSummary,
      );
    },
  );

  return {
    runtime: runtime,
    runtimeClient,
    metricsViewName: metricsViewNameStore,
    exploreName: exploreNameStore,
    metricsStore: metricsExplorerStore,
    timeRangeSummaryStore,
    validSpecStore,
    queryClient,
    dashboardStore,
    timeDimension,
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
