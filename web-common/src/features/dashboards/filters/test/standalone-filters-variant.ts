import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { ExpressionFiltersVariant } from "@rilldata/web-common/features/dashboards/filters/test/expression-filters-suite";
import {
  mockPointerEventsForComponentTesting,
  useDashboardFetchMocksForComponentTests,
  waitForBodyScrollCleanup,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import StandaloneExpressionFiltersTest from "@rilldata/web-common/features/dashboards/filters/test/StandaloneExpressionFiltersTest.svelte";
import {
  type HoistedPageForExploreTests,
  PageMockForExploreTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, screen, waitFor } from "@testing-library/svelte";
import { afterAll, beforeEach, expect } from "vitest";

// A standalone filter bar writes one param per metrics view, with no preset to merge in.
const FilterParamKey = `${ExploreStateURLParams.Filters}.${AD_BIDS_METRICS_NAME}`;

/**
 * Renders the filter bar on its own, the way canvas does, so that the filter manager is the only
 * holder of the filter state and the url is written straight from it.
 *
 * Registers the hooks the component tests need, so call it from a `describe`.
 */
export function useStandaloneFiltersVariant(
  hoistedPage: HoistedPageForExploreTests,
): ExpressionFiltersVariant {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();
  let pageMock!: PageMockForExploreTests;
  let expressionFilterManager: ExpressionFilterManager | undefined;

  beforeEach(() => {
    pageMock = new PageMockForExploreTests(hoistedPage);
    expressionFilterManager = undefined;

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
  });

  afterAll(waitForBodyScrollCleanup);

  return {
    initialUrlSearch: "",
    // Nothing is written to the url until a filter is applied.
    initialUrlSearchHistory: [],

    urlSearchWithFilter: (filter: string) =>
      new URLSearchParams(`${FilterParamKey}=${filter}`).toString(),

    pageMock: () => pageMock,

    expressionFilterManager: {
      getWhereFilter: () =>
        expressionFilterManager!.topLevelJoiner.expr[AD_BIDS_METRICS_NAME] ??
        createAndExpression([]),
      getInListDimensions: () => expressionFilterManager!.inList,
    },

    render: async () => {
      render(StandaloneExpressionFiltersTest, {
        props: {
          metricsViewNames: [AD_BIDS_METRICS_NAME],
          onManagerCreated: (manager: ExpressionFilterManager) =>
            (expressionFilterManager = manager),
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
    },
  };
}
