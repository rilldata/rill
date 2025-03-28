import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import type { OtherSourceOfState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateLoader.svelte";
import DashboardStateLoaderTest from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateLoaderTest.svelte";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
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
  AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SET_CUSTOM_TIME_RANGE_FILTER,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
  AD_BIDS_SORT_BY_PERCENT_VALUE,
  AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_IMPRESSIONS_MEASURE_VISIBILITY,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import {
  type HoistedPage,
  PageMock,
} from "@rilldata/web-common/features/dashboards/url-state/PageMock";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { DashboardState_LeaderboardSortDirection } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { render, screen, waitFor } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";

const hoistedPage: HoistedPage = vi.hoisted(() => ({}) as any);

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
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
  expectedExplore: Partial<MetricsExplorerEntity>;
}[] = [
  {
    title: "Changes to dashboard using actions",
    urlSearch: "",
    mutations: [
      AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_TOGGLE_IMPRESSIONS_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
      AD_BIDS_SORT_BY_PERCENT_VALUE,
      AD_BIDS_SORT_ASC_BY_BID_PRICE,
    ],

    expectedUrlSearch:
      "tr=P7D&compare_tr=rill-PP&grain=day&measures=bid_price&dims=publisher&sort_by=bid_price&sort_dir=ASC",
    expectedExplore: {
      selectedTimeRange: {
        name: "P7D",
        interval: undefined,
      } as DashboardTimeControls,
      showTimeComparison: true,
      selectedComparisonTimeRange: { name: "rill-PP" } as DashboardTimeControls,
      allMeasuresVisible: false,
      visibleMeasureKeys: new Set([AD_BIDS_BID_PRICE_MEASURE]),
      allDimensionsVisible: false,
      visibleDimensionKeys: new Set([AD_BIDS_PUBLISHER_DIMENSION]),
      leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
    },
  },
  {
    title: "Changes to dashboard using url",
    urlSearch:
      "tr=P7D&compare_tr=rill-PW&grain=hour&measures=bid_price&dims=publisher&sort_by=bid_price&sort_dir=ASC&f=publisher in ('Facebook','Yahoo')",
    mutations: [],

    expectedUrlSearch:
      "tr=P7D&compare_tr=rill-PP&grain=day&measures=bid_price&dims=publisher&sort_by=bid_price&sort_dir=ASC",
    expectedExplore: {
      selectedTimeRange: {
        name: "P7D",
        interval: undefined,
      } as DashboardTimeControls,
      showTimeComparison: true,
      selectedComparisonTimeRange: { name: "rill-PP" } as DashboardTimeControls,
      allMeasuresVisible: false,
      visibleMeasureKeys: new Set([AD_BIDS_BID_PRICE_MEASURE]),
      allDimensionsVisible: false,
      visibleDimensionKeys: new Set([AD_BIDS_PUBLISHER_DIMENSION]),
      leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
    },
  },
  {
    title: "Custom time range selected.",
    urlSearch: "",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_SET_CUSTOM_TIME_RANGE_FILTER,
    ],

    // Custom time range is not retained
    expectedUrlSearch: "tr=P7D&compare_tr=rill-PP&grain=day",
    expectedExplore: {
      selectedTimeRange: {
        name: "P7D",
        interval: undefined,
      } as DashboardTimeControls,
      showTimeComparison: true,
      selectedComparisonTimeRange: { name: "rill-PP" } as DashboardTimeControls,
    },
  },
  {
    title: "Comparison is enabled and disabled.",
    urlSearch: "",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER,
    ],

    // Custom time range is not retained
    expectedUrlSearch: "tr=P7D&grain=day",
    expectedExplore: {
      selectedTimeRange: {
        name: "P7D",
        interval: undefined,
      } as DashboardTimeControls,
      showTimeComparison: false,
      selectedComparisonTimeRange: undefined,
    },
  },

  {
    title: "Changes to dashboard from TDD.",
    urlSearch: "view=tdd&measure=impressions",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
    ],

    // Custom time range is not retained
    expectedUrlSearch: "tr=P7D&grain=day",
    expectedExplore: {
      selectedTimeRange: {
        name: "P7D",
        interval: undefined,
      } as DashboardTimeControls,
      showTimeComparison: false,
      selectedComparisonTimeRange: undefined,
    },
  },
];

describe("Most recent explore state", () => {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();
  let pageMock!: PageMock;

  beforeEach(async () => {
    pageMock = new PageMock(hoistedPage);

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
      } = renderDashboardStateLoader();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
      const initState = getCleanMetricsExploreForAssertion();

      pageMock.gotoSearch(urlSearch);
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

      // clear the old dashboard to simulate closing the tab
      unmount();
      sessionStorage.clear();
      queryClient.clear();
      pageMock.reset();
      metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);

      renderDashboardStateLoader();
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
function renderDashboardStateLoader(
  otherStateSourceQueries: OtherSourceOfState["query"][] = [],
) {
  const renderResults = render(DashboardStateLoaderTest, {
    props: {
      exploreName: AD_BIDS_EXPLORE_NAME,
      otherSourcesOfState: otherStateSourceQueries.map((query) => ({
        errorHeader: "",
        query,
      })),
    },
    // TODO: we need to make sure every single query uses an explicit queryClient instead of the global one
    //       only then we can use a fresh client here.
    context: new Map([["$$_queryClient", queryClient]]),
  });

  return { queryClient, renderResults };
}
