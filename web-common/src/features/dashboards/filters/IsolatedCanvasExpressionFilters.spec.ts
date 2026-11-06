import {
  getCanvasStore,
  removeCanvasStore,
} from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import IsolatedCanvasFiltersTest from "@rilldata/web-common/features/dashboards/filters/test/IsolatedCanvasFiltersTest.svelte";
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
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
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
// The rune based url stores read `page` from here, and SvelteKit only populates it through its
// client router, which a component test does not boot.
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
// A canvas with no rows, so the filter bar is all that renders.
const AD_BIDS_CANVAS_INIT: V1CanvasSpec = {
  displayName: "AdBids canvas",
  filtersEnabled: true,
};
const FilterParamKey = `${ExploreStateURLParams.Filters}.${AD_BIDS_METRICS_NAME}`;

/**
 * Covers the report form's filter bar: an isolated `CanvasProvider` fed the report's captured
 * state, with no `CanvasDashboardWrapper` and so no `syncStoreWithSource` to replay the params
 * `CanvasEntity.onUrlChange` holds back until the metrics views are ready. The filters the
 * dialog saves as the report's row restrictions are read off this manager, so a manager left
 * empty here silently drops them.
 */
describe("IsolatedCanvasExpressionFilters", () => {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  mockResizeObserverForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();

  beforeEach(() => {
    // Nothing here asserts on the url, but the canvas still reads it as it loads.
    new PageMockForComponentTests(hoistedPage);

    // The metrics view has a time dimension, so its time range summary is a second request that
    // only lands after the specs. That gap is what leaves `metricsViewsProvider.ready` false
    // while the canvas state is first applied.
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
    removeCanvasStore(AD_BIDS_CANVAS_NAME, INSTANCE_ID);
    lastVisitedState.clear();
  });

  afterAll(waitForBodyScrollCleanup);

  const CapturedState = `${FilterParamKey}=${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google')`;
  const CapturedFilter = createAndExpression([
    createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
  ]);

  function renderFilterBar() {
    return render(IsolatedCanvasFiltersTest, {
      props: {
        canvasName: AD_BIDS_CANVAS_NAME,
        canvasStateOverride: CapturedState,
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
  }

  async function assertCapturedFilterApplied() {
    await waitFor(() =>
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );

    // This is what the report dialog saves as the report's `metricsViewFilters`.
    expect(
      getCanvasStore(AD_BIDS_CANVAS_NAME, INSTANCE_ID).canvasEntity
        .expressionFilterManager.topLevelJoiner.expr[AD_BIDS_METRICS_NAME],
    ).toEqual(CapturedFilter);
  }

  it("Should apply the captured filters once the metrics views are ready", async () => {
    renderFilterBar();

    await assertCapturedFilterApplied();
  });

  it("Should apply the captured filters when only the time range summary is pending", async () => {
    // Opening the dialog from a page that already loaded the metrics view resources leaves the
    // specs cached and the time range summary as the only request still in flight, which is
    // enough to keep `metricsViewsProvider.ready` false while the state is first applied.
    const { unmount } = renderFilterBar();
    await assertCapturedFilterApplied();
    unmount();
    removeCanvasStore(AD_BIDS_CANVAS_NAME, INSTANCE_ID);

    renderFilterBar();

    await assertCapturedFilterApplied();
  });
});
