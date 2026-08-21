import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_VIEW,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
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
  V1ExplorePreset,
  V1ExploreSpec,
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
        ),
    );
    filterManager = created.value;
    destroyFilterManager = created.destroy;
  });

  afterEach(() => destroyFilterManager());

  afterAll(() => destroyProvider());

  return () => filterManager;
}

export function applyURLToExploreState(
  url: URL,
  exploreSpec: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
  filterManager: ExpressionFilterManager,
) {
  filterManager.setUrlParams(url.searchParams);

  const { partialExploreState: partialExploreStateDefaultUrl, errors } =
    convertURLSearchParamsToExploreState(
      url.searchParams,
      AD_BIDS_METRICS_VIEW,
      exploreSpec,
      defaultExplorePreset,
    );
  metricsExplorerStore.mergePartialExplorerEntity(
    AD_BIDS_EXPLORE_NAME,
    partialExploreStateDefaultUrl,
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
