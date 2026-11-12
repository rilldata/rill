import {
  getCanvasStore,
  removeCanvasStore,
} from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import CanvasExpressionFiltersTest from "@rilldata/web-common/features/dashboards/filters/test/CanvasExpressionFiltersTest.svelte";
import {
  mockPointerEventsForComponentTesting,
  mockResizeObserverForComponentTesting,
  useDashboardFetchMocksForComponentTests,
  waitForBodyScrollCleanup,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import {
  type HoistedPageForComponentTests,
  PageMockForComponentTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import {
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  DEFAULT_TIME_RANGE,
  RESOLVED_RILL_TIMES,
  TIME_RANGE_SUMMARY,
  YAML_TIME_RANGES,
  YAML_TIME_ZONES,
} from "@rilldata/web-common/features/dashboards/time-controls/test/rill-time-mocks";
import { waitForTimeRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/test/time-filter-test-utils";
import type { TimeFiltersVariant } from "@rilldata/web-common/features/dashboards/time-controls/test/time-filters-suite";
import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import type {
  V1CanvasSpec,
  V1ResolvedTimeRange,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, screen, waitFor } from "@testing-library/svelte";
import { afterAll, beforeEach, expect } from "vitest";

const INSTANCE_ID = "test";

const AD_BIDS_CANVAS_NAME = "AdBids_canvas";
// A canvas with no rows, so the filter bar is all that renders.
// Its metrics view has a time dimension, so the bar renders the time controls.
const AD_BIDS_CANVAS_INIT: V1CanvasSpec = {
  displayName: "AdBids canvas",
  filtersEnabled: true,
  timeRanges: YAML_TIME_RANGES,
  timeZones: YAML_TIME_ZONES,
  allowCustomTimeRange: true,
  defaultPreset: {
    timeRange: DEFAULT_TIME_RANGE,
  },
};

// Url params the canvas loads with, the same as the ones the explore variant loads with.
// A rilltime expression carries characters the url escapes, and the url search history holds the
// searches verbatim, so these are built through URLSearchParams rather than written out.
const PageURLForInitialState = new URLSearchParams([
  ["tr", DEFAULT_TIME_RANGE],
  ["grain", "day"],
]).toString();

/** The label of the range the dashboard loads with, which is what the picker shows on render. */
const DefaultTimeRangeLabel = /Last 7 days/;

/**
 * Renders the time filter bar the way a canvas dashboard does, so that the time range reaches the
 * canvas entity's time filter manager and the url through `CanvasDashboardWrapper`.
 *
 * Registers the hooks the component tests need, so call it from a `describe`.
 */
export function useCanvasTimeFiltersVariant(
  hoistedPage: HoistedPageForComponentTests,
): TimeFiltersVariant {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  mockResizeObserverForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();
  let pageMock!: PageMockForComponentTests;

  function mockMetricsViewAndCanvas(timeGrains?: V1TimeGrain[]) {
    const metricsView = timeGrains
      ? { ...AD_BIDS_METRICS_INIT_WITH_TIME, timeGrains }
      : AD_BIDS_METRICS_INIT_WITH_TIME;
    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, metricsView);
    mocks.mockCanvas(AD_BIDS_CANVAS_NAME, AD_BIDS_CANVAS_INIT, {
      [AD_BIDS_METRICS_NAME]: metricsView,
    });
  }

  beforeEach(() => {
    pageMock = new PageMockForComponentTests(hoistedPage);

    mockMetricsViewAndCanvas();
    mocks.mockTimeRangeSummary(AD_BIDS_METRICS_NAME, TIME_RANGE_SUMMARY);
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, RESOLVED_RILL_TIMES);

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    // The canvas store registry and the last visited state both outlive a test, and a canvas that
    // has a last visited state redirects to it as it loads.
    removeCanvasStore(AD_BIDS_CANVAS_NAME, INSTANCE_ID);
    lastVisitedState.clear();
  });

  afterAll(waitForBodyScrollCleanup);

  // The canvas entity holds the time filter manager, and the entity is rebuilt for every test.
  const canvasTimeFilterManager = () =>
    getCanvasStore(AD_BIDS_CANVAS_NAME, INSTANCE_ID).canvasEntity
      .timeFilterManager;

  return {
    initialUrlSearch: PageURLForInitialState,
    initialUrlSearchHistory: [PageURLForInitialState],
    // The canvas writes the url from its time filter manager, which keeps the grain as selected.
    urlKeepsDisallowedGrain: true,

    urlSearchWithTimeParams: (
      timeParams: Record<string, string | undefined>,
    ) => {
      const urlSearch = new URLSearchParams(PageURLForInitialState);
      Object.entries(timeParams).forEach(([key, value]) => {
        if (value === undefined) urlSearch.delete(key);
        else urlSearch.set(key, value);
      });
      return urlSearch.toString();
    },

    pageMock: () => pageMock,

    timeFilterManager: {
      getTimeRange: () => testTimeRange(canvasTimeFilterManager()),
      getTimeGrain: () => canvasTimeFilterManager().timeGrain,
      getComparisonTimeRange: () =>
        testComparisonTimeRange(canvasTimeFilterManager()),
      getComparisonEnabled: () => canvasTimeFilterManager().showComparison,
    },

    render: async ({ initUrlSearch, timeGrains } = {}) => {
      if (timeGrains) mockMetricsViewAndCanvas(timeGrains);
      // A canvas with url params loads on them rather than redirecting to its default preset.
      pageMock.gotoSearch(initUrlSearch ?? PageURLForInitialState);
      render(CanvasExpressionFiltersTest, {
        props: {
          canvasName: AD_BIDS_CANVAS_NAME,
        },
        // TODO: we need to make sure every single query uses an explicit queryClient instead of the
        //       global one. Only then we can use a fresh client here.
        context: new Map<string | symbol, unknown>([
          ["$$_queryClient", queryClient],
          [
            RUNTIME_CONTEXT_KEY,
            new RuntimeClient({
              host: "http://localhost",
              instanceId: INSTANCE_ID,
            }),
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

function testTimeRange(timeFilterManager: TimeFilterManager) {
  if (!timeFilterManager.timeRange) return undefined;
  return <V1ResolvedTimeRange>{
    expression: timeFilterManager.timeRange,
    start: timeFilterManager.timeStart,
    end: timeFilterManager.timeEnd,
    grain: timeFilterManager.timeGrain,
  };
}

function testComparisonTimeRange(timeFilterManager: TimeFilterManager) {
  if (!timeFilterManager.comparisonTimeRange) return undefined;
  return <V1ResolvedTimeRange>{
    expression: timeFilterManager.comparisonTimeRange,
    start: timeFilterManager.comparisonTimeStart,
    end: timeFilterManager.comparisonTimeEnd,
  };
}
