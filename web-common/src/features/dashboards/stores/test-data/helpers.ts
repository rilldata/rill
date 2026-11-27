import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { renderInRuntimeContext } from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_DEFAULT_TIME_RANGE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_MIRROR_NAME,
  AD_BIDS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getExploreUrlParamsFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { convertURLSearchParamsToFullExploreState } from "@rilldata/web-common/features/dashboards/url-state/convert-url-params-to-full-explore-state";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params";
import {
  mergeExploreUrlParams,
  normalizeExploreUrlParams,
} from "@rilldata/web-common/features/dashboards/url-state/merge-explore-url-params";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { deepClone } from "@vitest/utils/helpers";
import { get } from "svelte/store";
import { expect } from "vitest";

export function resetDashboardStore() {
  metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  metricsExplorerStore.remove(AD_BIDS_MIRROR_NAME);
  initAdBidsInStore();
  initAdBidsMirrorInStore();
}

export function initAdBidsInStore() {
  metricsExplorerStore.init(
    AD_BIDS_EXPLORE_NAME,
    getInitExploreStateForTest(AD_BIDS_METRICS_INIT, AD_BIDS_EXPLORE_INIT, {
      timeRangeSummary: {
        min: "2022-01-01",
        max: "2022-03-31",
      },
    }),
  );
}

export function initAdBidsMirrorInStore() {
  metricsExplorerStore.init(
    AD_BIDS_MIRROR_NAME,
    getInitExploreStateForTest(AD_BIDS_METRICS_INIT, AD_BIDS_EXPLORE_INIT, {
      timeRangeSummary: {
        min: "2022-01-01",
        max: "2022-03-31",
      },
    }),
  );
}

export function getInitExploreStateForTest(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1MetricsViewTimeRangeResponse | undefined = undefined,
) {
  return convertURLSearchParamsToFullExploreStateForTest(
    new URLSearchParams(),
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary?.timeRangeSummary,
  ).exploreState;
}

/**
 * Converts url params to a full explore state the way the dashboard loader does on a visit without a session:
 * a url without params opens with the yaml and rill defaults,
 * while params missing from a url with params are filled from the rill defaults only.
 */
export function convertURLSearchParamsToFullExploreStateForTest(
  urlParams: URLSearchParams,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const mergedUrlParams = mergeExploreUrlParams({
    urlParams: normalizeExploreUrlParams(urlParams),
    sessionParams: undefined,
    mostRecentParams: undefined,
    bookmarkParams: undefined,
    yamlParams: getExploreUrlParamsFromYAMLConfig(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
    ),
    rillDefaultParams: getRillDefaultExploreUrlParams(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
    ),
    backButtonUsed: false,
  });
  return convertURLSearchParamsToFullExploreState(
    mergedUrlParams,
    metricsViewSpec,
    exploreSpec,
    getRillDefaultExploreState(metricsViewSpec, exploreSpec, timeRangeSummary),
  );
}

// Wrapper function to simplify assert call
export function assertMetricsView(
  name: string,
  filters: V1Expression = createAndExpression([]),
  timeRange = {
    name: AD_BIDS_DEFAULT_TIME_RANGE.name,
    interval: AD_BIDS_DEFAULT_TIME_RANGE.interval,
  } as DashboardTimeControls,
  selectedMeasure = AD_BIDS_IMPRESSIONS_MEASURE,
) {
  assertMetricsViewRaw(name, filters, timeRange, selectedMeasure);
}

// Raw assert function without any optional params.
// This allows us to assert for `undefined`
// TODO: find a better solution that this hack
export function assertMetricsViewRaw(
  name: string,
  filters: V1Expression,
  timeRange: DashboardTimeControls | undefined,
  selectedMeasure: string,
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  expect(metricsView.whereFilter).toEqual(filters);
  expect(metricsView.selectedTimeRange).toEqual(timeRange);
  expect(metricsView.leaderboardSortByMeasureName).toEqual(selectedMeasure);
}

/**
 * createStateManagers builds providers that call `createQuery`, so it can only run during
 * component initialisation. renderInRuntimeContext mounts a harness that supplies the
 * QueryClient context and an effect owner. Callers own the returned `destroy`.
 */
export function initStateManagers(metricsViewName = AD_BIDS_NAME) {
  const exploreName = metricsViewName + "_explore";

  metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  initAdBidsInStore();

  const {
    value: stateManagers,
    queryClient,
    destroy,
  } = renderInRuntimeContext(({ queryClient, runtimeClient }) =>
    createStateManagers({
      queryClient,
      metricsViewName,
      exploreName,
      runtimeClient,
    }),
  );

  return { stateManagers, queryClient, destroy };
}

export function getPartialDashboard(
  name: string,
  keys: (keyof ExploreState)[],
) {
  const dashboard = get(metricsExplorerStore).entities[name];
  const partialDashboard = {} as ExploreState;
  keys.forEach(
    (key) => ((partialDashboard as any)[key] = deepClone(dashboard[key])),
  );
  return partialDashboard;
}
