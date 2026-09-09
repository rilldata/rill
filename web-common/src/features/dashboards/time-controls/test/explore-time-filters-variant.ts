import ExploreExpressionFiltersTest from "@rilldata/web-common/features/dashboards/filters/test/ExploreExpressionFiltersTest.svelte";
import {
  mockPointerEventsForComponentTesting,
  useDashboardFetchMocksForComponentTests,
  waitForBodyScrollCleanup,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import {
  type HoistedPageForComponentTests,
  PageMockForComponentTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  DEFAULT_TIME_RANGE,
  RESOLVED_RILL_TIMES,
  TIME_RANGE_SUMMARY,
  YAML_TIME_RANGES,
} from "@rilldata/web-common/features/dashboards/time-controls/test/rill-time-mocks";
import { waitForTimeRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/test/time-filter-test-utils";
import type { TimeFiltersVariant } from "@rilldata/web-common/features/dashboards/time-controls/test/time-filters-suite";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, screen, waitFor } from "@testing-library/svelte";
import { get } from "svelte/store";
import { afterAll, beforeEach, expect } from "vitest";

// Url params for the dashboard before any time filter is applied, coming from the yaml preset.
// A rilltime expression carries characters the url escapes, and the url search history holds the
// searches verbatim, so these are built through URLSearchParams rather than written out.
const pageURLForInitialState = (grain: string) =>
  new URLSearchParams([
    ["tr", DEFAULT_TIME_RANGE],
    ["grain", grain],
    ["measures", "impressions"],
    ["dims", "publisher"],
    ["sort_type", "percent"],
    ["sort_dir", "ASC"],
  ]).toString();
const PageURLForInitialState = pageURLForInitialState("day");
// Explore derives the grain of a range from its resolved start and end, which a component test does
// not resolve, so the dashboard loads on the smallest grain and the time filter then corrects it to
// the grain of the range.
const PageURLBeforeGrainIsResolved = pageURLForInitialState("minute");

/** The label of the range the dashboard loads with, which is what the picker shows on render. */
const DefaultTimeRangeLabel = /Last 7 days/;

/**
 * Renders the time filter bar the way the explore dashboard does, so that the time range reaches
 * the explore state and the url through `DashboardStateSync`.
 *
 * Registers the hooks the component tests need, so call it from a `describe`.
 */
export function useExploreTimeFiltersVariant(
  hoistedPage: HoistedPageForComponentTests,
): TimeFiltersVariant {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();
  let pageMock!: PageMockForComponentTests;

  beforeEach(() => {
    pageMock = new PageMockForComponentTests(hoistedPage);

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT_WITH_TIME);
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      {
        ...AD_BIDS_EXPLORE_INIT,
        timeRanges: YAML_TIME_RANGES,
        defaultPreset: {
          ...AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
          timeRange: DEFAULT_TIME_RANGE,
        },
      },
    );
    mocks.mockTimeRangeSummary(AD_BIDS_METRICS_NAME, TIME_RANGE_SUMMARY);
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, RESOLVED_RILL_TIMES);

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  afterAll(waitForBodyScrollCleanup);

  return {
    initialUrlSearch: PageURLForInitialState,
    // The preset lands in the url as soon as the dashboard loads.
    initialUrlSearchHistory: [
      PageURLBeforeGrainIsResolved,
      PageURLForInitialState,
    ],

    urlSearchWithTimeParams: (timeParams: Record<string, string>) => {
      // Overriding the params of the initial url keeps them in the order the dashboard writes them,
      // which is the order the url search assertions compare against.
      const urlSearch = new URLSearchParams(PageURLForInitialState);
      Object.entries(timeParams).forEach(([key, value]) =>
        urlSearch.set(key, value),
      );
      return urlSearch.toString();
    },

    pageMock: () => pageMock,

    timeFilterManager: {
      getTimeRange: () => testTimeRange(exploreState()?.selectedTimeRange),
      getTimeGrain: () => exploreState()?.selectedTimeRange?.interval,
      getComparisonTimeRange: () =>
        testTimeRange(exploreState()?.selectedComparisonTimeRange),
    },

    render: async (initUrlSearch?: string) => {
      // Make sure to populate the url with the initial states.
      pageMock.gotoSearch(PageURLForInitialState);
      if (initUrlSearch) pageMock.gotoSearch(initUrlSearch);
      render(ExploreExpressionFiltersTest, {
        props: {
          exploreName: AD_BIDS_EXPLORE_NAME,
        },
        // TODO: we need to make sure every single query uses an explicit queryClient instead of the
        //       global one. Only then we can use a fresh client here.
        context: new Map<string | symbol, unknown>([
          ["$$_queryClient", queryClient],
          [
            RUNTIME_CONTEXT_KEY,
            new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
          ],
        ]),
      });
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
      // Resolving the interval of the initial range is a network call, so the picker is only ready
      // for interaction once it shows the range the dashboard loaded with.
      await waitForTimeRangeLabel(DefaultTimeRangeLabel);
    },
  };
}

function exploreState() {
  return get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME];
}

function testTimeRange(timeRange: DashboardTimeControls | undefined) {
  if (!timeRange?.name) return undefined;
  return {
    name: timeRange.name,
    start: timeRange.start.toISOString(),
    end: timeRange.end.toISOString(),
  };
}
