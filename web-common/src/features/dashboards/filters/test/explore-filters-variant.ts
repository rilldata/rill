import type { ExpressionFiltersVariant } from "@rilldata/web-common/features/dashboards/filters/test/expression-filters-suite";
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
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/test/url-state-test-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, screen, waitFor } from "@testing-library/svelte";
import { afterAll, beforeEach, expect } from "vitest";

// Url params for the dashboard before any filter is applied, coming from the yaml preset.
const PageURLForInitialState =
  "measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC";

/**
 * Renders the filter bar the way the explore dashboard does, so that the filter reaches the explore
 * state and the url through `DashboardStateSync`.
 *
 * Registers the hooks the component tests need, so call it from a `describe`.
 */
export function useExploreFiltersVariant(
  hoistedPage: HoistedPageForComponentTests,
): ExpressionFiltersVariant {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();
  let pageMock!: PageMockForComponentTests;

  beforeEach(() => {
    pageMock = new PageMockForComponentTests(hoistedPage);

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
    mocks.mockMetricsExplore(AD_BIDS_EXPLORE_NAME, AD_BIDS_METRICS_INIT, {
      ...AD_BIDS_EXPLORE_INIT,
      defaultPreset: AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
    });

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  afterAll(waitForBodyScrollCleanup);

  return {
    initialUrlSearch: PageURLForInitialState,
    // The preset lands in the url as soon as the dashboard loads.
    initialUrlSearchHistory: [PageURLForInitialState],

    urlSearchWithFilter: (filter: string) =>
      new URLSearchParams(`f=${filter}&${PageURLForInitialState}`).toString(),

    pageMock: () => pageMock,

    expressionFilterManager: {
      getWhereFilter: () => getCleanMetricsExploreForAssertion().whereFilter,
      getInListDimensions: () =>
        getCleanMetricsExploreForAssertion().dimensionsWithInlistFilter ?? [],
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
    },
  };
}
