import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import {
  type HoistedPageForExploreTests,
  PageMockForExploreTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
  AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SORT_ASC_BY_IMPRESSIONS,
  AD_BIDS_SORT_BY_PERCENT_VALUE,
  AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
  AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD,
  AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import DashboardStateManagerTest from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/DashboardStateManagerTest.svelte";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { render, screen, waitFor } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";

const hoistedPage: HoistedPageForExploreTests = vi.hoisted(() => ({}) as any);

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

type TestView = {
  view: string;
  additionalParams?: string;
  mutations: TestDashboardMutation[];
  expectedSearch: string;
};
const TestCases: {
  title: string;
  initView: TestView;
  view: TestView;
}[] = [
  {
    title: "Explore <=> tdd",
    initView: {
      view: "explore",
      mutations: [],
      expectedSearch:
        "tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&sort_type=percent",
    },
    view: {
      view: "tdd",
      additionalParams: "&measure=" + AD_BIDS_IMPRESSIONS_MEASURE,
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedSearch:
        "view=tdd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
  },
  {
    title: "dimension table <=> tdd",
    initView: {
      view: "explore",
      mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
      expectedSearch:
        "tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&expand_dim=publisher&sort_type=percent",
    },
    view: {
      view: "tdd",
      additionalParams: "&measure=" + AD_BIDS_IMPRESSIONS_MEASURE,
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedSearch:
        "view=tdd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
  },

  {
    title: "Explore <=> Pivot",
    initView: {
      view: "explore",
      mutations: [],
      expectedSearch:
        "tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&sort_type=percent",
    },
    view: {
      view: "pivot",
      mutations: [
        AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
        AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
      ],
      expectedSearch:
        "view=pivot&tr=P7D&compare_tr=rill-PP&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC&table_mode=nest",
    },
  },
  {
    title: "dimension table <=> Pivot",
    initView: {
      view: "explore",
      mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
      expectedSearch:
        "tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&expand_dim=publisher&sort_type=percent",
    },
    view: {
      view: "pivot",
      mutations: [
        AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
        AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
      ],
      expectedSearch:
        "view=pivot&tr=P7D&compare_tr=rill-PP&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC&table_mode=nest",
    },
  },
  {
    title: "tdd <=> Pivot",
    initView: {
      view: "tdd",
      additionalParams: "&measure=" + AD_BIDS_IMPRESSIONS_MEASURE,
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedSearch:
        "view=tdd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
    view: {
      view: "pivot",
      mutations: [
        AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
        AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
      ],
      expectedSearch:
        "view=pivot&tr=P7D&compare_tr=rill-PP&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC&table_mode=nest",
    },
  },
];

describe("Explore web view store", () => {
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

  for (const { title, initView, view } of TestCases) {
    it(title, async () => {
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      // apply mutations to main view to setup the initial state
      await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, [
        AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
        AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
        AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
        AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
        AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
        AD_BIDS_SORT_ASC_BY_IMPRESSIONS,
        AD_BIDS_SORT_BY_PERCENT_VALUE,
      ]);

      const initialSearch = `view=${initView.view}${initView.additionalParams ?? ""}`;
      // simulate going to the init view's url
      pageMock.gotoSearch(initialSearch);
      // apply any mutations in the init view
      await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, initView.mutations);
      const initState = getCleanMetricsExploreForAssertion();

      const viewSearch = `view=${view.view}${view.additionalParams ?? ""}`;
      // simulate going to the view's url
      pageMock.gotoSearch(viewSearch);
      // apply any mutations in the view
      await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, view.mutations);
      const stateInView = getCleanMetricsExploreForAssertion();

      // All history changes before this are a combination of visiting the view and mutations.
      const historyCutoff = pageMock.urlSearchHistory.length;

      // go back to init view without any additional params
      pageMock.gotoSearch(initialSearch);
      // new url should be filled with params from initView
      pageMock.assertSearchParams(initView.expectedSearch);
      // assert state is the same as initial view
      expect(getCleanMetricsExploreForAssertion()).toEqual(initState);
      // Revisiting the same view doesn't break anything.
      pageMock.gotoSearch(initialSearch);
      pageMock.assertSearchParams(initView.expectedSearch);
      expect(getCleanMetricsExploreForAssertion()).toEqual(initState);

      // go back to view without any additional params
      pageMock.gotoSearch(viewSearch);
      // new url should be filled with params from view
      pageMock.assertSearchParams(view.expectedSearch);
      // assert state is the same as we 1st entered view
      expect(getCleanMetricsExploreForAssertion()).toEqual(stateInView);
      // Revisiting the same view doesn't break anything.
      pageMock.gotoSearch(viewSearch);
      pageMock.assertSearchParams(view.expectedSearch);
      expect(getCleanMetricsExploreForAssertion()).toEqual(stateInView);

      // History after the all mutations are finished should only be of visiting the views.
      // This makes sure that replaceState in init is working as expected.
      expect(pageMock.urlSearchHistory.slice(historyCutoff)).toEqual([
        initView.expectedSearch,
        initView.expectedSearch,
        view.expectedSearch,
        view.expectedSearch,
      ]);
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
