import {
  getCanvasStore,
  removeCanvasStore,
} from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import CanvasExpressionFiltersTest from "@rilldata/web-common/features/dashboards/filters/test/CanvasExpressionFiltersTest.svelte";
import type { ExpressionFiltersVariant } from "@rilldata/web-common/features/dashboards/filters/test/expression-filters-suite";
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
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, screen, waitFor } from "@testing-library/svelte";
import { afterAll, beforeEach, expect } from "vitest";

const INSTANCE_ID = "test";

const AD_BIDS_CANVAS_NAME = "AdBids_canvas";
// A canvas with no rows, so the filter bar is all that renders. The metrics view has no time
// dimension either, so there are no time controls for the bar to wait on.
const AD_BIDS_CANVAS_INIT: V1CanvasSpec = {
  displayName: "AdBids canvas",
  filtersEnabled: true,
};

// A canvas filter bar writes one param per metrics view, the same as any standalone bar.
const FilterParamKey = `${ExploreStateURLParams.Filters}.${AD_BIDS_METRICS_NAME}`;

/**
 * Renders the filter bar the way a canvas dashboard does, so that the filter reaches the canvas
 * entity's filter manager and the url through `CanvasDashboardWrapper`.
 *
 * Registers the hooks the component tests need, so call it from a `describe`.
 */
export function useCanvasFiltersVariant(
  hoistedPage: HoistedPageForComponentTests,
): ExpressionFiltersVariant {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  mockResizeObserverForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();
  let pageMock!: PageMockForComponentTests;

  beforeEach(() => {
    pageMock = new PageMockForComponentTests(hoistedPage);

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
    mocks.mockCanvas(AD_BIDS_CANVAS_NAME, AD_BIDS_CANVAS_INIT, {
      [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT,
    });

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    // The canvas store registry and the last visited state both outlive a test, and a canvas that
    // has a last visited state redirects to it as it loads.
    removeCanvasStore(AD_BIDS_CANVAS_NAME, INSTANCE_ID);
    lastVisitedState.clear();
  });

  afterAll(waitForBodyScrollCleanup);

  // The canvas entity holds the filter manager, and the entity is rebuilt for every test.
  const canvasFilterManager = () =>
    getCanvasStore(AD_BIDS_CANVAS_NAME, INSTANCE_ID).canvasEntity
      .expressionFilterManager;

  return {
    initialUrlSearch: "",
    // Nothing is written to the url until a filter is applied.
    initialUrlSearchHistory: [],

    urlSearchWithFilter: (filter: string) =>
      new URLSearchParams(`${FilterParamKey}=${filter}`).toString(),

    pageMock: () => pageMock,

    expressionFilterManager: {
      getWhereFilter: () =>
        canvasFilterManager().topLevelJoiner.expr[AD_BIDS_METRICS_NAME] ??
        createAndExpression([]),
      getInListDimensions: () => canvasFilterManager().inList,
    },

    render: async (initUrlSearch?: string) => {
      if (initUrlSearch) pageMock.gotoSearch(initUrlSearch);
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
    },
  };
}
