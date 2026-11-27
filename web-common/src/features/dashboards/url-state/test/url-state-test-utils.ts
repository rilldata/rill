import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { convertURLSearchParamsToFullExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { normalizeExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/merge-explore-url-params";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import {
  createInEffectRoot,
  createTestMetricsViewsProvider,
  type MetricsViewSpecs,
  useMetricsViewMocks,
} from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { ALL_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { deepClone } from "@vitest/utils/helpers";
import { get } from "svelte/store";
import { afterAll, afterEach, beforeAll, beforeEach } from "vitest";

/**
 * Serves `specs` from ListResources and hands each test its own ExpressionFilterManager over them.
 *
 * The specs are read-only, so a single MetricsViewsProvider serves the whole file. The filter
 * manager holds the filter state though, so it is rebuilt for every test.
 */
export function useTestFilterManager(specs: MetricsViewSpecs) {
  useMetricsViewMocks(specs);

  let metricsViewsProvider: MetricsViewsProvider;
  let destroyProvider: () => void;
  let filterManager: ExpressionFilterManager;
  let destroyFilterManager: () => void;

  beforeAll(async () => {
    const provider = await createTestMetricsViewsProvider(Object.keys(specs));
    metricsViewsProvider = provider.value;
    destroyProvider = provider.destroy;
  });

  beforeEach(() => {
    const created = createInEffectRoot(
      () =>
        new ExpressionFilterManager(
          metricsViewsProvider,
          new YAMLConfigProvider(),
          true,
        ),
    );
    filterManager = created.value;
    destroyFilterManager = created.destroy;
  });

  afterEach(() => destroyFilterManager());

  afterAll(() => destroyProvider());

  return () => filterManager;
}

/**
 * Loads the url into the explore store, as the dashboard loader does on a visit without a session.
 * See {@link convertURLSearchParamsToFullExploreStateForTest}.
 */
export function applyURLToExploreState(
  url: URL,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  filterManager: ExpressionFilterManager,
) {
  filterManager.storeSync.setUrlParams(
    normalizeExploreUrlParams(url.searchParams),
  );

  const { exploreState, errors } =
    convertURLSearchParamsToFullExploreStateForTest(
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );
  metricsExplorerStore.mergePartialExplorerEntity(
    AD_BIDS_EXPLORE_NAME,
    // The full state leaves out an undefined time dimension, which is the primary one,
    // so the store would keep the previous one. In the app the time filter manager resets it, but these tests do not run one.
    {
      ...exploreState,
      selectedTimeDimension: exploreState.selectedTimeDimension,
    },
    filterManager,
  );
  return errors;
}

// cleans the metrics explore of any state that is not stored or restored from url state
export function getCleanMetricsExploreForAssertion() {
  // clone the existing state so that any mutations do affect the copy during assertion
  const cleanedState = deepClone(
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
  ) as Partial<ExploreState>;

  delete cleanedState.name;
  delete cleanedState.proto;
  delete cleanedState.dimensionFilterExcludeMode;
  delete cleanedState.temporaryFilterName;
  delete cleanedState.contextColumnWidths;
  if (cleanedState.selectedTimeRange) {
    cleanedState.selectedTimeRange = {
      name: cleanedState.selectedTimeRange?.name ?? ALL_TIME_RANGE_ALIAS,
      interval: cleanedState.selectedTimeRange?.interval,
    } as DashboardTimeControls;
  }
  delete cleanedState.lastDefinedScrubRange;

  // TODO
  delete cleanedState.leaderboardContextColumn;

  return cleanedState;
}
