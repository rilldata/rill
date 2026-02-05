import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import DashboardStateManagerTest from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/DashboardStateManagerTest.svelte";
import {
  type HoistedPageForExploreTests,
  PageMockForExploreTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
  AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
  AD_BIDS_SORT_BY_PERCENT_VALUE,
  AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_IMPRESSIONS_MEASURE_VISIBILITY,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { render, screen, waitFor } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";

const hoistedPage: HoistedPageForExploreTests = vi.hoisted(() => ({}) as any);

vi.stubEnv("TZ", "UTC");

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: () => {},
  };
});
vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
  };
});

const TestCases: {
  title: string;
  urlSearch: string;
  mutations: TestDashboardMutation[];

  expectedUrlSearch: string;
  expectedExplore: Partial<ExploreState>;
}[] = [
  {
    title: "Changes to dashboard using actions",
    urlSearch: "",
    mutations: [
      AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_TOGGLE_IMPRESSIONS_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
      AD_BIDS_SORT_BY_PERCENT_VALUE,
      AD_BIDS_SORT_ASC_BY_BID_PRICE,
    ],

    expectedUrlSearch:
      "measures=bid_price&dims=publisher&sort_by=bid_price&sort_type=percent&sort_dir=ASC&leaderboard_measures=bid_price",
    expectedExplore: {
      allMeasuresVisible: false,
      visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
      allDimensionsVisible: false,
      visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
      leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
    },
  },
  {
    title: "Changes to dashboard using url",
    urlSearch:
      "view=explore&tr=P4W&tz=UTC&compare_tr=&grain=week&compare_dim=&f=publisher in ('Google')&measures=bid_price&dims=publisher&expand_dim=&sort_by=bid_price&sort_type=percent&sort_dir=ASC&leaderboard_measures=bid_price",
    mutations: [],

    expectedUrlSearch:
      "measures=bid_price&dims=publisher&sort_by=bid_price&sort_type=percent&sort_dir=ASC&leaderboard_measures=bid_price",
    expectedExplore: {
      allMeasuresVisible: false,
      visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
      allDimensionsVisible: false,
      visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
      leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
    },
  },
];

describe("Most recent explore state", () => {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();
  let pageMock!: PageMockForExploreTests;

  beforeEach(async () => {
    pageMock = new PageMockForExploreTests(hoistedPage);

    mocks.mockMetricsView(
      AD_BIDS_METRICS_NAME,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
    );
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      AD_BIDS_EXPLORE_INIT,
    );
    mocks.mockTimeRangeSummary(
      AD_BIDS_METRICS_NAME,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary!,
    );

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  for (const {
    title,
    urlSearch,
    mutations,
    expectedUrlSearch,
    expectedExplore,
  } of TestCases) {
    it(title, async () => {
      const {
        renderResults: { unmount },
      } = renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
      const initState = getCleanMetricsExploreForAssertion();

      pageMock.gotoSearch(urlSearch);
      await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

      // clear the old dashboard to simulate closing the tab
      unmount();
      sessionStorage.clear();
      queryClient.clear();
      pageMock.reset();
      metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);

      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      pageMock.assertSearchParams(expectedUrlSearch);
      expect(getCleanMetricsExploreForAssertion()).toEqual({
        ...initState,
        ...expectedExplore,
      });
      // Make sure no extra url is added to the history
      expect(pageMock.urlSearchHistory).toEqual([expectedUrlSearch]);
    });
  }
});

// This needs to be there each file because of how hoisting works with vitest.
// TODO: find if there is a way to share code.
function renderDashboardStateManager() {
  const renderResults = render(DashboardStateManagerTest, {
    props: {
      exploreName: AD_BIDS_EXPLORE_NAME,
    },
    // TODO: we need to make sure every single query uses an explicit queryClient instead of the global one
    //       only then we can use a fresh client here.
    context: new Map([["$$_queryClient", queryClient]]),
  });

  return { queryClient, renderResults };
}
