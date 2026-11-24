import { getCanvasFilters } from "@rilldata/web-common/features/canvas/filters/canvas-filter-expressions";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import {
  getFilterChip,
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
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import CanvasReportFiltersTest from "@rilldata/web-common/features/scheduled-reports/test/CanvasReportFiltersTest.svelte";
import { applyReportFiltersToParams } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, waitFor } from "@testing-library/svelte";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";

// The SvelteKit mocks have to be declared in the spec file, since `vi.mock` is hoisted per file.
const hoistedPage: HoistedPageForComponentTests = vi.hoisted(() => ({}) as any);

vi.stubEnv("TZ", "UTC");

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: () => {},
    beforeNavigate: () => {},
  };
});
vi.mock("$app/state", async () => {
  return {
    page: (
      await import(
        "@rilldata/web-common/features/dashboards/state-managers/loaders/test/page-state.mock.svelte"
      )
    ).pageStateMock,
  };
});
vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
    navigating: {
      subscribe: (run: (value: null) => void) => {
        run(null);
        return () => {};
      },
    },
  };
});

const INSTANCE_ID = "test";
const AD_BIDS_CANVAS_NAME = "AdBids_canvas";
const AD_BIDS_CANVAS_INIT: V1CanvasSpec = {
  displayName: "AdBids canvas",
  filtersEnabled: true,
};
const FilterParamKey = `${ExploreStateURLParams.Filters}.${AD_BIDS_METRICS_NAME}`;

/**
 * Covers the canvas report form's filters: in-memory managers seeded with the report's canvas state.
 * The dialog saves the report's state and row restrictions from these managers,
 * so a manager left empty here silently drops the filters.
 */
describe("CanvasReportFilters", () => {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  mockResizeObserverForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();

  beforeEach(() => {
    new PageMockForComponentTests(hoistedPage);

    // The metrics view has a time dimension, so its time range summary is a second request that
    // only lands after the specs, which keeps the managers from being ready while they are seeded.
    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT_WITH_TIME);
    mocks.mockCanvas(AD_BIDS_CANVAS_NAME, AD_BIDS_CANVAS_INIT, {
      [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT_WITH_TIME,
    });
    mocks.mockTimeRangeSummary(
      AD_BIDS_METRICS_NAME,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary!,
    );

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
  });

  afterAll(waitForBodyScrollCleanup);

  const CapturedFilterParam = `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google')`;
  const CapturedState = new URLSearchParams({
    [FilterParamKey]: CapturedFilterParam,
    [ExploreStateURLParams.TimeRange]: "P7D",
    tabs: "deep_dive.detail",
  }).toString();
  const CapturedFilter = createAndExpression([
    createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
  ]);

  async function renderFilters() {
    let expressionFilterManager: ExpressionFilterManager | undefined;
    let timeFilterManager: TimeFilterManager | undefined;
    render(CanvasReportFiltersTest, {
      props: {
        canvasName: AD_BIDS_CANVAS_NAME,
        canvasState: CapturedState,
        onManagersCreated: (efm, tfm) => {
          expressionFilterManager = efm;
          timeFilterManager = tfm;
        },
      },
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

    await waitFor(() =>
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );
    await waitFor(() => expect(timeFilterManager?.ready).toBe(true));
    return {
      expressionFilterManager: expressionFilterManager!,
      timeFilterManager: timeFilterManager!,
    };
  }

  it("Should apply the captured filters once the metrics views are ready", async () => {
    const { expressionFilterManager } = await renderFilters();

    // This is what the report dialog saves as the report's `metricsViewFilters`.
    expect(getCanvasFilters(expressionFilterManager)).toEqual({
      [AD_BIDS_METRICS_NAME]: CapturedFilter,
    });
  });

  it("Should write edited filters over the captured state and keep the other params", async () => {
    const { expressionFilterManager, timeFilterManager } =
      await renderFilters();

    timeFilterManager.onSelectRange("P30D");
    const params = applyReportFiltersToParams(
      new URLSearchParams(CapturedState),
      expressionFilterManager,
      timeFilterManager,
    );

    expect(params.get(FilterParamKey)).toEqual(CapturedFilterParam);
    expect(params.get(ExploreStateURLParams.TimeRange)).toEqual("P30D");
    expect(params.get("tabs")).toEqual("deep_dive.detail");
  });
});
