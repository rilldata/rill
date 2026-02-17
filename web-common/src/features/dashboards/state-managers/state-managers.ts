import { page } from "$app/stores";
import { type ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { cleanUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params";
import { createViewingDefaultsStore } from "@rilldata/web-common/features/dashboards/url-state/viewing-defaults-store";
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
import {
  contextColWidthDefaults,
  type ContextColWidths,
} from "../leaderboard-context-column";

export type StateManagers = {
  runtime: Writable<Runtime>;
  metricsViewName: Writable<string>;
  exploreName: Writable<string>;
  metricsStore: Readable<MetricsExplorerStoreType>;
  dashboardStore: Readable<ExploreState>;
  timeDimension: Writable<string | undefined>;
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
  /**
   * Whether the current explore state matches the YAML-configured defaults.
   * Compares cleaned current URL params against YAML default URL params.
   */
  viewingDefaultsStore: Readable<boolean>;
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
  const timeDimension = writable<string | undefined>(undefined);

  const dashboardStore: Readable<ExploreState> = derived(
    [exploreNameStore],
    ([name], set) => {
      const exploreState = useExploreState(name);
      return exploreState.subscribe(set);
    },
  );

  const validSpecStore: Readable<
    QueryObserverResult<ExploreValidSpecResponse, RpcStatus>
  > = derived([runtime, exploreNameStore], ([r, exploreName], set) =>
    useExploreValidSpec(
      r.instanceId,
      exploreName,
      undefined,
      queryClient,
    ).subscribe(set),
  );

  const timeRangeSummaryStore: Readable<
    QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
  > = derived(
    [runtime, metricsViewNameStore, validSpecStore, dashboardStore],
    ([runtime, mvName, validSpec, $dashboardStore], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtime.instanceId,
        mvName,
        {
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

  // YAML default URL params — represents the state when explore is at its YAML-configured defaults.
  const yamlDefaultExploreUrlParams: Readable<URLSearchParams | undefined> =
    derived(
      [validSpecStore, timeRangeSummaryStore],
      ([validSpec, timeRangeSummary]) => {
        const metricsViewSpec = validSpec.data?.metricsView;
        const exploreSpec = validSpec.data?.explore;

        if (
          !metricsViewSpec ||
          !exploreSpec ||
          (metricsViewSpec.timeDimension &&
            !timeRangeSummary.data?.timeRangeSummary)
        ) {
          return undefined;
        }

        const yamlExploreState = getExploreStateFromYAMLConfig(
          exploreSpec,
          timeRangeSummary.data?.timeRangeSummary,
          metricsViewSpec.smallestTimeGrain,
        );

        // Build a minimal TimeControlState directly from the YAML explore state.
        // We avoid getTimeControlState() here because it calls isoDurationToFullTimeRange
        // which can't handle rill-time expressions (e.g. "14D as of latest/D+1D").
        // The YAML state already has the correct name and interval from getGrainForRange.
        const timeControlState: Partial<TimeControlState> = {
          selectedTimeRange: yamlExploreState.selectedTimeRange,
          selectedComparisonTimeRange:
            yamlExploreState.selectedComparisonTimeRange,
        };

        return convertPartialExploreStateToUrlParams(
          exploreSpec,
          metricsViewSpec,
          yamlExploreState,
          timeControlState as TimeControlState,
        );
      },
    );

  // Current URL params cleaned of YAML-default values (so only the "interesting" params remain).
  const currentCleanedUrlParams: Readable<URLSearchParams> = derived(
    [page, yamlDefaultExploreUrlParams],
    ([$page, $yamlDefaults]) => {
      if (!$yamlDefaults) return $page.url.searchParams;
      return cleanUrlParams($page.url.searchParams, $yamlDefaults);
    },
  );

  // Rill opinionated default URL params — the baseline that DashboardStateSync cleans against.
  // Needed to identify which YAML defaults are "significant" (differ from rill defaults).
  const rillDefaultExploreUrlParams: Readable<URLSearchParams | undefined> =
    derived(
      [validSpecStore, timeRangeSummaryStore],
      ([validSpec, timeRangeSummary]) => {
        const metricsViewSpec = validSpec.data?.metricsView;
        const exploreSpec = validSpec.data?.explore;

        if (
          !metricsViewSpec ||
          !exploreSpec ||
          (metricsViewSpec.timeDimension &&
            !timeRangeSummary.data?.timeRangeSummary)
        ) {
          return undefined;
        }

        return getRillDefaultExploreUrlParams(
          metricsViewSpec,
          exploreSpec,
          timeRangeSummary.data?.timeRangeSummary,
        );
      },
    );

  const rawUrlParams: Readable<URLSearchParams> = derived(
    page,
    ($page) => $page.url.searchParams,
  );

  // Viewing defaults when:
  // 1. Forward: cleaned params are empty (no non-YAML-default params in browser URL)
  // 2. Reverse: all YAML defaults that differ from rill defaults are present in the browser URL
  const viewingDefaultsStore = createViewingDefaultsStore(
    currentCleanedUrlParams,
    yamlDefaultExploreUrlParams,
    rillDefaultExploreUrlParams,
    rawUrlParams,
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
    viewingDefaultsStore,
  };
}
